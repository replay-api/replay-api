const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  const [deployer] = await hre.ethers.getSigners();

  console.log("Deploying contracts with account:", deployer.address);
  console.log("Account balance:", (await hre.ethers.provider.getBalance(deployer.address)).toString());

  // Deploy MockUSDC
  console.log("\n1. Deploying MockUSDC...");
  const MockUSDC = await hre.ethers.getContractFactory("MockUSDC");
  const usdc = await MockUSDC.deploy(deployer.address);
  await usdc.waitForDeployment();
  const usdcAddress = await usdc.getAddress();
  console.log("MockUSDC deployed to:", usdcAddress);

  // Deploy MockUSDT
  console.log("\n2. Deploying MockUSDT...");
  const MockUSDT = await hre.ethers.getContractFactory("MockUSDT");
  const usdt = await MockUSDT.deploy(deployer.address);
  await usdt.waitForDeployment();
  const usdtAddress = await usdt.getAddress();
  console.log("MockUSDT deployed to:", usdtAddress);

  // Deploy GameNFT
  console.log("\n3. Deploying GameNFT...");
  const GameNFT = await hre.ethers.getContractFactory("GameNFT");
  const gameNFT = await GameNFT.deploy(deployer.address);
  await gameNFT.waitForDeployment();
  const gameNFTAddress = await gameNFT.getAddress();
  console.log("GameNFT deployed to:", gameNFTAddress);

  // Deploy LeetVault (treasury = deployer for testing)
  console.log("\n4. Deploying LeetVault...");
  const LeetVault = await hre.ethers.getContractFactory("LeetVault");
  const vault = await LeetVault.deploy(deployer.address);
  await vault.waitForDeployment();
  const vaultAddress = await vault.getAddress();
  console.log("LeetVault deployed to:", vaultAddress);

  // Deploy LeetLedger
  console.log("\n5. Deploying LeetLedger...");
  const LeetLedger = await hre.ethers.getContractFactory("LeetLedger");
  const ledger = await LeetLedger.deploy();
  await ledger.waitForDeployment();
  const ledgerAddress = await ledger.getAddress();
  console.log("LeetLedger deployed to:", ledgerAddress);

  // Deploy LeetSmartWallet implementation
  console.log("\n6. Deploying LeetSmartWallet implementation...");
  const LeetSmartWallet = await hre.ethers.getContractFactory("LeetSmartWallet");
  const walletImpl = await LeetSmartWallet.deploy();
  await walletImpl.waitForDeployment();
  const walletImplAddress = await walletImpl.getAddress();
  console.log("LeetSmartWallet implementation deployed to:", walletImplAddress);

  // Deploy LeetPaymaster
  console.log("\n7. Deploying LeetPaymaster...");
  const LeetPaymaster = await hre.ethers.getContractFactory("LeetPaymaster");
  // entryPoint, treasury, verifyingSigner (using deployer for all in test)
  const paymaster = await LeetPaymaster.deploy(deployer.address, deployer.address, deployer.address);
  await paymaster.waitForDeployment();
  const paymasterAddress = await paymaster.getAddress();
  console.log("LeetPaymaster deployed to:", paymasterAddress);

  // Configure LeetVault
  console.log("\n8. Configuring LeetVault...");
  await vault.addSupportedToken(usdcAddress);
  console.log("   Added USDC as supported token");
  await vault.addSupportedToken(usdtAddress);
  console.log("   Added USDT as supported token");

  // Configure LeetLedger - grant recorder role to vault
  console.log("\n9. Configuring LeetLedger...");
  const RECORDER_ROLE = await ledger.RECORDER_ROLE();
  await ledger.grantRole(RECORDER_ROLE, vaultAddress);
  console.log("   Granted RECORDER_ROLE to LeetVault");

  // Configure LeetPaymaster
  console.log("\n10. Configuring LeetPaymaster...");
  await paymaster.setAcceptedToken(usdcAddress, 1000000); // 1 USDC per 1M gas
  console.log("    Added USDC as accepted token for gas payments");

  // Save deployment addresses to JSON file
  const deploymentInfo = {
    network: hre.network.name,
    chainId: (await hre.ethers.provider.getNetwork()).chainId.toString(),
    deployer: deployer.address,
    contracts: {
      MockUSDC: usdcAddress,
      MockUSDT: usdtAddress,
      GameNFT: gameNFTAddress,
      LeetVault: vaultAddress,
      LeetLedger: ledgerAddress,
      LeetSmartWalletImpl: walletImplAddress,
      LeetPaymaster: paymasterAddress
    },
    roles: {
      OPERATOR_ROLE: await vault.OPERATOR_ROLE(),
      ORACLE_ROLE: await vault.ORACLE_ROLE(),
      RECORDER_ROLE: RECORDER_ROLE
    },
    timestamp: new Date().toISOString()
  };

  const deploymentsDir = path.join(__dirname, "../deployments");
  if (!fs.existsSync(deploymentsDir)) {
    fs.mkdirSync(deploymentsDir, { recursive: true });
  }

  const deploymentPath = path.join(deploymentsDir, `${hre.network.name}.json`);
  fs.writeFileSync(deploymentPath, JSON.stringify(deploymentInfo, null, 2));

  console.log("\n========================================");
  console.log("DEPLOYMENT COMPLETE");
  console.log("========================================");
  console.log("\nContract Addresses:");
  console.log("-------------------");
  console.log("MockUSDC:           ", usdcAddress);
  console.log("MockUSDT:           ", usdtAddress);
  console.log("GameNFT:            ", gameNFTAddress);
  console.log("LeetVault:          ", vaultAddress);
  console.log("LeetLedger:         ", ledgerAddress);
  console.log("LeetSmartWalletImpl:", walletImplAddress);
  console.log("LeetPaymaster:      ", paymasterAddress);
  console.log("\nDeployer:", deployer.address);
  console.log("Network:", hre.network.name);
  console.log("Chain ID:", deploymentInfo.chainId);
  console.log("\nDeployment info saved to:", deploymentPath);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
