import api from './client';
import type { Product, ProductsList } from '../types';

export const productApi = {
  getAll: async (page = 1, size = 12): Promise<ProductsList> => {
    const response = await api.get(`/products?page=${page}&size=${size}`);
    return response.data;
  },

  getById: async (id: string): Promise<Product> => {
    const response = await api.get(`/products/${id}`);
    return response.data;
  },

  search: async (query: string, page = 1, size = 12): Promise<ProductsList> => {
    const response = await api.get(`/products/search?search=${query}&page=${page}&size=${size}`);
    return response.data;
  },
};
