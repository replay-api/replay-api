// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";

/**
 * @title LeetLedger
 * @dev Immutable distributed ledger for LeetGaming platform
 * Records all financial transactions as source of truth
 * Append-only: entries can never be modified or deleted
 */
contract LeetLedger is AccessControl {
    bytes32 public constant RECORDER_ROLE = keccak256("RECORDER_ROLE");

    // Transaction categories
    bytes32 public constant CAT_DEPOSIT = keccak256("DEPOSIT");
    bytes32 public constant CAT_WITHDRAWAL = keccak256("WITHDRAWAL");
    bytes32 public constant CAT_ENTRY_FEE = keccak256("ENTRY_FEE");
    bytes32 public constant CAT_PRIZE = keccak256("PRIZE");
    bytes32 public constant CAT_REFUND = keccak256("REFUND");
    bytes32 public constant CAT_PLATFORM_FEE = keccak256("PLATFORM_FEE");
    bytes32 public constant CAT_TRANSFER = keccak256("TRANSFER");

    struct LedgerEntry {
        bytes32 transactionId;      // Unique transaction ID
        address account;            // User account
        address token;              // Token address
        int256 amount;              // Positive = credit, Negative = debit
        bytes32 category;           // Transaction category
        bytes32 matchId;            // Related match (optional)
        bytes32 tournamentId;       // Related tournament (optional)
        uint256 timestamp;          // Block timestamp
        uint256 blockNumber;        // Block number
        bytes32 previousHash;       // Hash of previous entry (chain)
        bytes32 merkleRoot;         // State merkle root at this point
    }

    // Entry storage
    LedgerEntry[] public entries;
    mapping(bytes32 => uint256) public entryIndexByTxId;
    mapping(bytes32 => bool) public transactionExists;

    // Account ledgers
    mapping(address => uint256[]) public accountEntries;

    // Match/Tournament ledgers
    mapping(bytes32 => uint256[]) public matchEntries;
    mapping(bytes32 => uint256[]) public tournamentEntries;

    // Running balance per account per token
    mapping(address => mapping(address => int256)) public accountBalances;

    // State tracking
    bytes32 public currentMerkleRoot;
    uint256 public totalEntries;

    // Events
    event EntryRecorded(
        bytes32 indexed transactionId,
        address indexed account,
        address indexed token,
        int256 amount,
        bytes32 category,
        uint256 entryIndex
    );

    event BatchRecorded(
        bytes32 indexed batchId,
        uint256 entryCount,
        bytes32 merkleRoot
    );

    event BalanceSnapshot(
        address indexed account,
        address indexed token,
        int256 balance,
        uint256 blockNumber
    );

    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(RECORDER_ROLE, msg.sender);

        // Initialize with genesis entry
        currentMerkleRoot = keccak256(abi.encodePacked(block.timestamp, block.chainid));
    }

    /**
     * @dev Record a single ledger entry
     * @param transactionId Unique transaction ID (from backend)
     * @param account User account address
     * @param token Token address
     * @param amount Amount (positive = credit, negative = debit)
     * @param category Transaction category
     * @param matchId Related match ID (bytes32(0) if none)
     * @param tournamentId Related tournament ID (bytes32(0) if none)
     */
    function recordEntry(
        bytes32 transactionId,
        address account,
        address token,
        int256 amount,
        bytes32 category,
        bytes32 matchId,
        bytes32 tournamentId
    ) external onlyRole(RECORDER_ROLE) returns (uint256 entryIndex) {
        require(!transactionExists[transactionId], "Transaction exists");
        require(account != address(0), "Invalid account");
        require(amount != 0, "Zero amount");

        // Get previous hash for chaining
        bytes32 previousHash = entries.length > 0
            ? _hashEntry(entries[entries.length - 1])
            : bytes32(0);

        // Update merkle root
        currentMerkleRoot = keccak256(abi.encodePacked(
            currentMerkleRoot,
            transactionId,
            account,
            amount
        ));

        // Create entry
        LedgerEntry memory entry = LedgerEntry({
            transactionId: transactionId,
            account: account,
            token: token,
            amount: amount,
            category: category,
            matchId: matchId,
            tournamentId: tournamentId,
            timestamp: block.timestamp,
            blockNumber: block.number,
            previousHash: previousHash,
            merkleRoot: currentMerkleRoot
        });

        entryIndex = entries.length;
        entries.push(entry);

        // Update indexes
        transactionExists[transactionId] = true;
        entryIndexByTxId[transactionId] = entryIndex;
        accountEntries[account].push(entryIndex);

        if (matchId != bytes32(0)) {
            matchEntries[matchId].push(entryIndex);
        }
        if (tournamentId != bytes32(0)) {
            tournamentEntries[tournamentId].push(entryIndex);
        }

        // Update running balance
        accountBalances[account][token] += amount;
        totalEntries++;

        emit EntryRecorded(transactionId, account, token, amount, category, entryIndex);

        return entryIndex;
    }

    /**
     * @dev Record multiple entries in a batch (gas efficient)
     * @param batchId Batch identifier
     * @param transactionIds Array of transaction IDs
     * @param accounts Array of account addresses
     * @param tokens Array of token addresses
     * @param amounts Array of amounts
     * @param categories Array of categories
     * @param matchIds Array of match IDs
     */
    function recordBatch(
        bytes32 batchId,
        bytes32[] calldata transactionIds,
        address[] calldata accounts,
        address[] calldata tokens,
        int256[] calldata amounts,
        bytes32[] calldata categories,
        bytes32[] calldata matchIds
    ) external onlyRole(RECORDER_ROLE) {
        require(
            transactionIds.length == accounts.length &&
            accounts.length == tokens.length &&
            tokens.length == amounts.length &&
            amounts.length == categories.length &&
            categories.length == matchIds.length,
            "Length mismatch"
        );

        for (uint256 i = 0; i < transactionIds.length; i++) {
            if (!transactionExists[transactionIds[i]] && amounts[i] != 0) {
                _recordEntryInternal(
                    transactionIds[i],
                    accounts[i],
                    tokens[i],
                    amounts[i],
                    categories[i],
                    matchIds[i],
                    bytes32(0)
                );
            }
        }

        emit BatchRecorded(batchId, transactionIds.length, currentMerkleRoot);
    }

    /**
     * @dev Record double-entry (debit one account, credit another)
     * @param transactionId Transaction ID
     * @param fromAccount Source account (debited)
     * @param toAccount Destination account (credited)
     * @param token Token address
     * @param amount Amount (positive)
     * @param category Transaction category
     * @param matchId Related match ID
     */
    function recordTransfer(
        bytes32 transactionId,
        address fromAccount,
        address toAccount,
        address token,
        uint256 amount,
        bytes32 category,
        bytes32 matchId
    ) external onlyRole(RECORDER_ROLE) {
        require(amount > 0, "Zero amount");
        require(fromAccount != toAccount, "Same account");

        // Debit from source
        bytes32 debitTxId = keccak256(abi.encodePacked(transactionId, "DEBIT"));
        _recordEntryInternal(
            debitTxId,
            fromAccount,
            token,
            -int256(amount),
            category,
            matchId,
            bytes32(0)
        );

        // Credit to destination
        bytes32 creditTxId = keccak256(abi.encodePacked(transactionId, "CREDIT"));
        _recordEntryInternal(
            creditTxId,
            toAccount,
            token,
            int256(amount),
            category,
            matchId,
            bytes32(0)
        );
    }

    /**
     * @dev Internal entry recording
     */
    function _recordEntryInternal(
        bytes32 transactionId,
        address account,
        address token,
        int256 amount,
        bytes32 category,
        bytes32 matchId,
        bytes32 tournamentId
    ) internal {
        bytes32 previousHash = entries.length > 0
            ? _hashEntry(entries[entries.length - 1])
            : bytes32(0);

        currentMerkleRoot = keccak256(abi.encodePacked(
            currentMerkleRoot,
            transactionId,
            account,
            amount
        ));

        LedgerEntry memory entry = LedgerEntry({
            transactionId: transactionId,
            account: account,
            token: token,
            amount: amount,
            category: category,
            matchId: matchId,
            tournamentId: tournamentId,
            timestamp: block.timestamp,
            blockNumber: block.number,
            previousHash: previousHash,
            merkleRoot: currentMerkleRoot
        });

        uint256 entryIndex = entries.length;
        entries.push(entry);

        transactionExists[transactionId] = true;
        entryIndexByTxId[transactionId] = entryIndex;
        accountEntries[account].push(entryIndex);

        if (matchId != bytes32(0)) {
            matchEntries[matchId].push(entryIndex);
        }

        accountBalances[account][token] += amount;
        totalEntries++;

        emit EntryRecorded(transactionId, account, token, amount, category, entryIndex);
    }

    /**
     * @dev Hash an entry for chaining
     */
    function _hashEntry(LedgerEntry memory entry) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(
            entry.transactionId,
            entry.account,
            entry.token,
            entry.amount,
            entry.category,
            entry.timestamp,
            entry.previousHash
        ));
    }

    // ============ View Functions ============

    function getEntry(uint256 index) external view returns (LedgerEntry memory) {
        require(index < entries.length, "Invalid index");
        return entries[index];
    }

    function getEntryByTxId(bytes32 transactionId) external view returns (LedgerEntry memory) {
        require(transactionExists[transactionId], "Not found");
        return entries[entryIndexByTxId[transactionId]];
    }

    function getAccountEntryCount(address account) external view returns (uint256) {
        return accountEntries[account].length;
    }

    function getAccountEntriesRange(
        address account,
        uint256 start,
        uint256 limit
    ) external view returns (LedgerEntry[] memory) {
        uint256[] storage indices = accountEntries[account];
        uint256 end = start + limit;
        if (end > indices.length) {
            end = indices.length;
        }

        LedgerEntry[] memory result = new LedgerEntry[](end - start);
        for (uint256 i = start; i < end; i++) {
            result[i - start] = entries[indices[i]];
        }
        return result;
    }

    function getMatchEntries(bytes32 matchId) external view returns (LedgerEntry[] memory) {
        uint256[] storage indices = matchEntries[matchId];
        LedgerEntry[] memory result = new LedgerEntry[](indices.length);
        for (uint256 i = 0; i < indices.length; i++) {
            result[i] = entries[indices[i]];
        }
        return result;
    }

    function getAccountBalance(address account, address token) external view returns (int256) {
        return accountBalances[account][token];
    }

    function getTotalEntries() external view returns (uint256) {
        return totalEntries;
    }

    function getCurrentMerkleRoot() external view returns (bytes32) {
        return currentMerkleRoot;
    }

    /**
     * @dev Verify chain integrity from start to end index
     */
    function verifyChainIntegrity(uint256 startIndex, uint256 endIndex) external view returns (bool) {
        require(endIndex < entries.length, "Invalid end index");
        require(startIndex <= endIndex, "Invalid range");

        for (uint256 i = startIndex + 1; i <= endIndex; i++) {
            bytes32 expectedHash = _hashEntry(entries[i - 1]);
            if (entries[i].previousHash != expectedHash) {
                return false;
            }
        }
        return true;
    }

    /**
     * @dev Generate proof of entry for external verification
     */
    function generateEntryProof(uint256 entryIndex) external view returns (
        bytes32 entryHash,
        bytes32 previousHash,
        bytes32 merkleRoot,
        uint256 blockNumber
    ) {
        require(entryIndex < entries.length, "Invalid index");
        LedgerEntry memory entry = entries[entryIndex];
        return (
            _hashEntry(entry),
            entry.previousHash,
            entry.merkleRoot,
            entry.blockNumber
        );
    }
}
