import React, { useState } from 'react';
import {
  Box,
  Button,
  TextField,
  Typography,
  Paper,
  Grid,
  MenuItem,
  FormControl,
  InputLabel,
  Select,
  Alert,
  CircularProgress,
  Card,
  CardMedia
} from '@mui/material';
import { useDropzone } from 'react-dropzone';
import { useSelector } from 'react-redux';
import api from '../services/api';
import { useNavigate } from 'react-router-dom';

const categories = [
  'Art',
  'Music',
  'Photography',
  'Sports',
  'Collectibles',
  'Virtual Worlds',
  'Trading Cards',
  'Utility',
  'Domain Names',
  'Other'
];

const UploadNFT = () => {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState('');
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const wallet = useSelector((state) => state.wallet);
  const navigate = useNavigate();

  const { getRootProps, getInputProps } = useDropzone({
    accept: {
      'image/*': ['.jpeg', '.jpg', '.png', '.gif', '.webp']
    },
    maxSize: 10485760, // 10MB
    onDrop: (acceptedFiles) => {
      if (acceptedFiles.length > 0) {
        const selectedFile = acceptedFiles[0];
        setFile(selectedFile);
        setPreview(URL.createObjectURL(selectedFile));
      }
    },
    onDropRejected: (rejectedFiles) => {
      if (rejectedFiles[0].errors[0].code === 'file-too-large') {
        setError('File is too large. Maximum size is 10MB.');
      } else {
        setError('Invalid file type. Please upload an image.');
      }
    }
  });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');

    if (!wallet.isConnected) {
      setError('Please connect your wallet first');
      setLoading(false);
      return;
    }

    if (!title || !description || !category || !file) {
      setError('Please fill in all fields and upload an image');
      setLoading(false);
      return;
    }

    try {
      // Create form data
      const formData = new FormData();
      formData.append('title', title);
      formData.append('description', description);
      formData.append('category', category);
      formData.append('file', file);

      // Upload NFT
      const response = await api.post('/nfts/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      setSuccess('NFT uploaded successfully! You can now mint it.');
      setTimeout(() => {
        navigate(`/nft/${response.data.id}`);
      }, 2000);
    } catch (err) {
      console.error('Error uploading NFT:', err);
      setError(err.response?.data?.error || 'Failed to upload NFT');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Paper elevation={3} sx={{ p: 4, maxWidth: 800, mx: 'auto', mt: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Upload New NFT
      </Typography>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}

      <form onSubmit={handleSubmit}>
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <TextField
              label="Title"
              fullWidth
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
              disabled={loading}
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <FormControl fullWidth required>
              <InputLabel>Category</InputLabel>
              <Select
                value={category}
                label="Category"
                onChange={(e) => setCategory(e.target.value)}
                disabled={loading}
              >
                {categories.map((cat) => (
                  <MenuItem key={cat} value={cat}>{cat}</MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12}>
            <TextField
              label="Description"
              fullWidth
              multiline
              rows={4}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              required
              disabled={loading}
            />
          </Grid>
          <Grid item xs={12}>
            <Box
              {...getRootProps()}
              sx={{
                border: '2px dashed #ccc',
                borderRadius: 2,
                p: 3,
                textAlign: 'center',
                cursor: 'pointer',
                mb: 2,
                '&:hover': {
                  borderColor: 'primary.main'
                }
              }}
            >
              <input {...getInputProps()} />
              <Typography>
                Drag and drop an image here, or click to select a file
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Supported formats: JPG, PNG, GIF, WEBP (Max 10MB)
              </Typography>
            </Box>
            {preview && (
              <Card sx={{ maxWidth: 300, mx: 'auto', mb: 2 }}>
                <CardMedia
                  component="img"
                  height="200"
                  image={preview}
                  alt="Preview"
                  sx={{ objectFit: 'contain' }}
                />
              </Card>
            )}
          </Grid>
          <Grid item xs={12}>
            <Button
              type="submit"
              variant="contained"
              color="primary"
              size="large"
              fullWidth
              disabled={loading || !wallet.isConnected}
              startIcon={loading && <CircularProgress size={20} color="inherit" />}
            >
              {loading ? 'Uploading...' : 'Upload NFT'}
            </Button>
          </Grid>
        </Grid>
      </form>
    </Paper>
  );
};

export default UploadNFT; 