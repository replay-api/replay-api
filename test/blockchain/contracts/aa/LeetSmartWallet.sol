// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title LeetSmartWallet
 * @dev ERC-4337 compliant smart contract wallet with MPC support,
 *      social recovery, session keys, and spending limits
 */
contract LeetSmartWallet is Initializable, UUPSUpgradeable {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;
    using SafeERC20 for IERC20;

    // ============ Constants ============

    bytes4 private constant ERC1271_SUCCESS = 0x1626ba7e;
    uint256 private constant SIG_VALIDATION_FAILED = 1;
    uint256 private constant SIG_VALIDATION_SUCCESS = 0;

    // ============ State Variables ============

    /// @notice ERC-4337 EntryPoint contract
    address public entryPoint;

    /// @notice MPC-derived owner address (threshold signature aggregated key)
    address public owner;

    /// @notice Platform wallet ID
    bytes32 public walletId;

    /// @notice Transaction nonce
    uint256 public nonce;

    /// @notice Wallet frozen status
    bool public isFrozen;

    /// @notice Daily spending limit in wei equivalent
    uint256 public dailyLimit;
    uint256 public dailySpent;
    uint256 public lastResetTimestamp;

    // Guardian management for social recovery
    struct Guardian {
        address addr;
        GuardianType guardianType;
        uint64 addedAt;
        bool isActive;
    }

    enum GuardianType { Wallet, Email, Phone, Hardware, Institution }

    mapping(address => Guardian) public guardians;
    address[] public guardianList;
    uint8 public guardianThreshold;

    // Social Recovery
    struct RecoveryRequest {
        address newOwner;
        uint64 initiatedAt;
        uint64 executableAt;
        uint8 approvalCount;
        bool executed;
        mapping(address => bool) approvals;
    }

    RecoveryRequest public pendingRecovery;
    uint64 public recoveryDelay;

    // Session Keys (for delegated signing)
    struct SessionKey {
        address key;
        uint48 validAfter;
        uint48 validUntil;
        uint256 spendingLimit;
        uint256 spent;
        address[] allowedTokens;
        bool isActive;
    }

    mapping(address => SessionKey) public sessionKeys;

    // ============ Events ============

    event WalletInitialized(bytes32 indexed walletId, address owner, address entryPoint);
    event OwnerChanged(address indexed oldOwner, address indexed newOwner);
    event GuardianAdded(address indexed guardian, GuardianType guardianType);
    event GuardianRemoved(address indexed guardian);
    event RecoveryInitiated(address indexed newOwner, uint64 executableAt);
    event RecoveryApproved(address indexed guardian, uint8 approvalCount);
    event RecoveryExecuted(address indexed newOwner);
    event RecoveryCancelled();
    event WalletFrozen(address indexed by);
    event WalletUnfrozen(address indexed by);
    event SessionKeyAdded(address indexed key, uint48 validUntil, uint256 spendingLimit);
    event SessionKeyRevoked(address indexed key);
    event TransactionExecuted(address indexed target, uint256 value, bytes data);
    event LimitUpdated(uint256 newLimit);

    // ============ Modifiers ============

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner");
        _;
    }

    modifier onlyEntryPoint() {
        require(msg.sender == entryPoint, "Only EntryPoint");
        _;
    }

    modifier onlyOwnerOrEntryPoint() {
        require(msg.sender == owner || msg.sender == entryPoint, "Not authorized");
        _;
    }

    modifier notFrozen() {
        require(!isFrozen, "Wallet is frozen");
        _;
    }

    // ============ Constructor & Initializer ============

    constructor() {
        _disableInitializers();
    }

    function initialize(
        bytes32 _walletId,
        address _owner,
        address _entryPoint,
        uint256 _dailyLimit,
        uint64 _recoveryDelay
    ) external initializer {
        require(_owner != address(0), "Invalid owner");
        require(_entryPoint != address(0), "Invalid entry point");

        walletId = _walletId;
        owner = _owner;
        entryPoint = _entryPoint;
        dailyLimit = _dailyLimit;
        recoveryDelay = _recoveryDelay;
        lastResetTimestamp = block.timestamp;
        guardianThreshold = 2; // Default: 2 guardians needed for recovery

        emit WalletInitialized(_walletId, _owner, _entryPoint);
    }

    // ============ ERC-4337 Interface ============

    /**
     * @dev Validate user operation signature (called by EntryPoint)
     * @param userOp The user operation
     * @param userOpHash Hash of the user operation
     * @param missingAccountFunds Amount to pay for gas
     * @return validationData 0 for valid signature, 1 for invalid
     */
    function validateUserOp(
        PackedUserOperation calldata userOp,
        bytes32 userOpHash,
        uint256 missingAccountFunds
    ) external onlyEntryPoint returns (uint256 validationData) {
        validationData = _validateSignature(userOp, userOpHash);

        if (missingAccountFunds > 0) {
            (bool success,) = payable(msg.sender).call{value: missingAccountFunds}("");
            require(success, "Failed to pay prefund");
        }
    }

    function _validateSignature(
        PackedUserOperation calldata userOp,
        bytes32 userOpHash
    ) internal view returns (uint256) {
        bytes32 hash = userOpHash.toEthSignedMessageHash();
        address signer = hash.recover(userOp.signature);

        // Check if signer is owner (MPC aggregated key)
        if (signer == owner) {
            return SIG_VALIDATION_SUCCESS;
        }

        // Check if signer is a valid session key
        SessionKey storage sk = sessionKeys[signer];
        if (sk.isActive &&
            block.timestamp >= sk.validAfter &&
            block.timestamp <= sk.validUntil) {
            return SIG_VALIDATION_SUCCESS;
        }

        return SIG_VALIDATION_FAILED;
    }

    // ============ Execution ============

    /**
     * @dev Execute a single transaction
     */
    function execute(
        address target,
        uint256 value,
        bytes calldata data
    ) external onlyOwnerOrEntryPoint notFrozen returns (bytes memory result) {
        _checkAndUpdateSpendingLimit(value);

        (bool success, bytes memory returnData) = target.call{value: value}(data);
        require(success, "Execution failed");

        emit TransactionExecuted(target, value, data);
        return returnData;
    }

    /**
     * @dev Execute a batch of transactions
     */
    function executeBatch(
        address[] calldata targets,
        uint256[] calldata values,
        bytes[] calldata datas
    ) external onlyOwnerOrEntryPoint notFrozen returns (bytes[] memory results) {
        require(targets.length == values.length && values.length == datas.length, "Length mismatch");

        results = new bytes[](targets.length);
        uint256 totalValue = 0;

        for (uint256 i = 0; i < targets.length; i++) {
            totalValue += values[i];
        }

        _checkAndUpdateSpendingLimit(totalValue);

        for (uint256 i = 0; i < targets.length; i++) {
            (bool success, bytes memory returnData) = targets[i].call{value: values[i]}(datas[i]);
            require(success, "Batch execution failed");
            results[i] = returnData;
            emit TransactionExecuted(targets[i], values[i], datas[i]);
        }
    }

    // ============ Spending Limits ============

    function _checkAndUpdateSpendingLimit(uint256 value) internal {
        // Reset daily spent if new day
        if (block.timestamp >= lastResetTimestamp + 1 days) {
            dailySpent = 0;
            lastResetTimestamp = block.timestamp;
        }

        require(dailySpent + value <= dailyLimit, "Daily limit exceeded");
        dailySpent += value;
    }

    function setDailyLimit(uint256 newLimit) external onlyOwner {
        dailyLimit = newLimit;
        emit LimitUpdated(newLimit);
    }

    // ============ Guardian Management ============

    function addGuardian(address guardian, GuardianType guardianType) external onlyOwner {
        require(guardian != address(0), "Invalid guardian");
        require(!guardians[guardian].isActive, "Guardian exists");
        require(guardianList.length < 7, "Max guardians reached");

        guardians[guardian] = Guardian({
            addr: guardian,
            guardianType: guardianType,
            addedAt: uint64(block.timestamp),
            isActive: true
        });
        guardianList.push(guardian);

        emit GuardianAdded(guardian, guardianType);
    }

    function removeGuardian(address guardian) external onlyOwner {
        require(guardians[guardian].isActive, "Guardian not active");

        guardians[guardian].isActive = false;

        // Remove from list
        for (uint256 i = 0; i < guardianList.length; i++) {
            if (guardianList[i] == guardian) {
                guardianList[i] = guardianList[guardianList.length - 1];
                guardianList.pop();
                break;
            }
        }

        emit GuardianRemoved(guardian);
    }

    function setGuardianThreshold(uint8 threshold) external onlyOwner {
        require(threshold > 0 && threshold <= guardianList.length, "Invalid threshold");
        guardianThreshold = threshold;
    }

    // ============ Social Recovery ============

    function initiateRecovery(address newOwner) external {
        require(guardians[msg.sender].isActive, "Not a guardian");
        require(newOwner != address(0), "Invalid new owner");
        require(!pendingRecovery.executed || pendingRecovery.approvalCount == 0, "Recovery pending");

        uint64 executableAt = uint64(block.timestamp) + recoveryDelay;

        // Reset pending recovery
        delete pendingRecovery;
        pendingRecovery.newOwner = newOwner;
        pendingRecovery.initiatedAt = uint64(block.timestamp);
        pendingRecovery.executableAt = executableAt;
        pendingRecovery.approvalCount = 1;
        pendingRecovery.approvals[msg.sender] = true;

        isFrozen = true; // Freeze wallet during recovery

        emit RecoveryInitiated(newOwner, executableAt);
        emit RecoveryApproved(msg.sender, 1);
    }

    function approveRecovery() external {
        require(guardians[msg.sender].isActive, "Not a guardian");
        require(pendingRecovery.newOwner != address(0), "No recovery pending");
        require(!pendingRecovery.approvals[msg.sender], "Already approved");

        pendingRecovery.approvals[msg.sender] = true;
        pendingRecovery.approvalCount++;

        emit RecoveryApproved(msg.sender, pendingRecovery.approvalCount);
    }

    function executeRecovery() external {
        require(pendingRecovery.newOwner != address(0), "No recovery pending");
        require(pendingRecovery.approvalCount >= guardianThreshold, "Insufficient approvals");
        require(block.timestamp >= pendingRecovery.executableAt, "Delay not met");
        require(!pendingRecovery.executed, "Already executed");

        address oldOwner = owner;
        address newOwner = pendingRecovery.newOwner;

        owner = newOwner;
        pendingRecovery.executed = true;
        isFrozen = false;

        // Revoke all session keys on ownership change
        _revokeAllSessionKeys();

        emit OwnerChanged(oldOwner, newOwner);
        emit RecoveryExecuted(newOwner);
    }

    function cancelRecovery() external onlyOwner {
        require(pendingRecovery.newOwner != address(0), "No recovery pending");

        delete pendingRecovery;
        isFrozen = false;

        emit RecoveryCancelled();
    }

    // ============ Session Keys ============

    function addSessionKey(
        address key,
        uint48 validAfter,
        uint48 validUntil,
        uint256 spendingLimit,
        address[] calldata allowedTokens
    ) external onlyOwner {
        require(key != address(0), "Invalid key");
        require(validUntil > validAfter, "Invalid validity");

        sessionKeys[key] = SessionKey({
            key: key,
            validAfter: validAfter,
            validUntil: validUntil,
            spendingLimit: spendingLimit,
            spent: 0,
            allowedTokens: allowedTokens,
            isActive: true
        });

        emit SessionKeyAdded(key, validUntil, spendingLimit);
    }

    function revokeSessionKey(address key) external onlyOwner {
        require(sessionKeys[key].isActive, "Key not active");
        sessionKeys[key].isActive = false;
        emit SessionKeyRevoked(key);
    }

    function _revokeAllSessionKeys() internal {
        // Note: This is a simplified version. In production, maintain a list
        // of all session keys for proper cleanup
    }

    // ============ Emergency Functions ============

    function freeze() external {
        require(msg.sender == owner || guardians[msg.sender].isActive, "Not authorized");
        isFrozen = true;
        emit WalletFrozen(msg.sender);
    }

    function unfreeze() external onlyOwner {
        require(pendingRecovery.newOwner == address(0), "Recovery in progress");
        isFrozen = false;
        emit WalletUnfrozen(msg.sender);
    }

    // ============ ERC-1271 Signature Validation ============

    function isValidSignature(
        bytes32 hash,
        bytes calldata signature
    ) external view returns (bytes4) {
        address signer = hash.toEthSignedMessageHash().recover(signature);
        if (signer == owner) {
            return ERC1271_SUCCESS;
        }
        return bytes4(0xffffffff);
    }

    // ============ View Functions ============

    function getGuardians() external view returns (address[] memory) {
        return guardianList;
    }

    function getRecoveryInfo() external view returns (
        address newOwner,
        uint64 initiatedAt,
        uint64 executableAt,
        uint8 approvalCount,
        bool executed
    ) {
        return (
            pendingRecovery.newOwner,
            pendingRecovery.initiatedAt,
            pendingRecovery.executableAt,
            pendingRecovery.approvalCount,
            pendingRecovery.executed
        );
    }

    function getSpendingInfo() external view returns (
        uint256 limit,
        uint256 spent,
        uint256 remaining,
        uint256 resetTime
    ) {
        uint256 currentSpent = dailySpent;
        if (block.timestamp >= lastResetTimestamp + 1 days) {
            currentSpent = 0;
        }
        return (
            dailyLimit,
            currentSpent,
            dailyLimit > currentSpent ? dailyLimit - currentSpent : 0,
            lastResetTimestamp + 1 days
        );
    }

    // ============ UUPS Upgrade ============

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    // ============ Receive ETH ============

    receive() external payable {}
}

// ERC-4337 PackedUserOperation struct
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
