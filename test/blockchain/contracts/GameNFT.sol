// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Burnable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title GameNFT
 * @dev NFT contract for in-game items (skins, weapons, achievements)
 * Supports metadata URIs for item attributes
 */
contract GameNFT is ERC721, ERC721URIStorage, ERC721Burnable, Ownable {
    uint256 private _nextTokenId;

    // Mapping from token ID to item rarity (Common, Rare, Epic, Legendary)
    mapping(uint256 => uint8) public itemRarity;

    event ItemMinted(address indexed to, uint256 indexed tokenId, uint8 rarity, string tokenURI);

    constructor(address initialOwner)
        ERC721("LeetGaming NFT", "LGNFT")
        Ownable(initialOwner)
    {
        _nextTokenId = 1; // Start token IDs at 1
    }

    /**
     * @dev Mint a new NFT with metadata
     * @param to Address to receive the NFT
     * @param rarity Rarity level (0=Common, 1=Rare, 2=Epic, 3=Legendary)
     * @param uri Metadata URI (IPFS or HTTP)
     */
    function safeMint(address to, uint8 rarity, string memory uri) public onlyOwner returns (uint256) {
        require(rarity <= 3, "Invalid rarity level");

        uint256 tokenId = _nextTokenId++;
        _safeMint(to, tokenId);
        _setTokenURI(tokenId, uri);
        itemRarity[tokenId] = rarity;

        emit ItemMinted(to, tokenId, rarity, uri);

        return tokenId;
    }

    /**
     * @dev Batch mint multiple NFTs
     * @param to Address to receive the NFTs
     * @param count Number of NFTs to mint
     * @param rarity Rarity level for all NFTs
     */
    function batchMint(address to, uint256 count, uint8 rarity) public onlyOwner returns (uint256[] memory) {
        require(count > 0 && count <= 100, "Invalid count");
        require(rarity <= 3, "Invalid rarity level");

        uint256[] memory tokenIds = new uint256[](count);

        for (uint256 i = 0; i < count; i++) {
            uint256 tokenId = _nextTokenId++;
            _safeMint(to, tokenId);
            itemRarity[tokenId] = rarity;
            tokenIds[i] = tokenId;

            emit ItemMinted(to, tokenId, rarity, "");
        }

        return tokenIds;
    }

    /**
     * @dev Faucet for testing - anyone can mint 1 common NFT
     */
    function faucet() public returns (uint256) {
        uint256 tokenId = _nextTokenId++;
        _safeMint(msg.sender, tokenId);
        itemRarity[tokenId] = 0; // Common rarity
        emit ItemMinted(msg.sender, tokenId, 0, "");
        return tokenId;
    }

    /**
     * @dev Get rarity name string
     */
    function getRarityName(uint256 tokenId) public view returns (string memory) {
        require(ownerOf(tokenId) != address(0), "Token does not exist");

        uint8 rarity = itemRarity[tokenId];
        if (rarity == 0) return "Common";
        if (rarity == 1) return "Rare";
        if (rarity == 2) return "Epic";
        if (rarity == 3) return "Legendary";
        return "Unknown";
    }

    // Override required by Solidity for multiple inheritance
    function tokenURI(uint256 tokenId)
        public
        view
        override(ERC721, ERC721URIStorage)
        returns (string memory)
    {
        return super.tokenURI(tokenId);
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(ERC721, ERC721URIStorage)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }
}
