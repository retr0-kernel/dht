import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { type User } from '../lib/api';

interface AuthState {
    user: User | null;
    accessToken: string | null;
    setAuth: (user: User, accessToken: string) => void;
    logout: () => void;
    isAuthenticated: () => boolean;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set, get) => ({
            user: null,
            accessToken: null,

            setAuth: (user, accessToken) => {
                localStorage.setItem('access_token', accessToken);
                localStorage.setItem('user', JSON.stringify(user));
                set({ user, accessToken });
            },

            logout: () => {
                localStorage.removeItem('access_token');
                localStorage.removeItem('user');
                set({ user: null, accessToken: null });
            },

            isAuthenticated: () => {
                return get().accessToken !== null;
            },
        }),
        {
            name: 'auth-storage',
        }
    )
);