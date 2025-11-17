import { Moon, Sun } from 'lucide-react';
import { Button } from './ui/button';
import { useThemeStore } from '../store/themeStore';

type PlainTheme = 'dark' | 'light';

export function ThemeToggle() {
    const { setTheme, effectiveTheme } = useThemeStore();
    const currentTheme = effectiveTheme();

    const handleToggle = (e: React.MouseEvent<HTMLButtonElement>) => {
        const nextTheme: PlainTheme = currentTheme === 'dark' ? 'light' : 'dark';
        const x = e.clientX;
        const y = e.clientY;
        const ripple = document.createElement('span');
        ripple.className = 'theme-ripple';
        const maxDim = Math.max(window.innerWidth, window.innerHeight);
        ripple.style.left = x + 'px';
        ripple.style.top = y + 'px';
        ripple.style.width = ripple.style.height = maxDim * 2 + 'px';
        ripple.style.background = nextTheme === 'dark' ? 'hsl(222.2 84% 4.9%)' : 'hsl(0 0% 100%)';
        document.body.appendChild(ripple);
        ripple.addEventListener('animationend', () => ripple.remove());
        setTimeout(() => setTheme(nextTheme), 50);
    };

    return (
        <Button
            variant="ghost"
            size="icon"
            onClick={handleToggle}
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