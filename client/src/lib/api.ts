import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add a request interceptor to include the token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Auth API
export const AuthAPI = {
  login: (data: { email: string; password: string }) =>
    api.post('/auth/login', data),
};

// User Management API
export const UserAPI = {
  list: (params?: { page?: number; limit?: number }) => 
    api.get<{ users: any[]; total: number }>('/admin/users', { params }),
  
  create: (data: { email: string; password?: string; role: string }) => 
    api.post('/admin/users', data),
  
  updateRole: (email: string, role: string) => 
    api.put(`/admin/users/${email}/role`, { role }),
  
  delete: (email: string) => 
    api.delete(`/admin/users/${email}`),
};

// Domain Management API
export const DomainAPI = {
  list: (params?: { page?: number; limit?: number }) => 
    api.get<{ domains: any[]; total: number } | any[]>('/admin/domains', { params }),
  
  create: (name: string) => 
    api.post('/admin/domains', { name }),
  
  delete: (domain: string) => 
    api.delete(`/admin/domains/${domain}`),
};

export default api;
