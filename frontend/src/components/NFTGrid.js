import React from 'react';
import { 
  Grid, 
  Card, 
  CardMedia, 
  CardContent, 
  Typography, 
  CardActions, 
  Button,
  Chip,
  Box,
  Skeleton
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { ethers } from 'ethers';

const NFTCard = ({ nft, onBuy }) => {
  const navigate = useNavigate();
  const wallet = useSelector((state) => state.wallet);
  
  const isOwner = wallet.isConnected && nft.owner_id === wallet.user.id;
  const isForSale = nft.status === 'listed';
  
  const handleViewDetails = () => {
    navigate(`/nft/${nft.id}`);
  };
  
  const getStatusColor = (status) => {
    switch (status) {
      case 'uploaded':
        return 'default';
      case 'minted':
        return 'primary';
      case 'listed':
        return 'success';
      case 'owned':
        return 'secondary';
      default:
        return 'default';
    }
  };
  
  return (
    <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <CardMedia
        component="img"
        height="200"
        image={nft.image_url}
        alt={nft.title}
        sx={{ objectFit: 'cover' }}
      />
      <CardContent sx={{ flexGrow: 1 }}>
        <Typography gutterBottom variant="h6" component="div" noWrap>
          {nft.title}
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          display: '-webkit-box',
          WebkitLineClamp: 2,
          WebkitBoxOrient: 'vertical',
        }}>
          {nft.description}
        </Typography>
        
        <Box sx={{ mt: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          {isForSale && (
            <Typography variant="h6" color="primary">
              {nft.price} SPH
            </Typography>
          )}
          <Chip 
            label={nft.status.charAt(0).toUpperCase() + nft.status.slice(1)} 
            color={getStatusColor(nft.status)}
            size="small"
          />
        </Box>
      </CardContent>
      <CardActions>
        <Button size="small" onClick={handleViewDetails}>View Details</Button>
        {isForSale && !isOwner && wallet.isConnected && (
          <Button 
            size="small" 
            variant="contained" 
            color="primary"
            onClick={() => onBuy(nft.id)}
          >
            Buy Now
          </Button>
        )}
      </CardActions>
    </Card>
  );
};

const NFTGrid = ({ nfts, loading, onBuy }) => {
  // Create skeleton cards for loading state
  const skeletonCards = Array(8).fill().map((_, index) => (
    <Grid item xs={12} sm={6} md={4} lg={3} key={`skeleton-${index}`}>
      <Card sx={{ height: '100%' }}>
        <Skeleton variant="rectangular" height={200} />
        <CardContent>
          <Skeleton variant="text" />
          <Skeleton variant="text" />
          <Skeleton variant="text" width="60%" />
          <Box sx={{ mt: 2, display: 'flex', justifyContent: 'space-between' }}>
            <Skeleton variant="text" width="30%" />
            <Skeleton variant="circular" width={40} height={40} />
          </Box>
        </CardContent>
        <CardActions>
          <Skeleton variant="rectangular" width={80} height={30} />
          <Skeleton variant="rectangular" width={80} height={30} />
        </CardActions>
      </Card>
    </Grid>
  ));

  return (
    <Grid container spacing={3}>
      {loading ? (
        skeletonCards
      ) : nfts.length > 0 ? (
        nfts.map((nft) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={nft.id}>
            <NFTCard nft={nft} onBuy={onBuy} />
          </Grid>
        ))
      ) : (
        <Grid item xs={12}>
          <Typography variant="h6" align="center" color="text.secondary" sx={{ py: 5 }}>
            No NFTs found
          </Typography>
        </Grid>
      )}
    </Grid>
  );
};

export default NFTGrid; 