import { type ReactNode } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ThemeToggle } from './ThemeToggle';
import { Button } from './ui/button';
import { useAuthStore } from '../store/authStore';
import { LogOut, Database } from 'lucide-react';

interface LayoutProps {
    children: ReactNode;
}

export function Layout({ children }: LayoutProps) {
    const { user, logout } = useAuthStore();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <div className="min-h-screen bg-background">
            <header className="border-b">
                <div className="container mx-auto px-4 py-4">
                    <div className="flex items-center justify-between">
                        <Link to="/" className="flex items-center space-x-2">
                            <Database className="h-6 w-6" />
                            <span className="text-xl font-bold">dht</span>
                        </Link>

                        <div className="flex items-center space-x-4">
                            {user && (
                                <span className="text-sm text-muted-foreground">
                  {user.username}
                </span>
                            )}
                            <ThemeToggle />
                            {user && (
                                <Button variant="ghost" size="icon" onClick={handleLogout}>
                                    <LogOut className="h-5 w-5" />
                                </Button>
                            )}
                        </div>
                    </div>
                </div>
            </header>

            <main className="container mx-auto px-4 py-8">{children}</main>

            <footer className="border-t mt-auto">
                <div className="container mx-auto px-4 py-6 text-center text-sm text-muted-foreground">
                    © {new Date().getFullYear()} dht. All rights reserved.
                </div>
            </footer>
        </div>
    );
}