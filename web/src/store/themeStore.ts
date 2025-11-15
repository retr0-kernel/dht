import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type Theme = 'dark' | 'light' | 'system';

interface ThemeState {
    theme: Theme;
    setTheme: (theme: Theme) => void;
    effectiveTheme: () => 'dark' | 'light';
}

export const useThemeStore = create<ThemeState>()(
    persist(
        (set, get) => ({
            theme: 'system',

            setTheme: (theme) => {
                set({ theme });
                applyTheme(theme);
            },

            effectiveTheme: () => {
                const { theme } = get();
                if (theme === 'system') {
                    return window.matchMedia('(prefers-color-scheme: dark)').matches
                        ? 'dark'
                        : 'light';
                }
                return theme;
            },
        }),
        {
            name: 'theme-storage',
            onRehydrateStorage: () => (state) => {
                if (state) {
                    applyTheme(state.theme);
                }
            },
        }
    )
);

function applyTheme(theme: Theme) {
    const root = window.document.documentElement;
    root.classList.remove('light', 'dark');

    if (theme === 'system') {
        const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches
            ? 'dark'
            : 'light';
        root.classList.add(systemTheme);
    } else {
        root.classList.add(theme);
    }
}