import api from './client';
import type { Cart, CartItem } from '../types';

export const cartApi = {
  get: async (): Promise<Cart> => {
    const response = await api.get('/cart');
    return response.data;
  },

  addItem: async (item: Omit<CartItem, 'name' | 'price' | 'imageUrl'>): Promise<Cart> => {
    const response = await api.post('/cart/items', item);
    return response.data;
  },

  updateItem: async (productId: string, quantity: number): Promise<Cart> => {
    const response = await api.put(`/cart/items/${productId}`, { quantity });
    return response.data;
  },

  removeItem: async (productId: string): Promise<Cart> => {
    const response = await api.delete(`/cart/items/${productId}`);
    return response.data;
  },

  clear: async (): Promise<void> => {
    await api.delete('/cart');
  },
};
