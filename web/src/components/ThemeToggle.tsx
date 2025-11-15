import { Moon, Sun } from 'lucide-react';
import { Button } from './ui/button';
import { useThemeStore } from '../store/themeStore';

export function ThemeToggle() {
    // const { theme, setTheme, effectiveTheme } = useThemeStore();
    const { setTheme, effectiveTheme } = useThemeStore();
    const currentTheme = effectiveTheme();

    return (
        <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(currentTheme === 'dark' ? 'light' : 'dark')}
        >
            {currentTheme === 'dark' ? (
                <Sun className="h-5 w-5" />
            ) : (
                <Moon className="h-5 w-5" />
            )}
            <span className="sr-only">Toggle theme</span>
        </Button>
    );
}