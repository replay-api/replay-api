// SPDX-License-Identifier: MIT
/**
 * @title LeetLedger Tests
 * @notice Tests for immutable distributed ledger contract
 */

const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("LeetLedger", function () {
  let admin, recorder, user1, user2, attacker;
  let ledger, usdc;
  const RECORDER_ROLE = ethers.keccak256(ethers.toUtf8Bytes("RECORDER_ROLE"));

  // Category constants
  const CAT_DEPOSIT = ethers.keccak256(ethers.toUtf8Bytes("DEPOSIT"));
  const CAT_WITHDRAWAL = ethers.keccak256(ethers.toUtf8Bytes("WITHDRAWAL"));
  const CAT_ENTRY_FEE = ethers.keccak256(ethers.toUtf8Bytes("ENTRY_FEE"));
  const CAT_PRIZE = ethers.keccak256(ethers.toUtf8Bytes("PRIZE"));
  const CAT_REFUND = ethers.keccak256(ethers.toUtf8Bytes("REFUND"));
  const CAT_PLATFORM_FEE = ethers.keccak256(ethers.toUtf8Bytes("PLATFORM_FEE"));
  const CAT_TRANSFER = ethers.keccak256(ethers.toUtf8Bytes("TRANSFER"));

  beforeEach(async function () {
    [admin, recorder, user1, user2, attacker] = await ethers.getSigners();

    // Deploy mock USDC for token address
    const MockUSDC = await ethers.getContractFactory("MockUSDC");
    usdc = await MockUSDC.deploy(admin.address);

    // Deploy LeetLedger
    const LeetLedger = await ethers.getContractFactory("LeetLedger");
    ledger = await LeetLedger.deploy();

    // Grant recorder role
    await ledger.grantRole(RECORDER_ROLE, recorder.address);
  });

  describe("Deployment", function () {
    it("should set admin with admin role", async function () {
      const DEFAULT_ADMIN_ROLE = ethers.ZeroHash;
      expect(await ledger.hasRole(DEFAULT_ADMIN_ROLE, admin.address)).to.be.true;
    });

    it("should set admin with recorder role", async function () {
      expect(await ledger.hasRole(RECORDER_ROLE, admin.address)).to.be.true;
    });

    it("should initialize merkle root", async function () {
      const root = await ledger.currentMerkleRoot();
      expect(root).to.not.equal(ethers.ZeroHash);
    });

    it("should start with zero entries", async function () {
      expect(await ledger.totalEntries()).to.equal(0);
    });
  });

  describe("Recording Entries", function () {
    it("should record a deposit entry", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("tx-001"));
      const amount = ethers.parseUnits("100", 6);

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        await usdc.getAddress(),
        amount,
        CAT_DEPOSIT,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      expect(await ledger.totalEntries()).to.equal(1);
      expect(await ledger.transactionExists(txId)).to.be.true;
    });

    it("should emit EntryRecorded event", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("tx-002"));
      const amount = ethers.parseUnits("50", 6);

      await expect(
        ledger.connect(recorder).recordEntry(
          txId,
          user1.address,
          await usdc.getAddress(),
          amount,
          CAT_DEPOSIT,
          ethers.ZeroHash,
          ethers.ZeroHash
        )
      ).to.emit(ledger, "EntryRecorded");
    });

    it("should update account balance", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("tx-003"));
      const amount = ethers.parseUnits("100", 6);
      const tokenAddr = await usdc.getAddress();

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        tokenAddr,
        amount,
        CAT_DEPOSIT,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      const balance = await ledger.accountBalances(user1.address, tokenAddr);
      expect(balance).to.equal(amount);
    });

    it("should reject duplicate transaction ID", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("tx-004"));
      const amount = ethers.parseUnits("100", 6);

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        await usdc.getAddress(),
        amount,
        CAT_DEPOSIT,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      await expect(
        ledger.connect(recorder).recordEntry(
          txId,
          user1.address,
          await usdc.getAddress(),
          amount,
          CAT_DEPOSIT,
          ethers.ZeroHash,
          ethers.ZeroHash
        )
      ).to.be.revertedWith("Transaction exists");
    });

    it("should reject recording from non-recorder", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("tx-005"));

      await expect(
        ledger.connect(attacker).recordEntry(
          txId,
          user1.address,
          await usdc.getAddress(),
          ethers.parseUnits("100", 6),
          CAT_DEPOSIT,
          ethers.ZeroHash,
          ethers.ZeroHash
        )
      ).to.be.reverted;
    });
  });

  describe("Debit and Credit", function () {
    it("should track negative balance (debit)", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("tx-debit-001"));
      const amount = ethers.parseUnits("-50", 6);
      const tokenAddr = await usdc.getAddress();

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        tokenAddr,
        amount,
        CAT_ENTRY_FEE,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      const balance = await ledger.accountBalances(user1.address, tokenAddr);
      expect(balance).to.equal(amount);
    });

    it("should calculate correct running balance", async function () {
      const tokenAddr = await usdc.getAddress();

      // Deposit 100
      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("tx-1")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("100", 6),
        CAT_DEPOSIT,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      // Entry fee -10
      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("tx-2")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("-10", 6),
        CAT_ENTRY_FEE,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      // Prize +20
      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("tx-3")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("20", 6),
        CAT_PRIZE,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      // Balance: 100 - 10 + 20 = 110
      const balance = await ledger.accountBalances(user1.address, tokenAddr);
      expect(balance).to.equal(ethers.parseUnits("110", 6));
    });
  });

  describe("Match and Tournament Tracking", function () {
    it("should associate entry with match ID", async function () {
      const matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));
      const txId = ethers.keccak256(ethers.toUtf8Bytes("tx-match-001"));

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        await usdc.getAddress(),
        ethers.parseUnits("-10", 6),
        CAT_ENTRY_FEE,
        matchId,
        ethers.ZeroHash
      );

      // Verify entry was recorded and can be retrieved
      expect(await ledger.transactionExists(txId)).to.be.true;
      const entryIndex = await ledger.entryIndexByTxId(txId);
      const entry = await ledger.entries(entryIndex);
      expect(entry.matchId).to.equal(matchId);
    });

    it("should associate entry with tournament ID", async function () {
      const tournamentId = ethers.keccak256(ethers.toUtf8Bytes("tournament-001"));
      const txId = ethers.keccak256(ethers.toUtf8Bytes("tx-tournament-001"));

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        await usdc.getAddress(),
        ethers.parseUnits("100", 6),
        CAT_PRIZE,
        ethers.ZeroHash,
        tournamentId
      );

      // Verify entry was recorded with tournament ID
      expect(await ledger.transactionExists(txId)).to.be.true;
      const entryIndex = await ledger.entryIndexByTxId(txId);
      const entry = await ledger.entries(entryIndex);
      expect(entry.tournamentId).to.equal(tournamentId);
    });
  });

  describe("Batch Recording", function () {
    it("should record multiple entries in batch", async function () {
      const batchId = ethers.keccak256(ethers.toUtf8Bytes("batch-001"));
      const tokenAddr = await usdc.getAddress();

      // The recordBatch function takes separate arrays
      const txId1 = ethers.keccak256(ethers.toUtf8Bytes("batch-tx-1"));
      const txId2 = ethers.keccak256(ethers.toUtf8Bytes("batch-tx-2"));

      await ledger.connect(recorder).recordBatch(
        batchId,
        [txId1, txId2],
        [user1.address, user2.address],
        [tokenAddr, tokenAddr],
        [ethers.parseUnits("100", 6), ethers.parseUnits("50", 6)],
        [CAT_DEPOSIT, CAT_DEPOSIT],
        [ethers.ZeroHash, ethers.ZeroHash]
      );

      expect(await ledger.totalEntries()).to.equal(2);
      expect(await ledger.transactionExists(txId1)).to.be.true;
      expect(await ledger.transactionExists(txId2)).to.be.true;
    });

    it("should emit BatchRecorded event", async function () {
      const batchId = ethers.keccak256(ethers.toUtf8Bytes("batch-002"));
      const tokenAddr = await usdc.getAddress();
      const txId = ethers.keccak256(ethers.toUtf8Bytes("batch-tx-3"));

      await expect(
        ledger.connect(recorder).recordBatch(
          batchId,
          [txId],
          [user1.address],
          [tokenAddr],
          [ethers.parseUnits("100", 6)],
          [CAT_DEPOSIT],
          [ethers.ZeroHash]
        )
      ).to.emit(ledger, "BatchRecorded");
    });
  });

  describe("View Functions", function () {
    it("should track total entries", async function () {
      const tokenAddr = await usdc.getAddress();

      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("view-tx-1")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("100", 6),
        CAT_DEPOSIT,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("view-tx-2")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("-10", 6),
        CAT_ENTRY_FEE,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      expect(await ledger.totalEntries()).to.equal(2);
    });

    it("should get entry by index", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("view-tx-3"));

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        await usdc.getAddress(),
        ethers.parseUnits("100", 6),
        CAT_DEPOSIT,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      const entry = await ledger.entries(0);
      expect(entry.transactionId).to.equal(txId);
      expect(entry.account).to.equal(user1.address);
    });

    it("should lookup entry by transaction ID", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("view-tx-4"));

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        await usdc.getAddress(),
        ethers.parseUnits("200", 6),
        CAT_PRIZE,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      const entryIndex = await ledger.entryIndexByTxId(txId);
      expect(entryIndex).to.equal(0);
    });
  });

  describe("Category Constants", function () {
    it("should have correct category hashes", async function () {
      expect(await ledger.CAT_DEPOSIT()).to.equal(CAT_DEPOSIT);
      expect(await ledger.CAT_WITHDRAWAL()).to.equal(CAT_WITHDRAWAL);
      expect(await ledger.CAT_ENTRY_FEE()).to.equal(CAT_ENTRY_FEE);
      expect(await ledger.CAT_PRIZE()).to.equal(CAT_PRIZE);
      expect(await ledger.CAT_REFUND()).to.equal(CAT_REFUND);
      expect(await ledger.CAT_PLATFORM_FEE()).to.equal(CAT_PLATFORM_FEE);
      expect(await ledger.CAT_TRANSFER()).to.equal(CAT_TRANSFER);
    });
  });

  describe("Role Management", function () {
    it("should allow admin to grant recorder role", async function () {
      await ledger.connect(admin).grantRole(RECORDER_ROLE, user2.address);
      expect(await ledger.hasRole(RECORDER_ROLE, user2.address)).to.be.true;
    });

    it("should allow admin to revoke recorder role", async function () {
      await ledger.connect(admin).revokeRole(RECORDER_ROLE, recorder.address);
      expect(await ledger.hasRole(RECORDER_ROLE, recorder.address)).to.be.false;
    });
  });

  describe("Transfer Recording", function () {
    it("should record double-entry transfer", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("transfer-001"));
      const amount = ethers.parseUnits("50", 6);
      const tokenAddr = await usdc.getAddress();

      await ledger.connect(recorder).recordTransfer(
        txId,
        user1.address,
        user2.address,
        tokenAddr,
        amount,
        CAT_TRANSFER,
        ethers.ZeroHash
      );

      // Check both entries were recorded
      expect(await ledger.totalEntries()).to.equal(2);

      // Check balances: user1 should be debited, user2 credited
      const balance1 = await ledger.accountBalances(user1.address, tokenAddr);
      const balance2 = await ledger.accountBalances(user2.address, tokenAddr);
      expect(balance1).to.equal(-amount);
      expect(balance2).to.equal(amount);
    });

    it("should reject transfer with zero amount", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("transfer-002"));

      await expect(
        ledger.connect(recorder).recordTransfer(
          txId,
          user1.address,
          user2.address,
          await usdc.getAddress(),
          0,
          CAT_TRANSFER,
          ethers.ZeroHash
        )
      ).to.be.revertedWith("Zero amount");
    });

    it("should reject transfer to same account", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("transfer-003"));

      await expect(
        ledger.connect(recorder).recordTransfer(
          txId,
          user1.address,
          user1.address,
          await usdc.getAddress(),
          ethers.parseUnits("50", 6),
          CAT_TRANSFER,
          ethers.ZeroHash
        )
      ).to.be.revertedWith("Same account");
    });
  });

  describe("Extended View Functions", function () {
    let tokenAddr;

    beforeEach(async function () {
      tokenAddr = await usdc.getAddress();

      // Record multiple entries for testing
      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("ext-tx-1")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("100", 6),
        CAT_DEPOSIT,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("ext-tx-2")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("-10", 6),
        CAT_ENTRY_FEE,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("ext-tx-3")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("20", 6),
        CAT_PRIZE,
        ethers.ZeroHash,
        ethers.ZeroHash
      );
    });

    it("should get entry by index", async function () {
      const entry = await ledger.getEntry(0);
      expect(entry.account).to.equal(user1.address);
    });

    it("should reject invalid entry index", async function () {
      await expect(ledger.getEntry(999)).to.be.revertedWith("Invalid index");
    });

    it("should get entry by transaction ID", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("ext-tx-1"));
      const entry = await ledger.getEntryByTxId(txId);
      expect(entry.transactionId).to.equal(txId);
      expect(entry.account).to.equal(user1.address);
    });

    it("should reject non-existent transaction ID", async function () {
      const fakeTxId = ethers.keccak256(ethers.toUtf8Bytes("non-existent"));
      await expect(ledger.getEntryByTxId(fakeTxId)).to.be.revertedWith("Not found");
    });

    it("should get account entry count", async function () {
      const count = await ledger.getAccountEntryCount(user1.address);
      expect(count).to.equal(3);
    });

    it("should get account entries range", async function () {
      const entries = await ledger.getAccountEntriesRange(user1.address, 0, 2);
      expect(entries.length).to.equal(2);
      expect(entries[0].account).to.equal(user1.address);
    });

    it("should handle range beyond available entries", async function () {
      const entries = await ledger.getAccountEntriesRange(user1.address, 0, 100);
      expect(entries.length).to.equal(3);
    });

    it("should get account balance", async function () {
      // 100 - 10 + 20 = 110
      const balance = await ledger.getAccountBalance(user1.address, tokenAddr);
      expect(balance).to.equal(ethers.parseUnits("110", 6));
    });

    it("should get total entries", async function () {
      expect(await ledger.getTotalEntries()).to.equal(3);
    });

    it("should get current merkle root", async function () {
      const root = await ledger.getCurrentMerkleRoot();
      expect(root).to.not.equal(ethers.ZeroHash);
    });
  });

  describe("Match Entry Tracking", function () {
    it("should get all entries for a match", async function () {
      const matchId = ethers.keccak256(ethers.toUtf8Bytes("match-100"));
      const tokenAddr = await usdc.getAddress();

      // Record entries associated with match
      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("match-entry-1")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("-10", 6),
        CAT_ENTRY_FEE,
        matchId,
        ethers.ZeroHash
      );

      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("match-entry-2")),
        user2.address,
        tokenAddr,
        ethers.parseUnits("-10", 6),
        CAT_ENTRY_FEE,
        matchId,
        ethers.ZeroHash
      );

      await ledger.connect(recorder).recordEntry(
        ethers.keccak256(ethers.toUtf8Bytes("match-entry-3")),
        user1.address,
        tokenAddr,
        ethers.parseUnits("18", 6),
        CAT_PRIZE,
        matchId,
        ethers.ZeroHash
      );

      const entries = await ledger.getMatchEntries(matchId);
      expect(entries.length).to.equal(3);
    });
  });

  describe("Chain Integrity Verification", function () {
    beforeEach(async function () {
      const tokenAddr = await usdc.getAddress();

      // Record multiple entries to create a chain
      for (let i = 1; i <= 5; i++) {
        await ledger.connect(recorder).recordEntry(
          ethers.keccak256(ethers.toUtf8Bytes(`chain-tx-${i}`)),
          user1.address,
          tokenAddr,
          ethers.parseUnits("10", 6),
          CAT_DEPOSIT,
          ethers.ZeroHash,
          ethers.ZeroHash
        );
      }
    });

    it("should verify chain integrity", async function () {
      const isValid = await ledger.verifyChainIntegrity(0, 4);
      expect(isValid).to.be.true;
    });

    it("should reject invalid end index", async function () {
      await expect(ledger.verifyChainIntegrity(0, 999)).to.be.revertedWith("Invalid end index");
    });

    it("should reject invalid range", async function () {
      await expect(ledger.verifyChainIntegrity(3, 1)).to.be.revertedWith("Invalid range");
    });
  });

  describe("Entry Proof Generation", function () {
    it("should generate entry proof", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("proof-tx-1"));
      const tokenAddr = await usdc.getAddress();

      await ledger.connect(recorder).recordEntry(
        txId,
        user1.address,
        tokenAddr,
        ethers.parseUnits("100", 6),
        CAT_DEPOSIT,
        ethers.ZeroHash,
        ethers.ZeroHash
      );

      const proof = await ledger.generateEntryProof(0);
      expect(proof.entryHash).to.not.equal(ethers.ZeroHash);
      expect(proof.merkleRoot).to.not.equal(ethers.ZeroHash);
      expect(proof.blockNumber).to.be.gt(0);
    });

    it("should reject invalid index for proof", async function () {
      await expect(ledger.generateEntryProof(999)).to.be.revertedWith("Invalid index");
    });
  });

  describe("Input Validation", function () {
    it("should reject zero address account", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("zero-addr-tx"));

      await expect(
        ledger.connect(recorder).recordEntry(
          txId,
          ethers.ZeroAddress,
          await usdc.getAddress(),
          ethers.parseUnits("100", 6),
          CAT_DEPOSIT,
          ethers.ZeroHash,
          ethers.ZeroHash
        )
      ).to.be.revertedWith("Invalid account");
    });

    it("should reject zero amount", async function () {
      const txId = ethers.keccak256(ethers.toUtf8Bytes("zero-amount-tx"));

      await expect(
        ledger.connect(recorder).recordEntry(
          txId,
          user1.address,
          await usdc.getAddress(),
          0,
          CAT_DEPOSIT,
          ethers.ZeroHash,
          ethers.ZeroHash
        )
      ).to.be.revertedWith("Zero amount");
    });

    it("should reject batch with mismatched array lengths", async function () {
      const batchId = ethers.keccak256(ethers.toUtf8Bytes("batch-mismatch"));
      const tokenAddr = await usdc.getAddress();

      await expect(
        ledger.connect(recorder).recordBatch(
          batchId,
          [ethers.keccak256(ethers.toUtf8Bytes("tx-1"))],
          [user1.address, user2.address], // Mismatched length
          [tokenAddr],
          [ethers.parseUnits("100", 6)],
          [CAT_DEPOSIT],
          [ethers.ZeroHash]
        )
      ).to.be.revertedWith("Length mismatch");
    });
  });
});
