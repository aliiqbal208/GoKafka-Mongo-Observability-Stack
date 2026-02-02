import api from './client';
import type { Order, OrdersList } from '../types';

export interface CreateOrderRequest {
  shippingAddress: string;
}

export const orderApi = {
  create: async (data: CreateOrderRequest): Promise<Order> => {
    const response = await api.post('/orders', data);
    return response.data;
  },

  getAll: async (page = 1, size = 10): Promise<OrdersList> => {
    const response = await api.get(`/orders?page=${page}&size=${size}`);
    return response.data;
  },

  getById: async (id: string): Promise<Order> => {
    const response = await api.get(`/orders/${id}`);
    return response.data;
  },
};
