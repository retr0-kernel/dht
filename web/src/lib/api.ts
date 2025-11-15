import axios, { AxiosError } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081';
const GATEWAY_BASE_URL = import.meta.env.VITE_GATEWAY_BASE_URL || 'http://localhost:8080';

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

const gatewayApi = axios.create({
    baseURL: GATEWAY_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// Response interceptor for error handling
api.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
        if (error.response?.status === 401) {
            localStorage.removeItem('access_token');
            localStorage.removeItem('user');
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

// Types
export interface User {
    id: number;
    email: string;
    username: string;
    is_active: boolean;
    created_at: string;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    access_token: string;
    refresh_token: string;
    token_type: string;
    expires_in: number;
    user: User;
}

export interface SignupRequest {
    email: string;
    username: string;
    password: string;
}

export interface APIKey {
    id: number;
    name: string;
    key_prefix: string;
    key?: string; // Only present on creation
    scopes: string[];
    is_active: boolean;
    last_used_at?: string;
    expires_at?: string;
    created_at: string;
}

export interface CreateAPIKeyRequest {
    name: string;
    scopes?: string[];
    expires_in_days?: number;
}

export interface UsageRecord {
    id: number;
    user_id: number;
    api_key_id?: number;
    operation: string;
    key_accessed?: string;
    request_size_bytes: number;
    response_size_bytes: number;
    status_code: number;
    duration_ms: number;
    ip_address?: string;
    user_agent?: string;
    error_message?: string;
    created_at: string;
}

export interface UsageStats {
    total_requests: number;
    successful_requests: number;
    failed_requests: number;
    total_bytes_transferred: number;
    average_latency_ms: number;
    requests_by_operation: {
        [key: string]: number;
    };
}

// Auth APIs
export const authAPI = {
    login: async (data: LoginRequest): Promise<LoginResponse> => {
        const response = await api.post<LoginResponse>('/login', data);
        return response.data;
    },

    signup: async (data: SignupRequest): Promise<User> => {
        const response = await api.post<User>('/signup', data);
        return response.data;
    },
};

// API Key APIs
export const apiKeyAPI = {
    list: async (): Promise<{ api_keys: APIKey[]; count: number }> => {
        const response = await api.get('/apikeys');
        return response.data;
    },

    create: async (data: CreateAPIKeyRequest): Promise<APIKey> => {
        const response = await api.post<APIKey>('/apikeys', data);
        return response.data;
    },

    delete: async (id: number): Promise<void> => {
        await api.delete(`/apikeys/${id}`);
    },
};

// Usage APIs (you'll need to implement these endpoints in usermanager)
export const usageAPI = {
    list: async (params?: {
        start_date?: string;
        end_date?: string;
        limit?: number;
    }): Promise<UsageRecord[]> => {
        const response = await api.get('/usage', { params });
        return response.data;
    },

    stats: async (params?: {
        start_date?: string;
        end_date?: string;
    }): Promise<UsageStats> => {
        const response = await api.get('/usage/stats', { params });
        return response.data;
    },
};

// Gateway KV APIs (for key browser)
export const kvAPI = {
    get: async (key: string, apiKey: string): Promise<any> => {
        const response = await gatewayApi.get(`/v1/kv/${key}`, {
            headers: {
                'X-API-Key': apiKey,
            },
        });
        return response.data;
    },

    put: async (key: string, value: any, apiKey: string, ttl?: string): Promise<void> => {
        const url = ttl ? `/v1/kv/${key}?ttl=${ttl}` : `/v1/kv/${key}`;
        await gatewayApi.put(url, value, {
            headers: {
                'X-API-Key': apiKey,
                'Content-Type': 'application/json',
            },
        });
    },

    delete: async (key: string, apiKey: string): Promise<void> => {
        await gatewayApi.delete(`/v1/kv/${key}`, {
            headers: {
                'X-API-Key': apiKey,
            },
        });
    },

    list: async (apiKey: string): Promise<{ keys: any[]; count: number }> => {
        const response = await gatewayApi.get('/v1/kv', {
            headers: {
                'X-API-Key': apiKey,
            },
        });
        return response.data;
    },
};

export { api, gatewayApi };