import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  address: '',
  chainId: null,
  isConnected: false,
  user: null
};

export const walletSlice = createSlice({
  name: 'wallet',
  initialState,
  reducers: {
    setWallet: (state, action) => {
      state.address = action.payload.address;
      state.chainId = action.payload.chainId;
      state.isConnected = action.payload.isConnected;
      state.user = action.payload.user;
    },
    clearWallet: (state) => {
      state.address = '';
      state.chainId = null;
      state.isConnected = false;
      state.user = null;
    }
  }
});

export const { setWallet, clearWallet } = walletSlice.actions;

export default walletSlice.reducer; 