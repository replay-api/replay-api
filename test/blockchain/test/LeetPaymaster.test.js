// SPDX-License-Identifier: MIT
/**
 * @title LeetPaymaster Tests
 * @notice Comprehensive tests for ERC-4337 Paymaster
 * @dev Tests cover: payment modes, sponsorship, gas credits,
 *      token payments, rate limiting, and admin functions
 *
 * Object Calisthenics applied:
 * - One level of indentation per method
 * - Small focused test cases
 * - Descriptive naming
 */

const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("LeetPaymaster", function () {
  // Test accounts
  let admin, operator, signer, treasury, user1, user2, attacker;
  // Contract instances
  let paymaster, usdc, entryPoint;
  // Constants
  const OPERATOR_ROLE = ethers.keccak256(ethers.toUtf8Bytes("OPERATOR_ROLE"));
  const SIGNER_ROLE = ethers.keccak256(ethers.toUtf8Bytes("SIGNER_ROLE"));
  const DEFAULT_ADMIN_ROLE = ethers.ZeroHash;
  const PRICE_PER_GAS = 1000000; // 1 USDC per 1M gas units

  beforeEach(async function () {
    [admin, operator, signer, treasury, user1, user2, attacker] =
      await ethers.getSigners();

    // Deploy mock entry point (using admin address as mock)
    entryPoint = admin;

    // Deploy mock USDC
    const MockUSDC = await ethers.getContractFactory("MockUSDC");
    usdc = await MockUSDC.deploy(admin.address);

    // Deploy paymaster
    const LeetPaymaster = await ethers.getContractFactory("LeetPaymaster");
    paymaster = await LeetPaymaster.deploy(
      admin.address,  // entryPoint (using admin as mock)
      treasury.address,
      signer.address
    );

    // Setup roles
    await paymaster.connect(admin).grantRole(OPERATOR_ROLE, operator.address);

    // Setup accepted token
    await paymaster.connect(admin).setAcceptedToken(await usdc.getAddress(), PRICE_PER_GAS);

    // Fund paymaster for gas sponsorship
    await admin.sendTransaction({
      to: await paymaster.getAddress(),
      value: ethers.parseEther("100")
    });
  });

  // ===========================================================================
  // Initialization Tests
  // ===========================================================================
  describe("Initialization", function () {
    it("should initialize with correct parameters", async function () {
      expect(await paymaster.entryPoint()).to.equal(entryPoint.address);
      expect(await paymaster.treasury()).to.equal(treasury.address);
      expect(await paymaster.verifyingSigner()).to.equal(signer.address);
    });

    it("should reject zero entry point", async function () {
      const LeetPaymaster = await ethers.getContractFactory("LeetPaymaster");
      await expect(
        LeetPaymaster.deploy(ethers.ZeroAddress, treasury.address, signer.address)
      ).to.be.revertedWith("Invalid entry point");
    });

    it("should reject zero treasury", async function () {
      const LeetPaymaster = await ethers.getContractFactory("LeetPaymaster");
      await expect(
        LeetPaymaster.deploy(entryPoint.address, ethers.ZeroAddress, signer.address)
      ).to.be.revertedWith("Invalid treasury");
    });

    it("should reject zero signer", async function () {
      const LeetPaymaster = await ethers.getContractFactory("LeetPaymaster");
      await expect(
        LeetPaymaster.deploy(entryPoint.address, treasury.address, ethers.ZeroAddress)
      ).to.be.revertedWith("Invalid signer");
    });

    it("should set correct roles", async function () {
      expect(await paymaster.hasRole(DEFAULT_ADMIN_ROLE, admin.address)).to.be.true;
      expect(await paymaster.hasRole(OPERATOR_ROLE, admin.address)).to.be.true;
      expect(await paymaster.hasRole(SIGNER_ROLE, signer.address)).to.be.true;
    });
  });

  // ===========================================================================
  // Sponsorship Configuration Tests
  // ===========================================================================
  describe("Sponsorship Configuration", function () {
    const dailyLimit = 1000000; // 1M gas units
    const perTxLimit = 100000;  // 100K gas units

    it("should configure sponsorship", async function () {
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        dailyLimit,
        perTxLimit,
        []
      );

      const info = await paymaster.getSponsorshipInfo(user1.address);
      expect(info.isActive).to.be.true;
      expect(info.dailyLimit).to.equal(dailyLimit);
      expect(info.perTxLimit).to.equal(perTxLimit);
    });

    it("should reject configuration from non-operator", async function () {
      await expect(
        paymaster.connect(attacker).configureSponsorship(
          user1.address,
          dailyLimit,
          perTxLimit,
          []
        )
      ).to.be.reverted;
    });

    it("should disable sponsorship", async function () {
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        dailyLimit,
        perTxLimit,
        []
      );

      await paymaster.connect(operator).disableSponsorship(user1.address);

      const info = await paymaster.getSponsorshipInfo(user1.address);
      expect(info.isActive).to.be.false;
    });

    it("should track daily usage correctly", async function () {
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        dailyLimit,
        perTxLimit,
        []
      );

      const info = await paymaster.getSponsorshipInfo(user1.address);
      expect(info.dailyUsed).to.equal(0);
      expect(info.dailyRemaining).to.equal(dailyLimit);
    });
  });

  // ===========================================================================
  // Gas Credits Tests
  // ===========================================================================
  describe("Gas Credits", function () {
    const creditAmount = 500000;

    it("should add gas credits", async function () {
      await paymaster.connect(operator).addGasCredits(user1.address, creditAmount);
      expect(await paymaster.gasCredits(user1.address)).to.equal(creditAmount);
    });

    it("should reject adding credits from non-operator", async function () {
      await expect(
        paymaster.connect(attacker).addGasCredits(user1.address, creditAmount)
      ).to.be.reverted;
    });

    it("should accumulate gas credits", async function () {
      await paymaster.connect(operator).addGasCredits(user1.address, creditAmount);
      await paymaster.connect(operator).addGasCredits(user1.address, creditAmount);

      expect(await paymaster.gasCredits(user1.address)).to.equal(creditAmount * 2);
    });
  });

  // ===========================================================================
  // Token Payment Tests
  // ===========================================================================
  describe("Token Payment", function () {
    it("should set accepted token", async function () {
      expect(await paymaster.acceptedTokens(await usdc.getAddress())).to.be.true;
      expect(await paymaster.tokenPrices(await usdc.getAddress())).to.equal(PRICE_PER_GAS);
    });

    it("should remove accepted token", async function () {
      await paymaster.connect(admin).removeAcceptedToken(await usdc.getAddress());

      expect(await paymaster.acceptedTokens(await usdc.getAddress())).to.be.false;
      expect(await paymaster.tokenPrices(await usdc.getAddress())).to.equal(0);
    });

    it("should estimate token cost correctly", async function () {
      const gasAmount = 100000;
      const expectedCost = (gasAmount * PRICE_PER_GAS) / 1000000;

      const cost = await paymaster.estimateTokenCost(await usdc.getAddress(), gasAmount);
      expect(cost).to.equal(expectedCost);
    });

    it("should reject setting token from non-admin", async function () {
      const MockUSDT = await ethers.getContractFactory("MockUSDT");
      const usdt = await MockUSDT.deploy(admin.address);

      await expect(
        paymaster.connect(operator).setAcceptedToken(await usdt.getAddress(), PRICE_PER_GAS)
      ).to.be.reverted;
    });
  });

  // ===========================================================================
  // Admin Functions Tests
  // ===========================================================================
  describe("Admin Functions", function () {
    it("should set treasury", async function () {
      await paymaster.connect(admin).setTreasury(user2.address);
      expect(await paymaster.treasury()).to.equal(user2.address);
    });

    it("should reject zero treasury", async function () {
      await expect(
        paymaster.connect(admin).setTreasury(ethers.ZeroAddress)
      ).to.be.revertedWith("Invalid treasury");
    });

    it("should set verifying signer", async function () {
      await paymaster.connect(admin).setVerifyingSigner(user2.address);

      expect(await paymaster.verifyingSigner()).to.equal(user2.address);
      expect(await paymaster.hasRole(SIGNER_ROLE, user2.address)).to.be.true;
      expect(await paymaster.hasRole(SIGNER_ROLE, signer.address)).to.be.false;
    });

    it("should reject zero signer", async function () {
      await expect(
        paymaster.connect(admin).setVerifyingSigner(ethers.ZeroAddress)
      ).to.be.revertedWith("Invalid signer");
    });

    it("should set min tx interval", async function () {
      const newInterval = 10;
      await paymaster.connect(admin).setMinTxInterval(newInterval);
      expect(await paymaster.minTxInterval()).to.equal(newInterval);
    });
  });

  // ===========================================================================
  // Fund Management Tests
  // ===========================================================================
  describe("Fund Management", function () {
    it("should accept ETH deposits", async function () {
      const depositAmount = ethers.parseEther("10");
      const initialBalance = await ethers.provider.getBalance(await paymaster.getAddress());

      await paymaster.deposit({ value: depositAmount });

      const finalBalance = await ethers.provider.getBalance(await paymaster.getAddress());
      expect(finalBalance - initialBalance).to.equal(depositAmount);
    });

    it("should withdraw ETH to treasury", async function () {
      const withdrawAmount = ethers.parseEther("10");
      const initialTreasuryBalance = await ethers.provider.getBalance(treasury.address);

      await paymaster.connect(admin).withdrawETH(withdrawAmount);

      const finalTreasuryBalance = await ethers.provider.getBalance(treasury.address);
      expect(finalTreasuryBalance - initialTreasuryBalance).to.equal(withdrawAmount);
    });

    it("should reject withdrawal exceeding balance", async function () {
      const balance = await ethers.provider.getBalance(await paymaster.getAddress());
      await expect(
        paymaster.connect(admin).withdrawETH(balance + 1n)
      ).to.be.revertedWith("Insufficient balance");
    });

    it("should withdraw tokens to treasury", async function () {
      const tokenAmount = ethers.parseUnits("1000", 6);

      // Mint tokens to paymaster
      await usdc.mint(await paymaster.getAddress(), tokenAmount);

      const initialTreasuryBalance = await usdc.balanceOf(treasury.address);
      await paymaster.connect(admin).withdrawToken(await usdc.getAddress(), tokenAmount);

      const finalTreasuryBalance = await usdc.balanceOf(treasury.address);
      expect(finalTreasuryBalance - initialTreasuryBalance).to.equal(tokenAmount);
    });

    it("should reject withdrawal from non-admin", async function () {
      await expect(
        paymaster.connect(attacker).withdrawETH(ethers.parseEther("1"))
      ).to.be.reverted;
    });
  });

  // ===========================================================================
  // View Functions Tests
  // ===========================================================================
  describe("View Functions", function () {
    it("should return correct sponsorship info with reset", async function () {
      const dailyLimit = 1000000;
      const perTxLimit = 100000;

      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        dailyLimit,
        perTxLimit,
        []
      );

      // Advance time to trigger daily reset
      await time.increase(86400);

      const info = await paymaster.getSponsorshipInfo(user1.address);
      expect(info.dailyUsed).to.equal(0);
      expect(info.dailyRemaining).to.equal(dailyLimit);
    });

    it("should return token price", async function () {
      expect(await paymaster.getTokenPrice(await usdc.getAddress())).to.equal(PRICE_PER_GAS);
    });

    it("should return zero for unknown token", async function () {
      expect(await paymaster.getTokenPrice(user1.address)).to.equal(0);
    });
  });

  // ===========================================================================
  // Statistics Tests
  // ===========================================================================
  describe("Statistics", function () {
    it("should track total gas sponsored", async function () {
      expect(await paymaster.totalGasSponsored()).to.equal(0);
    });

    it("should track total tokens collected", async function () {
      expect(await paymaster.totalTokensCollected()).to.equal(0);
    });
  });

  // ===========================================================================
  // Access Control Tests
  // ===========================================================================
  describe("Access Control", function () {
    it("should allow admin to grant operator role", async function () {
      await paymaster.connect(admin).grantRole(OPERATOR_ROLE, user2.address);
      expect(await paymaster.hasRole(OPERATOR_ROLE, user2.address)).to.be.true;
    });

    it("should allow admin to revoke operator role", async function () {
      await paymaster.connect(admin).revokeRole(OPERATOR_ROLE, operator.address);
      expect(await paymaster.hasRole(OPERATOR_ROLE, operator.address)).to.be.false;
    });

    it("should reject role grant from non-admin", async function () {
      await expect(
        paymaster.connect(attacker).grantRole(OPERATOR_ROLE, attacker.address)
      ).to.be.reverted;
    });
  });

  // ===========================================================================
  // UserOp Validation Tests (calling from entryPoint)
  // ===========================================================================
  describe("UserOp Validation", function () {
    // Helper to create a mock UserOp
    function createMockUserOp(sender, paymasterData) {
      return {
        sender: sender,
        nonce: 0,
        initCode: "0x",
        callData: "0x",
        accountGasLimits: ethers.ZeroHash,
        preVerificationGas: 21000,
        gasFees: ethers.ZeroHash,
        paymasterAndData: paymasterData,
        signature: "0x"
      };
    }

    it("should reject validation from non-entrypoint", async function () {
      const paymasterAddr = await paymaster.getAddress();
      // Mode 0 = Sponsored
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x00" // Sponsored mode
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      await expect(
        paymaster.connect(attacker).validatePaymasterUserOp(userOp, userOpHash, 100000)
      ).to.be.revertedWith("Only EntryPoint");
    });

    it("should validate sponsored user op", async function () {
      // Configure sponsorship for user1
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        1000000, // dailyLimit
        100000,  // perTxLimit
        []
      );

      const paymasterAddr = await paymaster.getAddress();
      // Mode 0 = Sponsored
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x00" // Sponsored mode
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      // Call from admin (acting as entryPoint)
      const result = await paymaster.connect(admin).validatePaymasterUserOp.staticCall(
        userOp,
        userOpHash,
        50000 // maxCost within limits
      );

      // Should return valid context (result[1] is validationData)
      expect(result[1]).to.equal(0);
    });

    it("should reject sponsored op without active sponsorship", async function () {
      const paymasterAddr = await paymaster.getAddress();
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x00" // Sponsored mode
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      await expect(
        paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 100000)
      ).to.be.revertedWith("Sponsorship invalid");
    });

    it("should reject sponsored op exceeding per-tx limit", async function () {
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        1000000, // dailyLimit
        50000,   // perTxLimit (lower than maxCost)
        []
      );

      const paymasterAddr = await paymaster.getAddress();
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x00"
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      await expect(
        paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 100000)
      ).to.be.revertedWith("Sponsorship invalid");
    });

    it("should validate gas credits payment mode", async function () {
      // Add gas credits
      await paymaster.connect(operator).addGasCredits(user1.address, 500000);

      const paymasterAddr = await paymaster.getAddress();
      // Mode 1 = GasCredits
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x01"
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      const result = await paymaster.connect(admin).validatePaymasterUserOp.staticCall(
        userOp,
        userOpHash,
        100000
      );

      expect(result[1]).to.equal(0);
    });

    it("should reject gas credits op with insufficient credits", async function () {
      // Add insufficient credits
      await paymaster.connect(operator).addGasCredits(user1.address, 1000);

      const paymasterAddr = await paymaster.getAddress();
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x01"
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      await expect(
        paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 100000)
      ).to.be.revertedWith("Insufficient credits");
    });

    it("should enforce rate limiting", async function () {
      // Set min interval
      await paymaster.connect(admin).setMinTxInterval(60);

      // Configure sponsorship
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        1000000,
        100000,
        []
      );

      const paymasterAddr = await paymaster.getAddress();
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x00"
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      // First call should succeed
      await paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 50000);

      // Second call should fail (too frequent)
      await expect(
        paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 50000)
      ).to.be.revertedWith("Too frequent");
    });

    it("should allow transactions after rate limit period", async function () {
      await paymaster.connect(admin).setMinTxInterval(60);

      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        1000000,
        100000,
        []
      );

      const paymasterAddr = await paymaster.getAddress();
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x00"
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      await paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 50000);

      // Advance time past the interval
      await time.increase(61);

      // Should succeed now
      const result = await paymaster.connect(admin).validatePaymasterUserOp.staticCall(
        userOp,
        userOpHash,
        50000
      );
      expect(result[1]).to.equal(0);
    });
  });

  // ===========================================================================
  // PostOp Tests
  // ===========================================================================
  describe("PostOp Processing", function () {
    it("should reject postOp from non-entrypoint", async function () {
      const context = ethers.AbiCoder.defaultAbiCoder().encode(
        ["address", "uint8", "uint256"],
        [user1.address, 0, 100000]
      );

      await expect(
        paymaster.connect(attacker).postOp(0, context, 50000, 0)
      ).to.be.revertedWith("Only EntryPoint");
    });

    it("should process sponsored postOp", async function () {
      // Configure sponsorship
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        1000000,
        100000,
        []
      );

      const context = ethers.AbiCoder.defaultAbiCoder().encode(
        ["address", "uint8", "uint256"],
        [user1.address, 0, 100000] // Sponsored mode
      );

      // Call postOp from admin (entryPoint)
      await paymaster.connect(admin).postOp(0, context, 50000, 0);

      // Check stats updated
      expect(await paymaster.totalGasSponsored()).to.equal(50000);

      // Check usage tracked
      const info = await paymaster.getSponsorshipInfo(user1.address);
      expect(info.dailyUsed).to.equal(50000);
    });

    it("should process gas credits postOp", async function () {
      // Add credits
      await paymaster.connect(operator).addGasCredits(user1.address, 500000);

      const context = ethers.AbiCoder.defaultAbiCoder().encode(
        ["address", "uint8", "uint256"],
        [user1.address, 1, 100000] // GasCredits mode
      );

      await paymaster.connect(admin).postOp(0, context, 50000, 0);

      // Credits should be deducted
      expect(await paymaster.gasCredits(user1.address)).to.equal(450000);
    });

    it("should reset daily usage on new day", async function () {
      // Configure sponsorship
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        1000000,
        100000,
        []
      );

      // Process first postOp
      const context = ethers.AbiCoder.defaultAbiCoder().encode(
        ["address", "uint8", "uint256"],
        [user1.address, 0, 100000]
      );

      await paymaster.connect(admin).postOp(0, context, 50000, 0);

      // Advance to next day
      await time.increase(86400);

      // Check sponsorship info - should show reset
      const info = await paymaster.getSponsorshipInfo(user1.address);
      expect(info.dailyUsed).to.equal(0);
      expect(info.dailyRemaining).to.equal(1000000);
    });
  });

  // ===========================================================================
  // Receive ETH Tests
  // ===========================================================================
  describe("Receive ETH", function () {
    it("should accept ETH via receive function", async function () {
      const amount = ethers.parseEther("5");
      const initialBalance = await ethers.provider.getBalance(await paymaster.getAddress());

      await admin.sendTransaction({
        to: await paymaster.getAddress(),
        value: amount
      });

      const finalBalance = await ethers.provider.getBalance(await paymaster.getAddress());
      expect(finalBalance - initialBalance).to.equal(amount);
    });
  });

  // ===========================================================================
  // Token Payment Mode Tests
  // ===========================================================================
  describe("Token Payment Mode", function () {
    function createMockUserOp(sender, paymasterData) {
      return {
        sender: sender,
        nonce: 0,
        initCode: "0x",
        callData: "0x",
        accountGasLimits: ethers.ZeroHash,
        preVerificationGas: 21000,
        gasFees: ethers.ZeroHash,
        paymasterAndData: paymasterData,
        signature: "0x"
      };
    }

    it("should validate token payment mode", async function () {
      const tokenAddr = await usdc.getAddress();
      const paymasterAddr = await paymaster.getAddress();
      const tokenAmount = ethers.parseUnits("10", 6);

      // Mint tokens to user and approve paymaster
      await usdc.mint(user1.address, tokenAmount);
      await usdc.connect(user1).approve(paymasterAddr, tokenAmount);

      // Mode 2 = TokenPayment (20 bytes token + 32 bytes amount)
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x02", // TokenPayment mode
        tokenAddr,
        ethers.zeroPadValue(ethers.toBeHex(tokenAmount), 32)
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      const result = await paymaster.connect(admin).validatePaymasterUserOp.staticCall(
        userOp,
        userOpHash,
        100000
      );

      expect(result[1]).to.equal(0);
    });

    it("should reject token payment with unsupported token", async function () {
      const paymasterAddr = await paymaster.getAddress();
      const tokenAmount = ethers.parseUnits("10", 6);

      // Use an unsupported token address
      const fakeToken = user2.address;

      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x02",
        fakeToken,
        ethers.zeroPadValue(ethers.toBeHex(tokenAmount), 32)
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      await expect(
        paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 100000)
      ).to.be.revertedWith("Token not accepted");
    });

    it("should process token payment postOp with refund", async function () {
      const tokenAddr = await usdc.getAddress();
      const tokenAmount = ethers.parseUnits("10", 6);
      const actualCost = ethers.parseUnits("5", 6);

      // Mint tokens to paymaster (simulating tokens transferred during validatePaymasterUserOp)
      await usdc.mint(await paymaster.getAddress(), tokenAmount);

      // Context for token payment: [wallet, mode, maxCost, token, tokenAmount]
      const context = ethers.AbiCoder.defaultAbiCoder().encode(
        ["address", "uint8", "uint256", "address", "uint256"],
        [user1.address, 2, 100000, tokenAddr, tokenAmount] // TokenPayment mode
      );

      const treasuryBalanceBefore = await usdc.balanceOf(treasury.address);
      const user1BalanceBefore = await usdc.balanceOf(user1.address);

      // actualGasCost determines how much to actually charge
      // Using PRICE_PER_GAS = 1000000 and gasUsed = 5000000, cost = 5 USDC
      await paymaster.connect(admin).postOp(0, context, 5000000, 0);

      // Treasury should receive actual cost
      const treasuryBalanceAfter = await usdc.balanceOf(treasury.address);
      expect(treasuryBalanceAfter - treasuryBalanceBefore).to.equal(actualCost);

      // User should receive refund (10 - 5 = 5 USDC)
      const user1BalanceAfter = await usdc.balanceOf(user1.address);
      expect(user1BalanceAfter - user1BalanceBefore).to.equal(actualCost);

      // Total tokens collected should be updated
      expect(await paymaster.totalTokensCollected()).to.equal(actualCost);
    });
  });

  // ===========================================================================
  // Verifying Signature Mode Tests
  // ===========================================================================
  describe("Verifying Signature Mode", function () {
    function createMockUserOp(sender, paymasterData) {
      return {
        sender: sender,
        nonce: 0,
        initCode: "0x",
        callData: "0x",
        accountGasLimits: ethers.ZeroHash,
        preVerificationGas: 21000,
        gasFees: ethers.ZeroHash,
        paymasterAndData: paymasterData,
        signature: "0x"
      };
    }

    it("should validate verified free mode with valid signature", async function () {
      const paymasterAddr = await paymaster.getAddress();
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      // Create the hash that needs to be signed
      const chainId = await ethers.provider.getNetwork().then(n => n.chainId);
      const hash = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["bytes32", "address", "uint256"],
          [userOpHash, paymasterAddr, chainId]
        )
      );
      const ethSignedHash = ethers.hashMessage(ethers.getBytes(hash));

      // Sign with the signer
      const signature = await signer.signMessage(ethers.getBytes(hash));

      // Mode 3 = VerifiedFree (1 byte mode + 65 bytes signature)
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x03", // VerifiedFree mode
        signature
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);

      const result = await paymaster.connect(admin).validatePaymasterUserOp.staticCall(
        userOp,
        userOpHash,
        100000
      );

      expect(result[1]).to.equal(0);
    });

    it("should reject verified free mode with invalid signature", async function () {
      const paymasterAddr = await paymaster.getAddress();
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      // Sign with a non-signer account
      const chainId = await ethers.provider.getNetwork().then(n => n.chainId);
      const hash = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["bytes32", "address", "uint256"],
          [userOpHash, paymasterAddr, chainId]
        )
      );
      const signature = await attacker.signMessage(ethers.getBytes(hash));

      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x03",
        signature
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);

      await expect(
        paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 100000)
      ).to.be.revertedWith("Invalid signature");
    });
  });

  // ===========================================================================
  // Invalid Payment Mode Tests
  // ===========================================================================
  describe("Invalid Payment Mode", function () {
    function createMockUserOp(sender, paymasterData) {
      return {
        sender: sender,
        nonce: 0,
        initCode: "0x",
        callData: "0x",
        accountGasLimits: ethers.ZeroHash,
        preVerificationGas: 21000,
        gasFees: ethers.ZeroHash,
        paymasterAndData: paymasterData,
        signature: "0x"
      };
    }

    it("should reject invalid payment mode", async function () {
      const paymasterAddr = await paymaster.getAddress();

      // Mode 4 = Invalid (does not exist)
      // Solidity reverts with panic when enum conversion fails
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x04" // Invalid mode
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      await expect(
        paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 100000)
      ).to.be.reverted;
    });
  });

  // ===========================================================================
  // Daily Limit Reset in Sponsorship Validation
  // ===========================================================================
  describe("Sponsorship Daily Reset", function () {
    function createMockUserOp(sender, paymasterData) {
      return {
        sender: sender,
        nonce: 0,
        initCode: "0x",
        callData: "0x",
        accountGasLimits: ethers.ZeroHash,
        preVerificationGas: 21000,
        gasFees: ethers.ZeroHash,
        paymasterAndData: paymasterData,
        signature: "0x"
      };
    }

    it("should reset daily limit during validation on new day", async function () {
      // Configure sponsorship
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        100000, // dailyLimit
        50000,  // perTxLimit
        []
      );

      const paymasterAddr = await paymaster.getAddress();
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x00"
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      // First validation uses some of the daily limit
      await paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 50000);

      // Advance to next day
      await time.increase(86401);

      // Should succeed even though we used the limit before
      // because daily counter should reset
      const result = await paymaster.connect(admin).validatePaymasterUserOp.staticCall(
        userOp,
        userOpHash,
        50000
      );
      expect(result[1]).to.equal(0);
    });

    it("should reject when daily limit exceeded", async function () {
      // Configure sponsorship with low daily limit
      await paymaster.connect(operator).configureSponsorship(
        user1.address,
        60000, // dailyLimit
        50000, // perTxLimit
        []
      );

      const paymasterAddr = await paymaster.getAddress();
      const paymasterData = ethers.concat([
        paymasterAddr,
        "0x00"
      ]);

      const userOp = createMockUserOp(user1.address, paymasterData);
      const userOpHash = ethers.keccak256(ethers.toUtf8Bytes("test-op-hash"));

      // First call succeeds
      await paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 50000);

      // Process postOp to update daily usage
      const context = ethers.AbiCoder.defaultAbiCoder().encode(
        ["address", "uint8", "uint256"],
        [user1.address, 0, 50000] // Sponsored mode
      );
      await paymaster.connect(admin).postOp(0, context, 50000, 0);

      // Advance time to avoid rate limiting
      await time.increase(2);

      // Second call should fail (dailyUsed + maxCost > dailyLimit: 50000 + 50000 > 60000)
      await expect(
        paymaster.connect(admin).validatePaymasterUserOp(userOp, userOpHash, 50000)
      ).to.be.revertedWith("Sponsorship invalid");
    });
  });
});
