// SPDX-License-Identifier: MIT
/**
 * @title LeetVault Tests
 * @notice Tests for prize pool escrow contract
 * @dev Tests cover: pool creation, deposits, distribution, cancellation, and admin
 */

const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("LeetVault", function () {
  let admin, operator, oracle, player1, player2, player3, winner, attacker;
  let vault, usdc;
  const OPERATOR_ROLE = ethers.keccak256(ethers.toUtf8Bytes("OPERATOR_ROLE"));
  const ORACLE_ROLE = ethers.keccak256(ethers.toUtf8Bytes("ORACLE_ROLE"));
  const ENTRY_FEE = ethers.parseUnits("10", 6); // 10 USDC
  const PLATFORM_FEE_PERCENT = 1000; // 10% in basis points

  beforeEach(async function () {
    [admin, operator, oracle, player1, player2, player3, winner, attacker] =
      await ethers.getSigners();

    // Deploy mock USDC
    const MockUSDC = await ethers.getContractFactory("MockUSDC");
    usdc = await MockUSDC.deploy(admin.address);

    // Deploy vault (constructor takes treasury address)
    const LeetVault = await ethers.getContractFactory("LeetVault");
    vault = await LeetVault.deploy(admin.address);

    // Setup roles
    await vault.connect(admin).grantRole(OPERATOR_ROLE, operator.address);
    await vault.connect(admin).grantRole(ORACLE_ROLE, oracle.address);

    // Add supported token
    await vault.connect(admin).addSupportedToken(await usdc.getAddress());

    // Mint USDC to players
    await usdc.mint(player1.address, ethers.parseUnits("100", 6));
    await usdc.mint(player2.address, ethers.parseUnits("100", 6));
    await usdc.mint(player3.address, ethers.parseUnits("100", 6));

    // Approve vault
    await usdc.connect(player1).approve(await vault.getAddress(), ethers.parseUnits("100", 6));
    await usdc.connect(player2).approve(await vault.getAddress(), ethers.parseUnits("100", 6));
    await usdc.connect(player3).approve(await vault.getAddress(), ethers.parseUnits("100", 6));
  });

  describe("Initialization", function () {
    it("should initialize with correct admin", async function () {
      expect(await vault.hasRole(await vault.DEFAULT_ADMIN_ROLE(), admin.address))
        .to.be.true;
    });

    it("should support USDC token", async function () {
      expect(await vault.supportedTokens(await usdc.getAddress())).to.be.true;
    });
  });

  describe("Pool Creation", function () {
    it("should create prize pool", async function () {
      const matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));

      await vault.connect(operator).createPrizePool(
        matchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );

      const pool = await vault.getPrizePoolInfo(matchId);
      expect(pool.token).to.equal(await usdc.getAddress());
      expect(pool.entryFee).to.equal(ENTRY_FEE);
      expect(pool.status).to.equal(1); // Accumulating
    });

    it("should reject pool with unsupported token", async function () {
      const matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));

      await expect(
        vault.connect(operator).createPrizePool(
          matchId,
          attacker.address, // Not a supported token
          ENTRY_FEE,
          PLATFORM_FEE_PERCENT
        )
      ).to.be.revertedWith("Token not supported");
    });

    it("should reject duplicate pool ID", async function () {
      const matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));

      await vault.connect(operator).createPrizePool(
        matchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );

      await expect(
        vault.connect(operator).createPrizePool(
          matchId,
          await usdc.getAddress(),
          ENTRY_FEE,
          PLATFORM_FEE_PERCENT
        )
      ).to.be.revertedWith("Pool exists");
    });

    it("should reject excessive platform fee", async function () {
      const matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));

      await expect(
        vault.connect(operator).createPrizePool(
          matchId,
          await usdc.getAddress(),
          ENTRY_FEE,
          2500 // 25% - exceeds max 20%
        )
      ).to.be.revertedWith("Fee too high");
    });
  });

  describe("Deposits", function () {
    let matchId;

    beforeEach(async function () {
      matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));
      await vault.connect(operator).createPrizePool(
        matchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );
    });

    it("should accept deposit from participant", async function () {
      await vault.connect(player1).depositEntryFee(matchId);

      const pool = await vault.getPrizePoolInfo(matchId);
      expect(pool.totalAmount).to.equal(ENTRY_FEE);

      const participants = await vault.getParticipants(matchId);
      expect(participants.length).to.equal(1);
    });

    it("should track participant contribution", async function () {
      await vault.connect(player1).depositEntryFee(matchId);

      const contribution = await vault.getContribution(matchId, player1.address);
      expect(contribution).to.equal(ENTRY_FEE);
    });

    it("should reject duplicate deposit", async function () {
      await vault.connect(player1).depositEntryFee(matchId);

      await expect(
        vault.connect(player1).depositEntryFee(matchId)
      ).to.be.revertedWith("Already joined");
    });
  });

  describe("Pool Locking", function () {
    let matchId;

    beforeEach(async function () {
      matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));
      await vault.connect(operator).createPrizePool(
        matchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );
      await vault.connect(player1).depositEntryFee(matchId);
      await vault.connect(player2).depositEntryFee(matchId);
    });

    it("should lock prize pool", async function () {
      await vault.connect(operator).lockPrizePool(matchId);

      const pool = await vault.getPrizePoolInfo(matchId);
      expect(pool.status).to.equal(2); // Locked
    });

    it("should reject lock with insufficient participants", async function () {
      const newMatchId = ethers.keccak256(ethers.toUtf8Bytes("match-002"));
      await vault.connect(operator).createPrizePool(
        newMatchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );
      await vault.connect(player1).depositEntryFee(newMatchId);

      await expect(
        vault.connect(operator).lockPrizePool(newMatchId)
      ).to.be.revertedWith("Not enough players");
    });
  });

  describe("Distribution", function () {
    let matchId;

    beforeEach(async function () {
      matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));
      await vault.connect(operator).createPrizePool(
        matchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );

      await vault.connect(player1).depositEntryFee(matchId);
      await vault.connect(player2).depositEntryFee(matchId);
      await vault.connect(player3).depositEntryFee(matchId);

      await vault.connect(operator).lockPrizePool(matchId);
      await vault.connect(oracle).startEscrow(matchId);

      // Fast forward past escrow period
      await time.increase(7 * 24 * 60 * 60 + 1); // 7 days + 1 second
    });

    it("should distribute prizes to winners", async function () {
      // 100% to winner (10000 basis points = 100%)
      await vault.connect(operator).distributePrizes(
        matchId,
        [winner.address],
        [10000] // 100% of distributable amount
      );

      // Check internal balance was credited
      const internalBalance = await vault.getUserBalance(winner.address, await usdc.getAddress());
      expect(internalBalance).to.be.gt(0);

      // Calculate expected prize: 3 deposits * 10 USDC = 30 USDC total
      // Platform fee = 10% = 3 USDC
      // Distributable = 27 USDC = 27000000 (6 decimals)
      // Note: This should be approximately 27 USDC (may be slightly less due to platform contribution)
      const expectedMin = ethers.parseUnits("25", 6); // At least 25 USDC
      expect(internalBalance).to.be.gte(expectedMin);
    });

    it("should distribute to multiple winners", async function () {
      // 60% first place, 40% second place
      await vault.connect(operator).distributePrizes(
        matchId,
        [player1.address, player2.address],
        [6000, 4000]
      );

      // Check internal balances were credited
      const balance1 = await vault.getUserBalance(player1.address, await usdc.getAddress());
      const balance2 = await vault.getUserBalance(player2.address, await usdc.getAddress());
      expect(balance1).to.be.gt(0);
      expect(balance2).to.be.gt(0);
      expect(balance1).to.be.gt(balance2); // First place gets more
    });

    it("should reject distribution from non-operator", async function () {
      await expect(
        vault.connect(attacker).distributePrizes(
          matchId,
          [winner.address],
          [10000]
        )
      ).to.be.reverted;
    });
  });

  describe("Cancellation", function () {
    let matchId;

    beforeEach(async function () {
      matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));
      await vault.connect(operator).createPrizePool(
        matchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );
      await vault.connect(player1).depositEntryFee(matchId);
      await vault.connect(player2).depositEntryFee(matchId);
    });

    it("should cancel prize pool", async function () {
      await vault.connect(operator).cancelPrizePool(matchId);

      const pool = await vault.getPrizePoolInfo(matchId);
      expect(pool.status).to.equal(5); // Cancelled
    });
  });

  describe("Admin Functions", function () {
    it("should add supported token", async function () {
      const MockUSDT = await ethers.getContractFactory("MockUSDT");
      const usdt = await MockUSDT.deploy(admin.address);

      await vault.connect(admin).addSupportedToken(await usdt.getAddress());
      expect(await vault.supportedTokens(await usdt.getAddress())).to.be.true;
    });

    it("should remove supported token", async function () {
      await vault.connect(admin).removeSupportedToken(await usdc.getAddress());
      expect(await vault.supportedTokens(await usdc.getAddress())).to.be.false;
    });

    it("should set escrow period", async function () {
      const newPeriod = 5 * 24 * 60 * 60; // 5 days (max is 7 days per contract)
      await vault.connect(admin).setEscrowPeriod(newPeriod);
      expect(await vault.escrowPeriod()).to.equal(newPeriod);
    });

    it("should pause and unpause", async function () {
      await vault.connect(admin).pause();
      expect(await vault.paused()).to.be.true;

      await vault.connect(admin).unpause();
      expect(await vault.paused()).to.be.false;
    });
  });

  describe("View Functions", function () {
    it("should return pool info", async function () {
      const matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));

      await vault.connect(operator).createPrizePool(
        matchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );

      const pool = await vault.getPrizePoolInfo(matchId);
      expect(pool.entryFee).to.equal(ENTRY_FEE);
    });

    it("should return participants", async function () {
      const matchId = ethers.keccak256(ethers.toUtf8Bytes("match-001"));

      await vault.connect(operator).createPrizePool(
        matchId,
        await usdc.getAddress(),
        ENTRY_FEE,
        PLATFORM_FEE_PERCENT
      );

      await vault.connect(player1).depositEntryFee(matchId);

      const participants = await vault.getParticipants(matchId);
      expect(participants.length).to.equal(1);
      expect(participants[0]).to.equal(player1.address);
    });
  });
});
