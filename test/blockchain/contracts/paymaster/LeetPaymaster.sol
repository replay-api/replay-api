// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title LeetPaymaster
 * @dev ERC-4337 Paymaster for gas sponsorship
 * Supports:
 * - Free gas for verified users
 * - Gas payment in ERC-20 tokens (USDC/USDT)
 * - Sponsored transactions for specific operations
 * - Rate limiting and abuse prevention
 */
contract LeetPaymaster is AccessControl {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;
    using SafeERC20 for IERC20;

    // ============ Constants ============

    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");
    bytes32 public constant SIGNER_ROLE = keccak256("SIGNER_ROLE");

    uint256 private constant POST_OP_GAS = 50000;
    uint256 private constant PRICE_DENOMINATOR = 1e6;

    // ============ State Variables ============

    address public entryPoint;
    address public verifyingSigner;

    // Token payment settings
    mapping(address => bool) public acceptedTokens;
    mapping(address => uint256) public tokenPrices; // Price in token per gas unit (scaled by 1e6)

    // Sponsorship settings
    struct SponsorshipConfig {
        bool isActive;
        uint256 dailyLimit;      // Max gas units per day
        uint256 perTxLimit;      // Max gas units per transaction
        uint256 dailyUsed;
        uint256 lastResetDay;
        address[] allowedTargets; // Specific contracts that can be called
    }

    mapping(address => SponsorshipConfig) public sponsorships; // wallet => config

    // User gas credits (for promotional credits)
    mapping(address => uint256) public gasCredits;

    // Rate limiting
    mapping(address => uint256) public lastTxTimestamp;
    uint256 public minTxInterval = 1; // seconds between transactions

    // Platform wallet for token payments
    address public treasury;

    // Total gas sponsored stats
    uint256 public totalGasSponsored;
    uint256 public totalTokensCollected;

    // ============ Events ============

    event GasSponsored(address indexed wallet, uint256 gasUsed, bytes32 userOpHash);
    event TokenPayment(address indexed wallet, address indexed token, uint256 amount, uint256 gasUsed);
    event SponsorshipConfigured(address indexed wallet, uint256 dailyLimit, uint256 perTxLimit);
    event GasCreditsAdded(address indexed wallet, uint256 amount);
    event GasCreditsUsed(address indexed wallet, uint256 amount);
    event TokenAccepted(address indexed token, uint256 price);
    event TokenRemoved(address indexed token);

    // ============ Constructor ============

    constructor(address _entryPoint, address _treasury, address _verifyingSigner) {
        require(_entryPoint != address(0), "Invalid entry point");
        require(_treasury != address(0), "Invalid treasury");
        require(_verifyingSigner != address(0), "Invalid signer");

        entryPoint = _entryPoint;
        treasury = _treasury;
        verifyingSigner = _verifyingSigner;

        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(OPERATOR_ROLE, msg.sender);
        _grantRole(SIGNER_ROLE, _verifyingSigner);
    }

    // ============ ERC-4337 Paymaster Interface ============

    /**
     * @dev Validate paymaster user operation
     * @param userOp The user operation
     * @param userOpHash Hash of the user operation
     * @param maxCost Maximum cost the paymaster might pay
     * @return context Context to pass to postOp
     * @return validationData Signature validation result
     */
    function validatePaymasterUserOp(
        PackedUserOperation calldata userOp,
        bytes32 userOpHash,
        uint256 maxCost
    ) external returns (bytes memory context, uint256 validationData) {
        require(msg.sender == entryPoint, "Only EntryPoint");

        // Decode paymaster data
        // Format: [20 bytes token address][32 bytes signature][variable length additional data]
        bytes calldata paymasterData = userOp.paymasterAndData[20:];

        PaymentMode mode = _decodePaymentMode(paymasterData);
        address wallet = userOp.sender;

        if (mode == PaymentMode.Sponsored) {
            // Check sponsorship eligibility
            require(_validateSponsorship(wallet, maxCost), "Sponsorship invalid");
            context = abi.encode(wallet, PaymentMode.Sponsored, maxCost);
        }
        else if (mode == PaymentMode.GasCredits) {
            // Use pre-purchased gas credits
            require(gasCredits[wallet] >= maxCost, "Insufficient credits");
            context = abi.encode(wallet, PaymentMode.GasCredits, maxCost);
        }
        else if (mode == PaymentMode.TokenPayment) {
            // Pay gas in ERC-20 tokens
            (address token, uint256 tokenAmount) = _decodeTokenPayment(paymasterData);
            require(acceptedTokens[token], "Token not accepted");

            // Transfer tokens from wallet to paymaster
            // Note: Wallet must have approved the paymaster
            IERC20(token).safeTransferFrom(wallet, address(this), tokenAmount);

            context = abi.encode(wallet, PaymentMode.TokenPayment, maxCost, token, tokenAmount);
        }
        else if (mode == PaymentMode.VerifiedFree) {
            // Free gas for verified users (signature required)
            bytes memory signature = _extractSignature(paymasterData);
            require(_validateVerifyingSignature(userOp, userOpHash, signature), "Invalid signature");
            context = abi.encode(wallet, PaymentMode.VerifiedFree, maxCost);
        }
        else {
            revert("Invalid payment mode");
        }

        // Rate limiting
        require(block.timestamp >= lastTxTimestamp[wallet] + minTxInterval, "Too frequent");
        lastTxTimestamp[wallet] = block.timestamp;

        return (context, 0); // 0 = valid
    }

    /**
     * @dev Post-operation handling
     * @param mode Post-op mode (0 = success, 1 = user reverted, 2 = paymaster reverted)
     * @param context Context from validatePaymasterUserOp
     * @param actualGasCost Actual gas cost paid
     */
    function postOp(
        PostOpMode mode,
        bytes calldata context,
        uint256 actualGasCost,
        uint256 /*actualUserOpFeePerGas*/
    ) external {
        require(msg.sender == entryPoint, "Only EntryPoint");

        (address wallet, PaymentMode paymentMode, uint256 maxCost) =
            abi.decode(context[:96], (address, PaymentMode, uint256));

        if (paymentMode == PaymentMode.Sponsored) {
            // Update sponsorship usage
            SponsorshipConfig storage config = sponsorships[wallet];
            config.dailyUsed += actualGasCost;
            totalGasSponsored += actualGasCost;

            emit GasSponsored(wallet, actualGasCost, bytes32(0));
        }
        else if (paymentMode == PaymentMode.GasCredits) {
            // Deduct from credits
            gasCredits[wallet] -= actualGasCost;
            emit GasCreditsUsed(wallet, actualGasCost);
        }
        else if (paymentMode == PaymentMode.TokenPayment) {
            // Handle token payment reconciliation
            (, , , address token, uint256 tokenAmount) =
                abi.decode(context, (address, PaymentMode, uint256, address, uint256));

            // Calculate actual token cost based on gas used
            uint256 actualTokenCost = _calculateTokenCost(token, actualGasCost);

            // Refund excess tokens
            if (tokenAmount > actualTokenCost) {
                IERC20(token).safeTransfer(wallet, tokenAmount - actualTokenCost);
            }

            // Send collected tokens to treasury
            IERC20(token).safeTransfer(treasury, actualTokenCost);
            totalTokensCollected += actualTokenCost;

            emit TokenPayment(wallet, token, actualTokenCost, actualGasCost);
        }
        // VerifiedFree: nothing to do, gas is sponsored
    }

    // ============ Payment Mode Helpers ============

    enum PaymentMode {
        Sponsored,      // Platform-sponsored (whitelisted wallets)
        GasCredits,     // Pre-purchased credits
        TokenPayment,   // Pay in USDC/USDT
        VerifiedFree    // Free for verified users (signed by platform)
    }

    enum PostOpMode {
        OpSucceeded,
        OpReverted,
        PostOpReverted
    }

    function _decodePaymentMode(bytes calldata data) internal pure returns (PaymentMode) {
        if (data.length < 1) return PaymentMode.Sponsored;
        return PaymentMode(uint8(data[0]));
    }

    function _decodeTokenPayment(bytes calldata data) internal pure returns (address token, uint256 amount) {
        // Skip mode byte
        token = address(bytes20(data[1:21]));
        amount = uint256(bytes32(data[21:53]));
    }

    function _extractSignature(bytes calldata data) internal pure returns (bytes memory) {
        // Skip mode byte, extract 65-byte signature
        return data[1:66];
    }

    function _validateSponsorship(address wallet, uint256 maxCost) internal returns (bool) {
        SponsorshipConfig storage config = sponsorships[wallet];

        if (!config.isActive) return false;

        // Reset daily counter if new day
        uint256 currentDay = block.timestamp / 1 days;
        if (currentDay > config.lastResetDay) {
            config.dailyUsed = 0;
            config.lastResetDay = currentDay;
        }

        // Check limits
        if (maxCost > config.perTxLimit) return false;
        if (config.dailyUsed + maxCost > config.dailyLimit) return false;

        return true;
    }

    function _validateVerifyingSignature(
        PackedUserOperation calldata userOp,
        bytes32 userOpHash,
        bytes memory signature
    ) internal view returns (bool) {
        bytes32 hash = keccak256(abi.encode(
            userOpHash,
            address(this),
            block.chainid
        )).toEthSignedMessageHash();

        address signer = hash.recover(signature);
        return hasRole(SIGNER_ROLE, signer);
    }

    function _calculateTokenCost(address token, uint256 gasUsed) internal view returns (uint256) {
        uint256 price = tokenPrices[token];
        return (gasUsed * price) / PRICE_DENOMINATOR;
    }

    // ============ Admin Functions ============

    function configureSponsorship(
        address wallet,
        uint256 dailyLimit,
        uint256 perTxLimit,
        address[] calldata allowedTargets
    ) external onlyRole(OPERATOR_ROLE) {
        sponsorships[wallet] = SponsorshipConfig({
            isActive: true,
            dailyLimit: dailyLimit,
            perTxLimit: perTxLimit,
            dailyUsed: 0,
            lastResetDay: block.timestamp / 1 days,
            allowedTargets: allowedTargets
        });

        emit SponsorshipConfigured(wallet, dailyLimit, perTxLimit);
    }

    function disableSponsorship(address wallet) external onlyRole(OPERATOR_ROLE) {
        sponsorships[wallet].isActive = false;
    }

    function addGasCredits(address wallet, uint256 amount) external onlyRole(OPERATOR_ROLE) {
        gasCredits[wallet] += amount;
        emit GasCreditsAdded(wallet, amount);
    }

    function setAcceptedToken(address token, uint256 pricePerGas) external onlyRole(DEFAULT_ADMIN_ROLE) {
        acceptedTokens[token] = true;
        tokenPrices[token] = pricePerGas;
        emit TokenAccepted(token, pricePerGas);
    }

    function removeAcceptedToken(address token) external onlyRole(DEFAULT_ADMIN_ROLE) {
        acceptedTokens[token] = false;
        tokenPrices[token] = 0;
        emit TokenRemoved(token);
    }

    function setTreasury(address newTreasury) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(newTreasury != address(0), "Invalid treasury");
        treasury = newTreasury;
    }

    function setVerifyingSigner(address newSigner) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(newSigner != address(0), "Invalid signer");
        _revokeRole(SIGNER_ROLE, verifyingSigner);
        _grantRole(SIGNER_ROLE, newSigner);
        verifyingSigner = newSigner;
    }

    function setMinTxInterval(uint256 interval) external onlyRole(DEFAULT_ADMIN_ROLE) {
        minTxInterval = interval;
    }

    // ============ Fund Management ============

    function deposit() external payable {
        // Allow deposits to fund gas sponsorship
    }

    function withdrawETH(uint256 amount) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(address(this).balance >= amount, "Insufficient balance");
        (bool success,) = treasury.call{value: amount}("");
        require(success, "Transfer failed");
    }

    function withdrawToken(address token, uint256 amount) external onlyRole(DEFAULT_ADMIN_ROLE) {
        IERC20(token).safeTransfer(treasury, amount);
    }

    // ============ View Functions ============

    function getSponsorshipInfo(address wallet) external view returns (
        bool isActive,
        uint256 dailyLimit,
        uint256 perTxLimit,
        uint256 dailyUsed,
        uint256 dailyRemaining
    ) {
        SponsorshipConfig storage config = sponsorships[wallet];
        uint256 currentDay = block.timestamp / 1 days;
        uint256 used = config.lastResetDay < currentDay ? 0 : config.dailyUsed;

        return (
            config.isActive,
            config.dailyLimit,
            config.perTxLimit,
            used,
            config.dailyLimit > used ? config.dailyLimit - used : 0
        );
    }

    function getTokenPrice(address token) external view returns (uint256) {
        return tokenPrices[token];
    }

    function estimateTokenCost(address token, uint256 gasAmount) external view returns (uint256) {
        return _calculateTokenCost(token, gasAmount);
    }

    receive() external payable {}
}

// ERC-4337 PackedUserOperation struct (same as in wallet)
struct PackedUserOperation {
    address sender;
    uint256 nonce;
    bytes initCode;
    bytes callData;
    bytes32 accountGasLimits;
    uint256 preVerificationGas;
    bytes32 gasFees;
    bytes paymasterAndData;
    bytes signature;
}
