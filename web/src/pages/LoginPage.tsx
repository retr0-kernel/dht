import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { authAPI } from '@/lib/api';
import { useAuthStore } from '@/store/authStore';
import { Database, Loader2, Mail, Lock, Sparkles, Shield } from 'lucide-react';
import { ThemeToggle } from '@/components/ThemeToggle';
import { motion } from 'framer-motion';

export function LoginPage() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();
    const { setAuth } = useAuthStore();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            const response = await authAPI.login({ email, password });
            setAuth(response.user, response.access_token);
            navigate('/dashboard');
        } catch (err: any) {
            setError(err.response?.data?.error || 'Login failed. Please check your credentials.');
        } finally {
            setLoading(false);
        }
    };

    const containerVariants = {
        hidden: { opacity: 0 },
        visible: {
            opacity: 1,
            transition: {
                staggerChildren: 0.1,
                delayChildren: 0.2
            }
        }
    };

    const itemVariants = {
        hidden: { y: 20, opacity: 0 },
        visible: {
            y: 0,
            opacity: 1,
            transition: {
                type: 'spring' as const,
                stiffness: 100,
                damping: 12
            }
        }
    };

    const iconFloatVariants = {
        animate: {
            y: [-8, 8],
            rotate: [-5, 5],
            transition: {
                y: {
                    repeat: Infinity,
                    repeatType: 'reverse' as const,
                    duration: 3,
                    ease: 'easeInOut' as const
                },
                rotate: {
                    repeat: Infinity,
                    repeatType: 'reverse' as const,
                    duration: 4,
                    ease: 'easeInOut' as const
                }
            }
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-background p-4 relative overflow-hidden">
            {/* Animated background elements */}
            <div className="absolute inset-0 overflow-hidden pointer-events-none">
                <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 0.1, scale: 1 }}
                    transition={{ duration: 1 }}
                    className="absolute top-1/4 -left-20 w-96 h-96 bg-primary rounded-full blur-3xl"
                />
                <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 0.1, scale: 1 }}
                    transition={{ duration: 1, delay: 0.3 }}
                    className="absolute bottom-1/4 -right-20 w-96 h-96 bg-purple-500 rounded-full blur-3xl"
                />
                <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 0.05, scale: 1 }}
                    transition={{ duration: 1, delay: 0.6 }}
                    className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-blue-500 rounded-full blur-3xl"
                />
            </div>

            <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.3 }}
                className="absolute top-4 right-4 z-10"
            >
                <ThemeToggle />
            </motion.div>

            <motion.div
                variants={containerVariants}
                initial="hidden"
                animate="visible"
                className="relative z-10 w-full max-w-md"
            >
                <Card className="backdrop-blur-sm bg-background/95 border-2 shadow-2xl">
                    <CardHeader className="space-y-1">
                        <motion.div
                            variants={itemVariants}
                            className="flex items-center justify-center mb-4 relative"
                        >
                            <motion.div
                                variants={iconFloatVariants}
                                animate="animate"
                                className="relative"
                            >
                                <div className="absolute inset-0 bg-primary/20 rounded-full blur-xl animate-pulse" />
                                <Database className="h-16 w-16 text-primary relative z-10" />
                            </motion.div>
                        </motion.div>

                        <motion.div variants={itemVariants}>
                            <CardTitle className="text-3xl text-center font-bold bg-gradient-to-r from-primary to-primary/70 bg-clip-text text-transparent">
                                Welcome back
                            </CardTitle>
                        </motion.div>

                        <motion.div variants={itemVariants}>
                            <CardDescription className="text-center flex items-center justify-center gap-2">
                                <Shield className="h-4 w-4" />
                                Secure access to your distributed store
                            </CardDescription>
                        </motion.div>
                    </CardHeader>

                    <form onSubmit={handleSubmit}>
                        <CardContent className="space-y-4">
                            {error && (
                                <motion.div
                                    initial={{ opacity: 0, scale: 0.95, y: -10 }}
                                    animate={{ opacity: 1, scale: 1, y: 0 }}
                                    className="bg-destructive/10 text-destructive text-sm p-3 rounded-md border border-destructive/20 flex items-start gap-2"
                                >
                                    <Sparkles className="h-4 w-4 mt-0.5 flex-shrink-0" />
                                    <span>{error}</span>
                                </motion.div>
                            )}

                            <motion.div variants={itemVariants} className="space-y-2">
                                <Label htmlFor="email" className="flex items-center gap-2">
                                    <Mail className="h-4 w-4" />
                                    Email
                                </Label>
                                <Input
                                    id="email"
                                    type="email"
                                    placeholder="you@example.com"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    required
                                    disabled={loading}
                                    className="transition-all duration-200 focus:scale-[1.02] focus:shadow-lg"
                                />
                            </motion.div>

                            <motion.div variants={itemVariants} className="space-y-2">
                                <Label htmlFor="password" className="flex items-center gap-2">
                                    <Lock className="h-4 w-4" />
                                    Password
                                </Label>
                                <Input
                                    id="password"
                                    type="password"
                                    placeholder="••••••••"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    required
                                    disabled={loading}
                                    className="transition-all duration-200 focus:scale-[1.02] focus:shadow-lg"
                                />
                            </motion.div>
                        </CardContent>

                        <CardFooter className="flex flex-col space-y-4">
                            <motion.div
                                variants={itemVariants}
                                className="w-full"
                                whileHover={{ scale: 1.02 }}
                                whileTap={{ scale: 0.98 }}
                            >
                                <Button
                                    type="submit"
                                    className="w-full relative overflow-hidden group backdrop-blur-xl bg-gradient-to-r from-primary/90 via-primary to-primary/90 dark:from-primary/80 dark:via-primary/90 dark:to-primary/80 border-2 border-primary/30 dark:border-primary/50 shadow-[0_0_30px_rgba(59,130,246,0.3)] dark:shadow-[0_0_30px_rgba(59,130,246,0.5)] hover:shadow-[0_0_50px_rgba(59,130,246,0.5)] dark:hover:shadow-[0_0_50px_rgba(59,130,246,0.7)] transition-all duration-300"
                                    disabled={loading}
                                >
                                    {/* Glassmorphism overlay */}
                                    <span className="absolute inset-0 bg-white/10 dark:bg-white/5" />

                                    {/* Shimmer effect */}
                                    <span className="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 dark:via-white/20 to-transparent translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-1000" />

                                    {/* Animated border glow */}
                                    <span className="absolute inset-0 rounded-md opacity-0 group-hover:opacity-100 transition-opacity duration-300 bg-gradient-to-r from-blue-400/20 via-purple-400/20 to-blue-400/20 blur-sm" />

                                    {/* Button content with loading animation */}
                                    <span className="relative flex items-center justify-center">
                                        {loading ? (
                                            <motion.span
                                                initial={{ opacity: 0, y: 10 }}
                                                animate={{ opacity: 1, y: 0 }}
                                                className="flex items-center"
                                            >
                                                <motion.span
                                                    animate={{ rotate: 360 }}
                                                    transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                                                >
                                                    <Loader2 className="mr-2 h-4 w-4" />
                                                </motion.span>
                                                <motion.span
                                                    animate={{ opacity: [1, 0.5, 1] }}
                                                    transition={{ duration: 1.5, repeat: Infinity }}
                                                >
                                                    Signing in
                                                </motion.span>
                                                <motion.span
                                                    animate={{ opacity: [0, 1, 0] }}
                                                    transition={{ duration: 1.5, repeat: Infinity, delay: 0 }}
                                                >
                                                    .
                                                </motion.span>
                                                <motion.span
                                                    animate={{ opacity: [0, 1, 0] }}
                                                    transition={{ duration: 1.5, repeat: Infinity, delay: 0.2 }}
                                                >
                                                    .
                                                </motion.span>
                                                <motion.span
                                                    animate={{ opacity: [0, 1, 0] }}
                                                    transition={{ duration: 1.5, repeat: Infinity, delay: 0.4 }}
                                                >
                                                    .
                                                </motion.span>
                                            </motion.span>
                                        ) : (
                                            <>
                                                <motion.span
                                                    whileHover={{ rotate: [0, -10, 10, -10, 0], scale: [1, 1.2, 1] }}
                                                    transition={{ duration: 0.5 }}
                                                >
                                                    <Sparkles className="mr-2 h-4 w-4" />
                                                </motion.span>
                                                Sign in
                                            </>
                                        )}
                                    </span>
                                </Button>
                            </motion.div>

                            <motion.p
                                variants={itemVariants}
                                className="text-sm text-center text-muted-foreground"
                            >
                                Don't have an account?{' '}
                                <Link
                                    to="/signup"
                                    className="text-primary hover:underline font-semibold inline-flex items-center gap-1 transition-all hover:gap-2"
                                >
                                    Sign up
                                    <Sparkles className="h-3 w-3" />
                                </Link>
                            </motion.p>
                        </CardFooter>
                    </form>
                </Card>

                {/* Floating feature badges */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.8, duration: 0.6 }}
                    className="mt-8 flex justify-center gap-4 flex-wrap"
                >
                    <motion.div
                        whileHover={{ scale: 1.05, y: -2 }}
                        className="px-4 py-2 rounded-full bg-primary/10 border border-primary/20 backdrop-blur-sm text-sm flex items-center gap-2"
                    >
                        <Shield className="h-4 w-4 text-primary" />
                        <span className="text-muted-foreground">Encrypted</span>
                    </motion.div>
                    <motion.div
                        whileHover={{ scale: 1.05, y: -2 }}
                        className="px-4 py-2 rounded-full bg-purple-500/10 border border-purple-500/20 backdrop-blur-sm text-sm flex items-center gap-2"
                    >
                        <Database className="h-4 w-4 text-purple-500" />
                        <span className="text-muted-foreground">Distributed</span>
                    </motion.div>
                    <motion.div
                        whileHover={{ scale: 1.05, y: -2 }}
                        className="px-4 py-2 rounded-full bg-blue-500/10 border border-blue-500/20 backdrop-blur-sm text-sm flex items-center gap-2"
                    >
                        <Sparkles className="h-4 w-4 text-blue-500" />
                        <span className="text-muted-foreground">Fast</span>
                    </motion.div>
                </motion.div>
            </motion.div>
        </div>
    );
}