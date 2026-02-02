// Product types
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

export interface ProductsList {
  totalCount: number;
  totalPages: number;
  page: number;
  size: number;
  hasMore: boolean;
  products: Product[];
}

// Cart types
export interface CartItem {
  productId: string;
  name: string;
  price: number;
  quantity: number;
  imageUrl?: string;
}

export interface Cart {
  cartId: string;
  userId: string;
  items: CartItem[];
  totalAmount: number;
  createdAt: string;
  updatedAt: string;
}

// Order types
export type OrderStatus = 'pending' | 'confirmed' | 'processing' | 'shipped' | 'delivered' | 'cancelled';

export interface OrderItem {
  productId: string;
  name: string;
  price: number;
  quantity: number;
  subtotal: number;
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

export interface OrdersList {
  totalCount: number;
  totalPages: number;
  page: number;
  size: number;
  hasMore: boolean;
  orders: Order[];
}

// User types
export interface User {
  userId: string;
  email: string;
  name: string;
  createdAt: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  name: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

// API Response types
export interface ApiError {
  status: number;
  error: string;
  err_causes?: Record<string, string>;
}
