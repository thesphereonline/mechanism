import React, { useState, useEffect } from 'react';
import { ethers } from 'ethers';
import { useDispatch, useSelector } from 'react-redux';
import { setWallet, clearWallet } from '../redux/walletSlice';
import { Button, Box, Typography, Alert, CircularProgress } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';

const WalletConnect = () => {
  const [error, setError] = useState('');
  const [connecting, setConnecting] = useState(false);
  const wallet = useSelector((state) => state.wallet);
  const dispatch = useDispatch();
  const navigate = useNavigate();

  useEffect(() => {
    // Check if MetaMask is installed
    if (typeof window.ethereum === 'undefined') {
      setError('MetaMask is not installed. Please install MetaMask to use this application.');
    }

    // Listen for account changes
    if (window.ethereum) {
      window.ethereum.on('accountsChanged', handleAccountsChanged);
      window.ethereum.on('chainChanged', () => window.location.reload());
    }

    return () => {
      if (window.ethereum) {
        window.ethereum.removeListener('accountsChanged', handleAccountsChanged);
      }
    };
  }, []);

  const handleAccountsChanged = (accounts) => {
    if (accounts.length === 0) {
      // User disconnected their wallet
      dispatch(clearWallet());
      localStorage.removeItem('token');
    } else if (accounts[0] !== wallet.address) {
      // User switched accounts
      dispatch(clearWallet());
      localStorage.removeItem('token');
    }
  };

  const connectWallet = async () => {
    setConnecting(true);
    setError('');

    try {
      // Request account access
      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
      const address = accounts[0];

      // Get network
      const provider = new ethers.providers.Web3Provider(window.ethereum);
      const network = await provider.getNetwork();
      const chainId = network.chainId;

      // Check if on the correct network (e.g., Rinkeby testnet)
      if (chainId !== 4) {
        try {
          // Try to switch to Rinkeby
          await window.ethereum.request({
            method: 'wallet_switchEthereumChain',
            params: [{ chainId: '0x4' }], // Rinkeby chainId
          });
        } catch (switchError) {
          // If the network is not added, add it
          if (switchError.code === 4902) {
            await window.ethereum.request({
              method: 'wallet_addEthereumChain',
              params: [
                {
                  chainId: '0x4',
                  chainName: 'Rinkeby Test Network',
                  nativeCurrency: {
                    name: 'Ethereum',
                    symbol: 'ETH',
                    decimals: 18,
                  },
                  rpcUrls: ['https://rinkeby.infura.io/v3/your-infura-key'],
                  blockExplorerUrls: ['https://rinkeby.etherscan.io'],
                },
              ],
            });
          } else {
            throw switchError;
          }
        }
      }

      // Get nonce from server
      const response = await api.get(`/auth/nonce?address=${address}`);
      const { nonce, message, exists } = response.data;

      // Sign message
      const signature = await provider.getSigner().signMessage(message);

      if (exists) {
        // Login
        const loginResponse = await api.post('/auth/login', {
          address,
          signature,
          nonce,
        });

        const { token, user } = loginResponse.data;
        localStorage.setItem('token', token);

        dispatch(setWallet({
          address,
          chainId,
          isConnected: true,
          user,
        }));
      } else {
        // Navigate to registration page with pre-filled data
        navigate('/register', {
          state: {
            address,
            signature,
            nonce,
          },
        });
      }
    } catch (err) {
      console.error('Error connecting wallet:', err);
      setError(err.message || 'Failed to connect wallet');
    } finally {
      setConnecting(false);
    }
  };

  const disconnectWallet = () => {
    dispatch(clearWallet());
    localStorage.removeItem('token');
  };

  return (
    <Box>
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      
      {wallet.isConnected ? (
        <Box>
          <Typography variant="body2" sx={{ mb: 1 }}>
            Connected: {wallet.address.substring(0, 6)}...{wallet.address.substring(38)}
          </Typography>
          <Button 
            variant="outlined" 
            color="secondary" 
            onClick={disconnectWallet}
            size="small"
          >
            Disconnect
          </Button>
        </Box>
      ) : (
        <Button 
          variant="contained" 
          color="primary" 
          onClick={connectWallet}
          disabled={connecting || typeof window.ethereum === 'undefined'}
          startIcon={connecting && <CircularProgress size={20} color="inherit" />}
        >
          {connecting ? 'Connecting...' : 'Connect Wallet'}
        </Button>
      )}
    </Box>
  );
};

export default WalletConnect; 