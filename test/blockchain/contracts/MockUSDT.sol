// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title MockUSDT
 * @dev Mock Tether token for testing - mimics real USDT behavior
 * Real USDT uses 6 decimals, we follow that standard
 */
contract MockUSDT is ERC20, ERC20Burnable, Ownable {
    uint8 private constant DECIMALS = 6; // USDT uses 6 decimals

    constructor(address initialOwner)
        ERC20("Tether USD", "USDT")
        Ownable(initialOwner)
    {
        // Mint initial supply to owner (1 million USDT)
        _mint(initialOwner, 1_000_000 * 10**DECIMALS);
    }

    /**
     * @dev Returns the number of decimals used (6 for USDT)
     */
    function decimals() public pure override returns (uint8) {
        return DECIMALS;
    }

    /**
     * @dev Mint new tokens - only owner can mint
     */
    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
    }

    /**
     * @dev Mint with amount in dollars (for testing convenience)
     */
    function mintDollars(address to, uint256 amountInDollars) public onlyOwner {
        _mint(to, amountInDollars * 10**DECIMALS);
    }

    /**
     * @dev Faucet for testing - anyone can request 1000 USDT
     */
    function faucet() public {
        _mint(msg.sender, 1000 * 10**DECIMALS);
    }
}
