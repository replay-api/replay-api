const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  const [deployer] = await hre.ethers.getSigners();

  console.log("Deploying contracts with account:", deployer.address);
  console.log("Account balance:", (await hre.ethers.provider.getBalance(deployer.address)).toString());

  // Deploy MockUSDC
  console.log("\n Deploying MockUSDC...");
  const MockUSDC = await hre.ethers.getContractFactory("MockUSDC");
  const usdc = await MockUSDC.deploy(deployer.address);
  await usdc.waitForDeployment();
  const usdcAddress = await usdc.getAddress();
  console.log("✓ MockUSDC deployed to:", usdcAddress);

  // Deploy MockUSDT
  console.log("\nDeploying MockUSDT...");
  const MockUSDT = await hre.ethers.getContractFactory("MockUSDT");
  const usdt = await MockUSDT.deploy(deployer.address);
  await usdt.waitForDeployment();
  const usdtAddress = await usdt.getAddress();
  console.log("✓ MockUSDT deployed to:", usdtAddress);

  // Deploy GameNFT
  console.log("\nDeploying GameNFT...");
  const GameNFT = await hre.ethers.getContractFactory("GameNFT");
  const gameNFT = await GameNFT.deploy(deployer.address);
  await gameNFT.waitForDeployment();
  const gameNFTAddress = await gameNFT.getAddress();
  console.log("✓ GameNFT deployed to:", gameNFTAddress);

  // Save deployment addresses to JSON file
  const deploymentInfo = {
    network: hre.network.name,
    chainId: (await hre.ethers.provider.getNetwork()).chainId.toString(),
    deployer: deployer.address,
    contracts: {
      MockUSDC: usdcAddress,
      MockUSDT: usdtAddress,
      GameNFT: gameNFTAddress
    },
    timestamp: new Date().toISOString()
  };

  const deploymentsDir = path.join(__dirname, "../deployments");
  if (!fs.existsSync(deploymentsDir)) {
    fs.mkdirSync(deploymentsDir, { recursive: true });
  }

  const deploymentPath = path.join(deploymentsDir, `${hre.network.name}.json`);
  fs.writeFileSync(deploymentPath, JSON.stringify(deploymentInfo, null, 2));

  console.log("\n✓ Deployment info saved to:", deploymentPath);
  console.log("\nDeployment Summary:");
  console.log("==================");
  console.log("Network:", hre.network.name);
  console.log("Chain ID:", deploymentInfo.chainId);
  console.log("MockUSDC:", usdcAddress);
  console.log("MockUSDT:", usdtAddress);
  console.log("GameNFT:", gameNFTAddress);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
