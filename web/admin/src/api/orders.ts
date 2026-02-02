import client from './client';
import { Order, PaginatedResponse, OrderStatus } from '../types';

export const getOrders = async (page = 1, size = 10): Promise<PaginatedResponse<Order>> => {
  const response = await client.get(`/orders?page=${page}&size=${size}`);
  return response.data;
};

export const getOrder = async (id: string): Promise<Order> => {
  const response = await client.get(`/orders/${id}`);
  return response.data;
};

export const updateOrderStatus = async (id: string, status: OrderStatus): Promise<Order> => {
  const response = await client.patch(`/orders/${id}/status`, { status });
  return response.data;
};

export const getOrdersByStatus = async (status: OrderStatus, page = 1, size = 10): Promise<PaginatedResponse<Order>> => {
  const response = await client.get(`/orders?status=${status}&page=${page}&size=${size}`);
  return response.data;
};
