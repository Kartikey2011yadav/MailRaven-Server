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

// Types
export interface User {
  username?: string;
  email: string;
  role?: string;
  created_at?: string;
  last_login_at?: string;
}

export interface Domain {
  id: number;
  name: string;
  status: string;
  created_at?: string;
  dkim_public_key?: string;
  dkim_selector?: string;
}

// User Management API
export const UserAPI = {
  list: (params?: { page?: number; limit?: number }) => 
    api.get<{ users: User[]; total: number }>('/admin/users', { params }),
  
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
    api.get<{ domains: Domain[]; total: number } | Domain[]>('/admin/domains', { params }),
  
  create: (name: string) => 
    api.post('/admin/domains', { name }),
  
  delete: (domain: string) => 
    api.delete(`/admin/domains/${domain}`),
};

// User Self-Management API
export const MeAPI = {
  changePassword: (data: { current_password: string; new_password: string }) =>
    api.put('/users/self/password', data),
};

export interface SieveScript {
    name: string;
    content: string;
    is_active: boolean;
    created_at?: string;
    updated_at?: string;
}

export const SieveAPI = {
    list: () => api.get<SieveScript[]>('/sieve/scripts'),
    create: (name: string, content: string) => api.post('/sieve/scripts', { name, content }),
    get: (name: string) => api.get<SieveScript>(`/sieve/scripts/${name}`),
    delete: (name: string) => api.delete(`/sieve/scripts/${name}`),
    activate: (name: string) => api.put(`/sieve/scripts/${name}/active`),
}

// Message API
export interface MessageSummary {
  id: string;
  sender: string;
  recipient: string;
  subject: string;
  snippet: string;
  read_state: boolean;
  received_at: string;
  spf_result?: string;
  dkim_result?: string;
  dmarc_result?: string;
}

export interface MessageFull extends MessageSummary {
  message_id: string;
  body: string;
  body_size: number;
}

export const MessageAPI = {
  list: (params?: { page?: number; limit?: number; offset?: number }) => 
    api.get<{ messages: MessageSummary[]; total: number; has_more: boolean }>('/messages', { params }),
  
  get: (id: string) => 
    api.get<MessageFull>(`/messages/${id}`),
  
  update: (id: string, data: { read_state?: boolean }) =>
    api.patch(`/messages/${id}`, data),
    
  send: (data: { to: string; subject: string; body: string }) => api.post('/messages/send', data), 
};

export default api;
