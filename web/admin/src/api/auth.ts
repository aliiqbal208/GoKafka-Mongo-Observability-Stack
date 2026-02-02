import client from './client';

interface LoginResponse {
  token: string;
  user: {
    userId: string;
    email: string;
    name: string;
  };
}

export const login = async (email: string, password: string): Promise<LoginResponse> => {
  const response = await client.post('/auth/login', { email, password });
  return response.data;
};

export const logout = (): void => {
  localStorage.removeItem('admin_token');
};

export const isAuthenticated = (): boolean => {
  return !!localStorage.getItem('admin_token');
};
