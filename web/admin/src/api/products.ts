import client from './client';
import { Product, PaginatedResponse } from '../types';

export const getProducts = async (page = 1, size = 10): Promise<PaginatedResponse<Product>> => {
  const response = await client.get(`/products?page=${page}&size=${size}`);
  return response.data;
};

export const getProduct = async (id: string): Promise<Product> => {
  const response = await client.get(`/products/${id}`);
  return response.data;
};

export const createProduct = async (product: Omit<Product, 'productId' | 'createdAt' | 'updatedAt'>): Promise<Product> => {
  const response = await client.post('/products', product);
  return response.data;
};

export const updateProduct = async (id: string, product: Partial<Product>): Promise<Product> => {
  const response = await client.put(`/products/${id}`, product);
  return response.data;
};

export const deleteProduct = async (id: string): Promise<void> => {
  await client.delete(`/products/${id}`);
};

export const searchProducts = async (query: string, page = 1, size = 10): Promise<PaginatedResponse<Product>> => {
  const response = await client.get(`/products/search?search=${query}&page=${page}&size=${size}`);
  return response.data;
};
