// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";

/**
 * @title LeetVault
 * @dev Main vault contract for LeetGaming platform
 * Manages prize pools, entry fees, and prize distribution
 * Blockchain serves as source of truth for all financial state
 */
contract LeetVault is AccessControl, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    // Roles
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE");

    // Supported stablecoins
    mapping(address => bool) public supportedTokens;

    // Prize pool states
    enum PrizePoolStatus {
        NotCreated,     // 0
        Accumulating,   // 1 - Accepting entry fees
        Locked,         // 2 - Match in progress
        InEscrow,       // 3 - Match complete, escrow period
        Distributed,    // 4 - Prizes sent
        Cancelled       // 5 - Refunded
    }

    struct PrizePool {
        bytes32 matchId;
        address token;
        uint256 totalAmount;
        uint256 platformContribution;
        uint256 entryFeePerPlayer;
        uint256 platformFeePercent;  // basis points (100 = 1%)
        uint256 createdAt;
        uint256 lockedAt;
        uint256 escrowEndTime;
        PrizePoolStatus status;
        address[] participants;
        mapping(address => uint256) contributions;
        mapping(address => bool) hasWithdrawn;
    }

    // Match ID => Prize Pool
    mapping(bytes32 => PrizePool) public prizePools;

    // User balances (for withdrawals)
    mapping(address => mapping(address => uint256)) public userBalances;

    // Platform treasury
    address public treasury;

    // Escrow period (default 72 hours)
    uint256 public escrowPeriod = 72 hours;

    // Platform contribution per match
    uint256 public platformContributionPerMatch = 50 * 1e4; // $0.50 in 6 decimals

    // Events
    event TokenAdded(address indexed token);
    event TokenRemoved(address indexed token);
    event PrizePoolCreated(bytes32 indexed matchId, address token, uint256 entryFee, uint256 platformFee);
    event EntryFeeDeposited(bytes32 indexed matchId, address indexed player, uint256 amount);
    event PrizePoolLocked(bytes32 indexed matchId, uint256 totalAmount);
    event PrizePoolInEscrow(bytes32 indexed matchId, uint256 escrowEndTime);
    event PrizeDistributed(bytes32 indexed matchId, address indexed winner, uint256 amount, uint8 rank);
    event PrizePoolCancelled(bytes32 indexed matchId);
    event RefundIssued(bytes32 indexed matchId, address indexed player, uint256 amount);
    event PlatformFeeCollected(bytes32 indexed matchId, uint256 amount);
    event UserWithdrawal(address indexed user, address indexed token, uint256 amount);
    event EscrowPeriodUpdated(uint256 oldPeriod, uint256 newPeriod);

    constructor(address _treasury) {
        require(_treasury != address(0), "Invalid treasury");
        treasury = _treasury;

        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(OPERATOR_ROLE, msg.sender);
        _grantRole(ORACLE_ROLE, msg.sender);
    }

    // ============ Admin Functions ============

    function addSupportedToken(address token) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(token != address(0), "Invalid token");
        supportedTokens[token] = true;
        emit TokenAdded(token);
    }

    function removeSupportedToken(address token) external onlyRole(DEFAULT_ADMIN_ROLE) {
        supportedTokens[token] = false;
        emit TokenRemoved(token);
    }

    function setEscrowPeriod(uint256 newPeriod) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(newPeriod >= 1 hours && newPeriod <= 7 days, "Invalid period");
        uint256 oldPeriod = escrowPeriod;
        escrowPeriod = newPeriod;
        emit EscrowPeriodUpdated(oldPeriod, newPeriod);
    }

    function setTreasury(address newTreasury) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(newTreasury != address(0), "Invalid treasury");
        treasury = newTreasury;
    }

    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    // ============ Prize Pool Management ============

    /**
     * @dev Create a new prize pool for a match
     * @param matchId Unique match identifier
     * @param token Stablecoin address (USDC/USDT)
     * @param entryFee Entry fee per player (in token decimals)
     * @param platformFeePercent Platform fee in basis points (100 = 1%)
     */
    function createPrizePool(
        bytes32 matchId,
        address token,
        uint256 entryFee,
        uint256 platformFeePercent
    ) external onlyRole(OPERATOR_ROLE) whenNotPaused {
        require(supportedTokens[token], "Token not supported");
        require(prizePools[matchId].status == PrizePoolStatus.NotCreated, "Pool exists");
        require(platformFeePercent <= 2000, "Fee too high"); // Max 20%

        PrizePool storage pool = prizePools[matchId];
        pool.matchId = matchId;
        pool.token = token;
        pool.entryFeePerPlayer = entryFee;
        pool.platformFeePercent = platformFeePercent;
        pool.platformContribution = platformContributionPerMatch;
        pool.createdAt = block.timestamp;
        pool.status = PrizePoolStatus.Accumulating;

        emit PrizePoolCreated(matchId, token, entryFee, platformFeePercent);
    }

    /**
     * @dev Player deposits entry fee to join match
     * @param matchId Match to join
     */
    function depositEntryFee(bytes32 matchId) external nonReentrant whenNotPaused {
        PrizePool storage pool = prizePools[matchId];
        require(pool.status == PrizePoolStatus.Accumulating, "Pool not accepting");
        require(pool.contributions[msg.sender] == 0, "Already joined");

        IERC20 token = IERC20(pool.token);
        uint256 fee = pool.entryFeePerPlayer;

        token.safeTransferFrom(msg.sender, address(this), fee);

        pool.contributions[msg.sender] = fee;
        pool.participants.push(msg.sender);
        pool.totalAmount += fee;

        emit EntryFeeDeposited(matchId, msg.sender, fee);
    }

    /**
     * @dev Lock prize pool when match starts (no more entries)
     * @param matchId Match identifier
     */
    function lockPrizePool(bytes32 matchId) external onlyRole(OPERATOR_ROLE) {
        PrizePool storage pool = prizePools[matchId];
        require(pool.status == PrizePoolStatus.Accumulating, "Invalid status");
        require(pool.participants.length >= 2, "Not enough players");

        // Add platform contribution
        pool.totalAmount += pool.platformContribution;
        pool.lockedAt = block.timestamp;
        pool.status = PrizePoolStatus.Locked;

        emit PrizePoolLocked(matchId, pool.totalAmount);
    }

    /**
     * @dev Move to escrow after match completes (Oracle reports result)
     * @param matchId Match identifier
     */
    function startEscrow(bytes32 matchId) external onlyRole(ORACLE_ROLE) {
        PrizePool storage pool = prizePools[matchId];
        require(pool.status == PrizePoolStatus.Locked, "Invalid status");

        pool.escrowEndTime = block.timestamp + escrowPeriod;
        pool.status = PrizePoolStatus.InEscrow;

        emit PrizePoolInEscrow(matchId, pool.escrowEndTime);
    }

    /**
     * @dev Distribute prizes to winners after escrow period
     * @param matchId Match identifier
     * @param winners Array of winner addresses (ordered by rank)
     * @param shares Array of share percentages in basis points (must sum to 10000)
     */
    function distributePrizes(
        bytes32 matchId,
        address[] calldata winners,
        uint256[] calldata shares
    ) external onlyRole(OPERATOR_ROLE) nonReentrant {
        PrizePool storage pool = prizePools[matchId];
        require(pool.status == PrizePoolStatus.InEscrow, "Not in escrow");
        require(block.timestamp >= pool.escrowEndTime, "Escrow not complete");
        require(winners.length == shares.length, "Length mismatch");
        require(winners.length > 0, "No winners");

        // Verify shares sum to 10000 (100%)
        uint256 totalShares;
        for (uint256 i = 0; i < shares.length; i++) {
            totalShares += shares[i];
        }
        require(totalShares == 10000, "Shares must sum to 100%");

        // Calculate platform fee
        uint256 platformFee = (pool.totalAmount * pool.platformFeePercent) / 10000;
        uint256 distributableAmount = pool.totalAmount - platformFee;

        // Distribute to winners
        IERC20 token = IERC20(pool.token);
        for (uint256 i = 0; i < winners.length; i++) {
            uint256 prizeAmount = (distributableAmount * shares[i]) / 10000;
            if (prizeAmount > 0) {
                userBalances[winners[i]][pool.token] += prizeAmount;
                emit PrizeDistributed(matchId, winners[i], prizeAmount, uint8(i + 1));
            }
        }

        // Send platform fee to treasury
        if (platformFee > 0) {
            token.safeTransfer(treasury, platformFee);
            emit PlatformFeeCollected(matchId, platformFee);
        }

        pool.status = PrizePoolStatus.Distributed;
    }

    /**
     * @dev Cancel prize pool and refund all participants
     * @param matchId Match identifier
     */
    function cancelPrizePool(bytes32 matchId) external onlyRole(OPERATOR_ROLE) nonReentrant {
        PrizePool storage pool = prizePools[matchId];
        require(
            pool.status == PrizePoolStatus.Accumulating ||
            pool.status == PrizePoolStatus.Locked,
            "Cannot cancel"
        );

        // Refund all participants
        for (uint256 i = 0; i < pool.participants.length; i++) {
            address participant = pool.participants[i];
            uint256 refundAmount = pool.contributions[participant];
            if (refundAmount > 0 && !pool.hasWithdrawn[participant]) {
                pool.hasWithdrawn[participant] = true;
                userBalances[participant][pool.token] += refundAmount;
                emit RefundIssued(matchId, participant, refundAmount);
            }
        }

        pool.status = PrizePoolStatus.Cancelled;
        emit PrizePoolCancelled(matchId);
    }

    // ============ User Functions ============

    /**
     * @dev Withdraw user balance to external wallet
     * @param token Token to withdraw
     * @param amount Amount to withdraw
     */
    function withdraw(address token, uint256 amount) external nonReentrant whenNotPaused {
        require(amount > 0, "Zero amount");
        require(userBalances[msg.sender][token] >= amount, "Insufficient balance");

        userBalances[msg.sender][token] -= amount;
        IERC20(token).safeTransfer(msg.sender, amount);

        emit UserWithdrawal(msg.sender, token, amount);
    }

    /**
     * @dev Deposit tokens to user balance (for platform use)
     * @param token Token to deposit
     * @param amount Amount to deposit
     */
    function deposit(address token, uint256 amount) external nonReentrant whenNotPaused {
        require(supportedTokens[token], "Token not supported");
        require(amount > 0, "Zero amount");

        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);
        userBalances[msg.sender][token] += amount;
    }

    // ============ View Functions ============

    function getPrizePoolInfo(bytes32 matchId) external view returns (
        address token,
        uint256 totalAmount,
        uint256 entryFee,
        uint256 platformFee,
        uint256 participantCount,
        PrizePoolStatus status,
        uint256 escrowEndTime
    ) {
        PrizePool storage pool = prizePools[matchId];
        return (
            pool.token,
            pool.totalAmount,
            pool.entryFeePerPlayer,
            pool.platformFeePercent,
            pool.participants.length,
            pool.status,
            pool.escrowEndTime
        );
    }

    function getParticipants(bytes32 matchId) external view returns (address[] memory) {
        return prizePools[matchId].participants;
    }

    function getContribution(bytes32 matchId, address participant) external view returns (uint256) {
        return prizePools[matchId].contributions[participant];
    }

    function getUserBalance(address user, address token) external view returns (uint256) {
        return userBalances[user][token];
    }
}
