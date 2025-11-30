// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title MockUSDC
 * @dev Mock USDC token for testing - mimics real USDC behavior
 * Real USDC uses 6 decimals, we follow that standard
 */
contract MockUSDC is ERC20, ERC20Burnable, Ownable {
    uint8 private constant DECIMALS = 6; // USDC uses 6 decimals

    constructor(address initialOwner)
        ERC20("USD Coin", "USDC")
        Ownable(initialOwner)
    {
        // Mint initial supply to owner (1 million USDC)
        _mint(initialOwner, 1_000_000 * 10**DECIMALS);
    }

    /**
     * @dev Returns the number of decimals used (6 for USDC)
     */
    function decimals() public pure override returns (uint8) {
        return DECIMALS;
    }

    /**
     * @dev Mint new tokens - only owner can mint
     * @param to Address to receive the minted tokens
     * @param amount Amount to mint (in smallest unit - 6 decimals)
     */
    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
    }

    /**
     * @dev Mint with amount in dollars (for testing convenience)
     * @param to Address to receive tokens
     * @param amountInDollars Amount in dollars (will be converted to 6 decimals)
     */
    function mintDollars(address to, uint256 amountInDollars) public onlyOwner {
        _mint(to, amountInDollars * 10**DECIMALS);
    }

    /**
     * @dev Faucet for testing - anyone can request 1000 USDC
     */
    function faucet() public {
        _mint(msg.sender, 1000 * 10**DECIMALS);
    }
}
