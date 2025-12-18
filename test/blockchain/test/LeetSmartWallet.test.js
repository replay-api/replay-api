// SPDX-License-Identifier: MIT
/**
 * @title LeetSmartWallet Tests
 * @notice Comprehensive tests for ERC-4337 compliant smart wallet
 * @dev Tests cover: initialization, signature validation, execution,
 *      spending limits, guardians, session keys, and social recovery
 *
 * Object Calisthenics applied:
 * - Single responsibility per test
 * - No primitive obsession (using proper types)
 * - Small focused test functions
 */

const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("LeetSmartWallet", function () {
  // Test accounts
  let owner, entryPoint, guardian1, guardian2, guardian3, attacker, sessionKey;
  // Contract instances
  let wallet, walletFactory, usdc;
  // Constants
  const DAILY_LIMIT = ethers.parseEther("10"); // 10 ETH
  const RECOVERY_DELAY = 24 * 60 * 60; // 24 hours
  const WALLET_ID = ethers.keccak256(ethers.toUtf8Bytes("test-wallet-001"));

  beforeEach(async function () {
    [owner, entryPoint, guardian1, guardian2, guardian3, attacker, sessionKey] =
      await ethers.getSigners();

    // Deploy mock USDC for testing
    const MockUSDC = await ethers.getContractFactory("MockUSDC");
    usdc = await MockUSDC.deploy(owner.address);

    // Deploy wallet implementation
    const LeetSmartWallet = await ethers.getContractFactory("LeetSmartWallet");
    const walletImpl = await LeetSmartWallet.deploy();

    // For testing, we'll use a minimal proxy pattern
    // Deploy ERC1967Proxy from OpenZeppelin
    const proxyFactory = await ethers.getContractFactory("ERC1967Proxy", {
      libraries: {}
    });

    const initData = walletImpl.interface.encodeFunctionData("initialize", [
      WALLET_ID,
      owner.address,
      entryPoint.address,
      DAILY_LIMIT,
      RECOVERY_DELAY
    ]);

    const proxy = await proxyFactory.deploy(await walletImpl.getAddress(), initData);
    wallet = LeetSmartWallet.attach(await proxy.getAddress());
  });

  // ===========================================================================
  // Initialization Tests
  // ===========================================================================
  describe("Initialization", function () {
    it("should initialize with correct parameters", async function () {
      expect(await wallet.owner()).to.equal(owner.address);
      expect(await wallet.entryPoint()).to.equal(entryPoint.address);
      expect(await wallet.dailyLimit()).to.equal(DAILY_LIMIT);
      expect(await wallet.recoveryDelay()).to.equal(RECOVERY_DELAY);
      expect(await wallet.walletId()).to.equal(WALLET_ID);
      expect(await wallet.isFrozen()).to.equal(false);
    });

    it("should reject re-initialization", async function () {
      await expect(
        wallet.initialize(WALLET_ID, owner.address, entryPoint.address, DAILY_LIMIT, RECOVERY_DELAY)
      ).to.be.revertedWithCustomError(wallet, "InvalidInitialization");
    });

    it("should reject zero owner address", async function () {
      const LeetSmartWallet = await ethers.getContractFactory("LeetSmartWallet");
      const newWallet = await LeetSmartWallet.deploy();

      await expect(
        newWallet.initialize(WALLET_ID, ethers.ZeroAddress, entryPoint.address, DAILY_LIMIT, RECOVERY_DELAY)
      ).to.be.reverted; // Custom error or OwnableInvalidOwner
    });

    it("should reject zero entry point address", async function () {
      const LeetSmartWallet = await ethers.getContractFactory("LeetSmartWallet");
      const newWallet = await LeetSmartWallet.deploy();

      await expect(
        newWallet.initialize(WALLET_ID, owner.address, ethers.ZeroAddress, DAILY_LIMIT, RECOVERY_DELAY)
      ).to.be.reverted; // Custom error
    });
  });

  // ===========================================================================
  // Execution Tests
  // ===========================================================================
  describe("Execution", function () {
    beforeEach(async function () {
      // Fund the wallet
      await owner.sendTransaction({
        to: await wallet.getAddress(),
        value: ethers.parseEther("100")
      });
    });

    it("should execute transaction from owner", async function () {
      const recipient = guardian1.address;
      const amount = ethers.parseEther("1");
      const initialBalance = await ethers.provider.getBalance(recipient);

      await wallet.connect(owner).execute(recipient, amount, "0x");

      const finalBalance = await ethers.provider.getBalance(recipient);
      expect(finalBalance - initialBalance).to.equal(amount);
    });

    it("should execute transaction from entry point", async function () {
      const recipient = guardian1.address;
      const amount = ethers.parseEther("1");
      const initialBalance = await ethers.provider.getBalance(recipient);

      await wallet.connect(entryPoint).execute(recipient, amount, "0x");

      const finalBalance = await ethers.provider.getBalance(recipient);
      expect(finalBalance - initialBalance).to.equal(amount);
    });

    it("should reject execution from unauthorized account", async function () {
      await expect(
        wallet.connect(attacker).execute(attacker.address, ethers.parseEther("1"), "0x")
      ).to.be.revertedWith("Not authorized");
    });

    it("should execute batch transactions", async function () {
      const targets = [guardian1.address, guardian2.address];
      const values = [ethers.parseEther("1"), ethers.parseEther("2")];
      const datas = ["0x", "0x"];

      const initial1 = await ethers.provider.getBalance(guardian1.address);
      const initial2 = await ethers.provider.getBalance(guardian2.address);

      await wallet.connect(owner).executeBatch(targets, values, datas);

      expect(await ethers.provider.getBalance(guardian1.address) - initial1)
        .to.equal(values[0]);
      expect(await ethers.provider.getBalance(guardian2.address) - initial2)
        .to.equal(values[1]);
    });

    it("should reject execution when frozen", async function () {
      await wallet.connect(owner).freeze();

      await expect(
        wallet.connect(owner).execute(guardian1.address, ethers.parseEther("1"), "0x")
      ).to.be.revertedWith("Wallet is frozen");
    });
  });

  // ===========================================================================
  // Spending Limits Tests
  // ===========================================================================
  describe("Spending Limits", function () {
    beforeEach(async function () {
      await owner.sendTransaction({
        to: await wallet.getAddress(),
        value: ethers.parseEther("100")
      });
    });

    it("should enforce daily spending limit", async function () {
      // Spend up to the limit
      await wallet.connect(owner).execute(guardian1.address, DAILY_LIMIT, "0x");

      // Try to spend more
      await expect(
        wallet.connect(owner).execute(guardian1.address, ethers.parseEther("1"), "0x")
      ).to.be.revertedWith("Daily limit exceeded");
    });

    it("should reset daily limit after 24 hours", async function () {
      // Spend the limit
      await wallet.connect(owner).execute(guardian1.address, DAILY_LIMIT, "0x");

      // Advance time by 24 hours
      await time.increase(24 * 60 * 60);

      // Should be able to spend again
      await wallet.connect(owner).execute(
        guardian1.address,
        ethers.parseEther("1"),
        "0x"
      );
    });

    it("should allow owner to update daily limit", async function () {
      const newLimit = ethers.parseEther("20");
      await wallet.connect(owner).setDailyLimit(newLimit);
      expect(await wallet.dailyLimit()).to.equal(newLimit);
    });

    it("should return correct spending info", async function () {
      const amount = ethers.parseEther("5");
      await wallet.connect(owner).execute(guardian1.address, amount, "0x");

      const info = await wallet.getSpendingInfo();
      expect(info.limit).to.equal(DAILY_LIMIT);
      expect(info.spent).to.equal(amount);
      expect(info.remaining).to.equal(DAILY_LIMIT - amount);
    });
  });

  // ===========================================================================
  // Guardian Management Tests
  // ===========================================================================
  describe("Guardian Management", function () {
    it("should add guardian", async function () {
      await wallet.connect(owner).addGuardian(guardian1.address, 0); // 0 = Wallet type

      const guardians = await wallet.getGuardians();
      expect(guardians).to.include(guardian1.address);
    });

    it("should reject duplicate guardian", async function () {
      await wallet.connect(owner).addGuardian(guardian1.address, 0);

      await expect(
        wallet.connect(owner).addGuardian(guardian1.address, 0)
      ).to.be.revertedWith("Guardian exists");
    });

    it("should enforce max guardian limit", async function () {
      const signers = await ethers.getSigners();

      // Add 7 guardians (max)
      for (let i = 0; i < 7; i++) {
        await wallet.connect(owner).addGuardian(signers[i + 10].address, 0);
      }

      // Try to add 8th
      await expect(
        wallet.connect(owner).addGuardian(signers[17].address, 0)
      ).to.be.revertedWith("Max guardians reached");
    });

    it("should remove guardian", async function () {
      await wallet.connect(owner).addGuardian(guardian1.address, 0);
      await wallet.connect(owner).removeGuardian(guardian1.address);

      const guardians = await wallet.getGuardians();
      expect(guardians).to.not.include(guardian1.address);
    });

    it("should set guardian threshold", async function () {
      await wallet.connect(owner).addGuardian(guardian1.address, 0);
      await wallet.connect(owner).addGuardian(guardian2.address, 0);

      await wallet.connect(owner).setGuardianThreshold(2);
      expect(await wallet.guardianThreshold()).to.equal(2);
    });
  });

  // ===========================================================================
  // Social Recovery Tests
  // ===========================================================================
  describe("Social Recovery", function () {
    let newOwner;

    beforeEach(async function () {
      newOwner = guardian3;

      // Add guardians
      await wallet.connect(owner).addGuardian(guardian1.address, 0);
      await wallet.connect(owner).addGuardian(guardian2.address, 0);
      await wallet.connect(owner).setGuardianThreshold(2);
    });

    it("should initiate recovery", async function () {
      await wallet.connect(guardian1).initiateRecovery(newOwner.address);

      const info = await wallet.getRecoveryInfo();
      expect(info.newOwner).to.equal(newOwner.address);
      expect(info.approvalCount).to.equal(1);
      expect(await wallet.isFrozen()).to.equal(true);
    });

    it("should reject recovery initiation from non-guardian", async function () {
      await expect(
        wallet.connect(attacker).initiateRecovery(newOwner.address)
      ).to.be.revertedWith("Not a guardian");
    });

    it("should approve recovery", async function () {
      await wallet.connect(guardian1).initiateRecovery(newOwner.address);
      await wallet.connect(guardian2).approveRecovery();

      const info = await wallet.getRecoveryInfo();
      expect(info.approvalCount).to.equal(2);
    });

    it("should execute recovery after threshold and delay", async function () {
      await wallet.connect(guardian1).initiateRecovery(newOwner.address);
      await wallet.connect(guardian2).approveRecovery();

      // Advance time past recovery delay
      await time.increase(RECOVERY_DELAY + 1);

      await wallet.executeRecovery();

      expect(await wallet.owner()).to.equal(newOwner.address);
      expect(await wallet.isFrozen()).to.equal(false);
    });

    it("should reject recovery execution before delay", async function () {
      await wallet.connect(guardian1).initiateRecovery(newOwner.address);
      await wallet.connect(guardian2).approveRecovery();

      await expect(wallet.executeRecovery()).to.be.revertedWith("Delay not met");
    });

    it("should reject recovery with insufficient approvals", async function () {
      await wallet.connect(guardian1).initiateRecovery(newOwner.address);

      // Only 1 approval, need 2
      await time.increase(RECOVERY_DELAY + 1);

      await expect(wallet.executeRecovery()).to.be.revertedWith("Insufficient approvals");
    });

    it("should allow owner to cancel recovery", async function () {
      await wallet.connect(guardian1).initiateRecovery(newOwner.address);
      await wallet.connect(owner).cancelRecovery();

      expect(await wallet.isFrozen()).to.equal(false);
      const info = await wallet.getRecoveryInfo();
      expect(info.newOwner).to.equal(ethers.ZeroAddress);
    });
  });

  // ===========================================================================
  // Session Keys Tests
  // ===========================================================================
  describe("Session Keys", function () {
    beforeEach(async function () {
      await owner.sendTransaction({
        to: await wallet.getAddress(),
        value: ethers.parseEther("100")
      });
    });

    it("should add session key", async function () {
      const now = await time.latest();
      const validAfter = now;
      const validUntil = now + 3600; // 1 hour
      const spendingLimit = ethers.parseEther("5");

      await wallet.connect(owner).addSessionKey(
        sessionKey.address,
        validAfter,
        validUntil,
        spendingLimit,
        []
      );

      const sk = await wallet.sessionKeys(sessionKey.address);
      expect(sk.isActive).to.equal(true);
      expect(sk.spendingLimit).to.equal(spendingLimit);
    });

    it("should revoke session key", async function () {
      const now = await time.latest();
      await wallet.connect(owner).addSessionKey(
        sessionKey.address,
        now,
        now + 3600,
        ethers.parseEther("5"),
        []
      );

      await wallet.connect(owner).revokeSessionKey(sessionKey.address);

      const sk = await wallet.sessionKeys(sessionKey.address);
      expect(sk.isActive).to.equal(false);
    });

    it("should reject session key with invalid validity period", async function () {
      const now = await time.latest();

      await expect(
        wallet.connect(owner).addSessionKey(
          sessionKey.address,
          now + 3600, // validAfter
          now,        // validUntil (before validAfter)
          ethers.parseEther("5"),
          []
        )
      ).to.be.revertedWith("Invalid validity");
    });
  });

  // ===========================================================================
  // Emergency Functions Tests
  // ===========================================================================
  describe("Emergency Functions", function () {
    it("should freeze wallet by owner", async function () {
      await wallet.connect(owner).freeze();
      expect(await wallet.isFrozen()).to.equal(true);
    });

    it("should freeze wallet by guardian", async function () {
      await wallet.connect(owner).addGuardian(guardian1.address, 0);
      await wallet.connect(guardian1).freeze();
      expect(await wallet.isFrozen()).to.equal(true);
    });

    it("should reject freeze from unauthorized", async function () {
      await expect(
        wallet.connect(attacker).freeze()
      ).to.be.revertedWith("Not authorized");
    });

    it("should unfreeze wallet by owner", async function () {
      await wallet.connect(owner).freeze();
      await wallet.connect(owner).unfreeze();
      expect(await wallet.isFrozen()).to.equal(false);
    });

    it("should reject unfreeze during recovery", async function () {
      await wallet.connect(owner).addGuardian(guardian1.address, 0);
      await wallet.connect(guardian1).initiateRecovery(guardian2.address);

      await expect(
        wallet.connect(owner).unfreeze()
      ).to.be.revertedWith("Recovery in progress");
    });
  });

  // ===========================================================================
  // ERC-1271 Signature Validation Tests
  // ===========================================================================
  describe("ERC-1271 Signature Validation", function () {
    it("should validate owner signature", async function () {
      // For ERC-1271, we pass the raw hash (before EIP-191 prefix)
      // The contract adds the prefix before recovery
      const message = "Test message";
      const rawHash = ethers.keccak256(ethers.toUtf8Bytes(message));
      // Sign the raw hash bytes (signMessage adds EIP-191 prefix internally)
      const signature = await owner.signMessage(ethers.getBytes(rawHash));

      const result = await wallet.isValidSignature(rawHash, signature);
      expect(result).to.equal("0x1626ba7e"); // ERC1271_SUCCESS
    });

    it("should reject invalid signature", async function () {
      const message = "Test message";
      const rawHash = ethers.keccak256(ethers.toUtf8Bytes(message));
      const signature = await attacker.signMessage(ethers.getBytes(rawHash));

      const result = await wallet.isValidSignature(rawHash, signature);
      expect(result).to.equal("0xffffffff");
    });
  });

  // ===========================================================================
  // ETH Receive Tests
  // ===========================================================================
  describe("ETH Handling", function () {
    it("should receive ETH", async function () {
      const amount = ethers.parseEther("1");
      await owner.sendTransaction({
        to: await wallet.getAddress(),
        value: amount
      });

      expect(await ethers.provider.getBalance(await wallet.getAddress()))
        .to.equal(amount);
    });
  });
});
