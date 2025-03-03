import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import {
  Container,
  Grid,
  Card,
  CardMedia,
  Typography,
  Button,
  Chip,
  Box,
  Divider,
  Paper,
  CircularProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField
} from '@mui/material';
import api from '../services/api';

const NFTDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const wallet = useSelector((state) => state.wallet);
  
  const [nft, setNft] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState('');
  const [actionSuccess, setActionSuccess] = useState('');
  
  // Dialog states
  const [listDialogOpen, setListDialogOpen] = useState(false);
  const [price, setPrice] = useState('');
  
  useEffect(() => {
    const fetchNFT = async () => {
      try {
        const response = await api.get(`/nfts/${id}`);
        setNft(response.data);
      } catch (err) {
        console.error('Error fetching NFT:', err);
        setError('Failed to load NFT details');
      } finally {
        setLoading(false);
      }
    };
    
    fetchNFT();
  }, [id]);
  
  const isOwner = wallet.isConnected && nft?.owner_id === wallet.user?.id;
  const canMint = isOwner && nft?.status === 'uploaded';
  const canList = isOwner && nft?.status === 'minted';
  const canBuy = !isOwner && nft?.status === 'listed' && wallet.isConnected;
  
  const handleMint = async () => {
    setActionLoading(true);
    setActionError('');
    setActionSuccess('');
    
    try {
      await api.post(`/nfts/mint`, { nft_id: nft.id });
      setActionSuccess('NFT minted successfully!');
      // Refresh NFT data
      const response = await api.get(`/nfts/${id}`);
      setNft(response.data);
    } catch (err) {
      console.error('Error minting NFT:', err);
      setActionError(err.response?.data?.error || 'Failed to mint NFT');
    } finally {
      setActionLoading(false);
    }
  };
  
  const handleListDialogOpen = () => {
    setListDialogOpen(true);
  };
  
  const handleListDialogClose = () => {
    setListDialogOpen(false);
    setPrice('');
  };
  
  const handleList = async () => {
    if (!price || isNaN(price) || parseFloat(price) <= 0) {
      setActionError('Please enter a valid price');
      return;
    }
    
    setActionLoading(true);
    setActionError('');
    setActionSuccess('');
    handleListDialogClose();
    
    try {
      await api.post(`/nfts/list`, { 
        nft_id: nft.id,
        price: parseFloat(price)
      });
      setActionSuccess('NFT listed for sale successfully!');
      // Refresh NFT data
      const response = await api.get(`/nfts/${id}`);
      setNft(response.data);
    } catch (err) {
      console.error('Error listing NFT:', err);
      setActionError(err.response?.data?.error || 'Failed to list NFT');
    } finally {
      setActionLoading(false);
    }
  };
  
  const handleBuy = async () => {
    setActionLoading(true);
    setActionError('');
    setActionSuccess('');
    
    try {
      await api.post(`/nfts/buy`, { nft_id: nft.id });
      setActionSuccess('NFT purchased successfully!');
      // Refresh NFT data
      const response = await api.get(`/nfts/${id}`);
      setNft(response.data);
    } catch (err) {
      console.error('Error buying NFT:', err);
      setActionError(err.response?.data?.error || 'Failed to buy NFT');
    } finally {
      setActionLoading(false);
    }
  };
  
  if (loading) {
    return (
      <Container sx={{ py: 4, textAlign: 'center' }}>
        <CircularProgress />
        <Typography sx={{ mt: 2 }}>Loading NFT details...</Typography>
      </Container>
    );
  }
  
  if (error) {
    return (
      <Container sx={{ py: 4 }}>
        <Alert severity="error">{error}</Alert>
        <Button 
          variant="outlined" 
          sx={{ mt: 2 }}
          onClick={() => navigate('/marketplace')}
        >
          Back to Marketplace
        </Button>
      </Container>
    );
  }
  
  if (!nft) {
    return (
      <Container sx={{ py: 4 }}>
        <Alert severity="warning">NFT not found</Alert>
        <Button 
          variant="outlined" 
          sx={{ mt: 2 }}
          onClick={() => navigate('/marketplace')}
        >
          Back to Marketplace
        </Button>
      </Container>
    );
  }
  
  return (
    <Container sx={{ py: 4 }}>
      {actionError && <Alert severity="error" sx={{ mb: 3 }}>{actionError}</Alert>}
      {actionSuccess && <Alert severity="success" sx={{ mb: 3 }}>{actionSuccess}</Alert>}
      
      <Grid container spacing={4}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardMedia
              component="img"
              image={nft.image_url}
              alt={nft.title}
              sx={{ 
                height: { xs: 300, md: 400 },
                objectFit: 'contain',
                bgcolor: '#f5f5f5'
              }}
            />
          </Card>
        </Grid>
        
        <Grid item xs={12} md={6}>
          <Typography variant="h4" component="h1" gutterBottom>
            {nft.title}
          </Typography>
          
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
            <Chip 
              label={nft.status.charAt(0).toUpperCase() + nft.status.slice(1)}
              color={
                nft.status === 'listed' ? 'success' :
                nft.status === 'minted' ? 'primary' : 'default'
              }
              sx={{ mr: 1 }}
            />
            <Chip label={nft.category} variant="outlined" />
          </Box>
          
          {nft.status === 'listed' && (
            <Typography variant="h5" color="primary" sx={{ mb: 2 }}>
              {nft.price} SPH
            </Typography>
          )}
          
          <Typography variant="body1" paragraph>
            {nft.description}
          </Typography>
          
          <Divider sx={{ my: 2 }} />
          
          <Typography variant="subtitle1" gutterBottom>
            Creator: {nft.creator?.username || nft.creator_id}
          </Typography>
          
          <Typography variant="subtitle1" gutterBottom>
            Owner: {nft.owner?.username || nft.owner_id}
          </Typography>
          
          {nft.token_id && (
            <Typography variant="subtitle1" gutterBottom>
              Token ID: {nft.token_id}
            </Typography>
          )}
          
          <Box sx={{ mt: 3 }}>
            {canMint && (
              <Button
                variant="contained"
                color="primary"
                fullWidth
                onClick={handleMint}
                disabled={actionLoading}
                startIcon={actionLoading && <CircularProgress size={20} color="inherit" />}
              >
                {actionLoading ? 'Processing...' : 'Mint NFT'}
              </Button>
            )}
            
            {canList && (
              <Button
                variant="contained"
                color="success"
                fullWidth
                onClick={handleListDialogOpen}
                disabled={actionLoading}
                startIcon={actionLoading && <CircularProgress size={20} color="inherit" />}
              >
                {actionLoading ? 'Processing...' : 'List for Sale'}
              </Button>
            )}
            
            {canBuy && (
              <Button
                variant="contained"
                color="primary"
                fullWidth
                onClick={handleBuy}
                disabled={actionLoading}
                startIcon={actionLoading && <CircularProgress size={20} color="inherit" />}
              >
                {actionLoading ? 'Processing...' : `Buy for ${nft.price} SPH`}
              </Button>
            )}
            
            {!wallet.isConnected && (
              <Alert severity="info" sx={{ mt: 2 }}>
                Connect your wallet to interact with this NFT
              </Alert>
            )}
          </Box>
        </Grid>
      </Grid>
      
      {/* List NFT Dialog */}
      <Dialog open={listDialogOpen} onClose={handleListDialogClose}>
        <DialogTitle>List NFT for Sale</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Price (SPH)"
            type="number"
            fullWidth
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            InputProps={{ inputProps: { min: 0, step: 0.01 } }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleListDialogClose}>Cancel</Button>
          <Button onClick={handleList} color="primary">List</Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default NFTDetail; 