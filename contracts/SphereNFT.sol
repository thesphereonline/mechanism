// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Counters.sol";
import "./SphereToken.sol";

contract SphereNFT is ERC721URIStorage, Ownable {
    using Counters for Counters.Counter;
    
    // Events
    event NFTListed(uint256 indexed tokenId, address seller, uint256 price);
    event NFTSold(uint256 indexed tokenId, address seller, address buyer, uint256 price);
    
    // Token ID counter
    Counters.Counter private _tokenIds;
    
    // Reference to Sphere token
    SphereToken public sphereToken;
    
    // NFT marketplace data
    struct NFTListing {
        uint256 tokenId;
        address payable seller;
        uint256 price; // Price in Sphere tokens
        bool isActive;
    }
    
    // Mapping from token ID to listing
    mapping(uint256 => NFTListing) public listings;
    
    // Platform fee percentage (2.5%)
    uint256 public platformFeePercent = 250;
    
    // Constructor
    constructor(address sphereTokenAddress) ERC721("Sphere NFT", "SPHNFT") {
        sphereToken = SphereToken(sphereTokenAddress);
    }
    
    // Mint a new NFT
    function mintNFT(address recipient, string memory tokenURI) public onlyOwner returns (uint256) {
        _tokenIds.increment();
        uint256 newTokenId = _tokenIds.current();
        
        _mint(recipient, newTokenId);
        _setTokenURI(newTokenId, tokenURI);
        
        return newTokenId;
    }
    
    // List an NFT for sale
    function listNFT(uint256 tokenId, uint256 price) public {
        require(ownerOf(tokenId) == msg.sender, "Only the owner can list the NFT");
        require(price > 0, "Price must be greater than zero");
        
        // Transfer NFT to contract
        _transfer(msg.sender, address(this), tokenId);
        
        // Create listing
        listings[tokenId] = NFTListing({
            tokenId: tokenId,
            seller: payable(msg.sender),
            price: price,
            isActive: true
        });
        
        emit NFTListed(tokenId, msg.sender, price);
    }
    
    // Buy an NFT with Sphere tokens
    function buyNFT(uint256 tokenId) public {
        NFTListing storage listing = listings[tokenId];
        
        require(listing.isActive, "NFT not listed for sale");
        require(msg.sender != listing.seller, "Seller cannot buy their own NFT");
        
        uint256 price = listing.price;
        address seller = listing.seller;
        
        // Calculate platform fee
        uint256 platformFee = (price * platformFeePercent) / 10000;
        uint256 sellerAmount = price - platformFee;
        
        // Transfer Sphere tokens from buyer to seller and platform
        require(sphereToken.transferFrom(msg.sender, seller, sellerAmount), "Token transfer to seller failed");
        require(sphereToken.transferFrom(msg.sender, owner(), platformFee), "Token transfer to platform failed");
        
        // Transfer NFT to buyer
        _transfer(address(this), msg.sender, tokenId);
        
        // Update listing
        listing.isActive = false;
        
        emit NFTSold(tokenId, seller, msg.sender, price);
    }
    
    // Cancel NFT listing
    function cancelListing(uint256 tokenId) public {
        NFTListing storage listing = listings[tokenId];
        
        require(listing.seller == msg.sender, "Only the seller can cancel the listing");
        require(listing.isActive, "Listing is not active");
        
        // Transfer NFT back to seller
        _transfer(address(this), msg.sender, tokenId);
        
        // Update listing
        listing.isActive = false;
    }
    
    // Update platform fee (only owner)
    function updatePlatformFee(uint256 newFeePercent) public onlyOwner {
        require(newFeePercent <= 1000, "Fee cannot exceed 10%");
        platformFeePercent = newFeePercent;
    }
} 