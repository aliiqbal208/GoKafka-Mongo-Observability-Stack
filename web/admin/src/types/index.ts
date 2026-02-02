export interface Product {
  productId: string;
  categoryId: string;
  name: string;
  description: string;
  price: number;
  imageUrl?: string;
  photos?: string[];
  quantity: number;
  stock: number;
  rating: number;
  createdAt: string;
  updatedAt: string;
}

export interface Order {
  orderId: string;
  userId: string;
  items: OrderItem[];
  totalAmount: number;
  status: OrderStatus;
  shippingAddress: string;
  createdAt: string;
  updatedAt: string;
}

export interface OrderItem {
  productId: string;
  productName: string;
  quantity: number;
  price: number;
}

export type OrderStatus = 'pending' | 'processing' | 'shipped' | 'delivered' | 'cancelled';

export interface User {
  userId: string;
  email: string;
  name: string;
  createdAt: string;
}

export interface DashboardStats {
  totalOrders: number;
  totalRevenue: number;
  totalProducts: number;
  totalUsers: number;
  pendingOrders: number;
  recentOrders: Order[];
}

export interface PaginatedResponse<T> {
  totalCount: number;
  totalPages: number;
  page: number;
  size: number;
  hasMore: boolean;
  products?: T[];
  orders?: T[];
}
