// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract SphereToken is ERC20, Ownable {
    // Events
    event TokensPurchased(address indexed buyer, uint256 amount, uint256 cost);
    event FiatPurchaseInitiated(address indexed buyer, uint256 amount, string referenceId);
    
    // Token price in ETH (can be updated by owner)
    uint256 public tokenPriceInWei = 0.0001 ether;
    
    // Constructor
    constructor(uint256 initialSupply) ERC20("Sphere", "SPH") {
        _mint(msg.sender, initialSupply * 10**decimals());
    }
    
    // Buy tokens with ETH
    function buyTokens() public payable {
        require(msg.value > 0, "Must send ETH to buy tokens");
        
        uint256 tokenAmount = (msg.value * 10**decimals()) / tokenPriceInWei;
        require(tokenAmount > 0, "Not enough ETH sent");
        
        _mint(msg.sender, tokenAmount);
        
        emit TokensPurchased(msg.sender, tokenAmount, msg.value);
    }
    
    // Update token price (only owner)
    function updateTokenPrice(uint256 newPriceInWei) public onlyOwner {
        tokenPriceInWei = newPriceInWei;
    }
    
    // Mint new tokens (only owner)
    function mintTokens(address to, uint256 amount) public onlyOwner {
        _mint(to, amount * 10**decimals());
    }
    
    // Record fiat purchase (to be fulfilled by backend)
    function recordFiatPurchase(address buyer, uint256 amount, string memory referenceId) public onlyOwner {
        emit FiatPurchaseInitiated(buyer, amount, referenceId);
    }
    
    // Withdraw ETH from contract (only owner)
    function withdraw() public onlyOwner {
        payable(owner()).transfer(address(this).balance);
    }
} 