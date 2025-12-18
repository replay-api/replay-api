// SPDX-License-Identifier: MIT
/**
 * @title GameNFT Tests
 * @notice Tests for in-game NFT contract (skins, weapons, achievements)
 */

const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("GameNFT", function () {
  let admin, user1, user2, attacker;
  let nft;

  const COMMON = 0;
  const RARE = 1;
  const EPIC = 2;
  const LEGENDARY = 3;

  beforeEach(async function () {
    [admin, user1, user2, attacker] = await ethers.getSigners();

    const GameNFT = await ethers.getContractFactory("GameNFT");
    nft = await GameNFT.deploy(admin.address);
  });

  describe("Deployment", function () {
    it("should set correct name and symbol", async function () {
      expect(await nft.name()).to.equal("LeetGaming NFT");
      expect(await nft.symbol()).to.equal("LGNFT");
    });

    it("should set admin as owner", async function () {
      expect(await nft.owner()).to.equal(admin.address);
    });
  });

  describe("Minting", function () {
    it("should mint NFT with correct rarity", async function () {
      const uri = "ipfs://QmTest123";
      const tx = await nft.connect(admin).safeMint(user1.address, RARE, uri);

      expect(await nft.ownerOf(1)).to.equal(user1.address);
      expect(await nft.itemRarity(1)).to.equal(RARE);
      expect(await nft.tokenURI(1)).to.equal(uri);
    });

    it("should emit ItemMinted event", async function () {
      const uri = "ipfs://QmTest456";
      await expect(nft.connect(admin).safeMint(user1.address, EPIC, uri))
        .to.emit(nft, "ItemMinted")
        .withArgs(user1.address, 1, EPIC, uri);
    });

    it("should reject minting from non-owner", async function () {
      await expect(
        nft.connect(attacker).safeMint(user1.address, COMMON, "uri")
      ).to.be.reverted;
    });

    it("should reject invalid rarity level", async function () {
      await expect(
        nft.connect(admin).safeMint(user1.address, 4, "uri")
      ).to.be.revertedWith("Invalid rarity level");
    });

    it("should increment token IDs", async function () {
      await nft.connect(admin).safeMint(user1.address, COMMON, "uri1");
      await nft.connect(admin).safeMint(user2.address, RARE, "uri2");

      expect(await nft.ownerOf(1)).to.equal(user1.address);
      expect(await nft.ownerOf(2)).to.equal(user2.address);
    });
  });

  describe("Batch Minting", function () {
    it("should batch mint multiple NFTs", async function () {
      const count = 5;
      await nft.connect(admin).batchMint(user1.address, count, RARE);

      for (let i = 1; i <= count; i++) {
        expect(await nft.ownerOf(i)).to.equal(user1.address);
        expect(await nft.itemRarity(i)).to.equal(RARE);
      }
    });

    it("should reject batch mint with zero count", async function () {
      await expect(
        nft.connect(admin).batchMint(user1.address, 0, COMMON)
      ).to.be.revertedWith("Invalid count");
    });

    it("should reject batch mint exceeding limit", async function () {
      await expect(
        nft.connect(admin).batchMint(user1.address, 101, COMMON)
      ).to.be.revertedWith("Invalid count");
    });

    it("should reject batch mint from non-owner", async function () {
      await expect(
        nft.connect(attacker).batchMint(user1.address, 5, COMMON)
      ).to.be.reverted;
    });
  });

  describe("Faucet", function () {
    it("should allow anyone to mint via faucet", async function () {
      await nft.connect(user1).faucet();

      expect(await nft.ownerOf(1)).to.equal(user1.address);
      expect(await nft.itemRarity(1)).to.equal(COMMON);
    });

    it("should allow multiple faucet calls from same user", async function () {
      await nft.connect(user1).faucet();
      await nft.connect(user1).faucet();

      expect(await nft.balanceOf(user1.address)).to.equal(2);
    });
  });

  describe("Burning", function () {
    it("should allow owner to burn their NFT", async function () {
      await nft.connect(admin).safeMint(user1.address, LEGENDARY, "uri");

      await nft.connect(user1).burn(1);

      await expect(nft.ownerOf(1)).to.be.reverted;
    });
  });

  describe("Token URI", function () {
    it("should return correct token URI", async function () {
      const uri = "ipfs://QmTestMetadata123";
      await nft.connect(admin).safeMint(user1.address, RARE, uri);

      expect(await nft.tokenURI(1)).to.equal(uri);
    });
  });

  describe("Interface Support", function () {
    it("should support ERC721 interface", async function () {
      expect(await nft.supportsInterface("0x80ac58cd")).to.be.true;
    });
  });
});
