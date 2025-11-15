import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { authAPI } from '@/lib/api';
import { Database, Loader2, Mail, Lock, User, Sparkles, Shield, CheckCircle2 } from 'lucide-react';
import { ThemeToggle } from '@/components/ThemeToggle';
import { motion } from 'framer-motion';

export function SignupPage() {
    const [email, setEmail] = useState('');
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    // Password strength indicator
    const getPasswordStrength = (pass: string) => {
        if (pass.length === 0) return { strength: 0, label: '', color: '' };
        if (pass.length < 8) return { strength: 1, label: 'Weak', color: 'bg-red-500' };
        if (pass.length < 12) return { strength: 2, label: 'Medium', color: 'bg-yellow-500' };
        return { strength: 3, label: 'Strong', color: 'bg-green-500' };
    };

    const passwordStrength = getPasswordStrength(password);
    const passwordsMatch = confirmPassword.length > 0 && password === confirmPassword;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        if (password !== confirmPassword) {
            setError('Passwords do not match');
            return;
        }

        if (password.length < 8) {
            setError('Password must be at least 8 characters');
            return;
        }

        setLoading(true);

        try {
            await authAPI.signup({ email, username, password });
            navigate('/login', {
                state: { message: 'Account created successfully! Please login.' }
            });
        } catch (err: any) {
            setError(
                err.response?.data?.error ||
                'Signup failed. Email or username may already exist.'
            );
        } finally {
            setLoading(false);
        }
    };

    const containerVariants = {
        hidden: { opacity: 0 },
        visible: {
            opacity: 1,
            transition: {
                staggerChildren: 0.08,
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
                    className="absolute top-1/4 -left-20 w-96 h-96 bg-green-500 rounded-full blur-3xl"
                />
                <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 0.1, scale: 1 }}
                    transition={{ duration: 1, delay: 0.3 }}
                    className="absolute bottom-1/4 -right-20 w-96 h-96 bg-blue-500 rounded-full blur-3xl"
                />
                <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 0.05, scale: 1 }}
                    transition={{ duration: 1, delay: 0.6 }}
                    className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-purple-500 rounded-full blur-3xl"
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
                                <div className="absolute inset-0 bg-green-500/20 rounded-full blur-xl animate-pulse" />
                                <Database className="h-16 w-16 text-green-500 relative z-10" />
                            </motion.div>
                        </motion.div>

                        <motion.div variants={itemVariants}>
                            <CardTitle className="text-3xl text-center font-bold bg-gradient-to-r from-green-600 to-green-400 bg-clip-text text-transparent">
                                Create an account
                            </CardTitle>
                        </motion.div>

                        <motion.div variants={itemVariants}>
                            <CardDescription className="text-center flex items-center justify-center gap-2">
                                <Sparkles className="h-4 w-4" />
                                Join the distributed future
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
                                    <Shield className="h-4 w-4 mt-0.5 flex-shrink-0" />
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
                                <Label htmlFor="username" className="flex items-center gap-2">
                                    <User className="h-4 w-4" />
                                    Username
                                </Label>
                                <Input
                                    id="username"
                                    type="text"
                                    placeholder="johndoe"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
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
                                {password.length > 0 && (
                                    <motion.div
                                        initial={{ opacity: 0, height: 0 }}
                                        animate={{ opacity: 1, height: 'auto' }}
                                        className="space-y-2"
                                    >
                                        <div className="flex items-center gap-2 text-xs">
                                            <div className="flex-1 bg-muted rounded-full h-1.5 overflow-hidden">
                                                <motion.div
                                                    initial={{ width: 0 }}
                                                    animate={{ width: `${(passwordStrength.strength / 3) * 100}%` }}
                                                    className={`h-full ${passwordStrength.color}`}
                                                    transition={{ duration: 0.3 }}
                                                />
                                            </div>
                                            <span className={`font-medium ${
                                                passwordStrength.strength === 1 ? 'text-red-500' :
                                                passwordStrength.strength === 2 ? 'text-yellow-500' :
                                                'text-green-500'
                                            }`}>
                                                {passwordStrength.label}
                                            </span>
                                        </div>
                                        <p className="text-xs text-muted-foreground">
                                            At least 8 characters recommended
                                        </p>
                                    </motion.div>
                                )}
                            </motion.div>

                            <motion.div variants={itemVariants} className="space-y-2">
                                <Label htmlFor="confirmPassword" className="flex items-center gap-2">
                                    <Lock className="h-4 w-4" />
                                    Confirm Password
                                </Label>
                                <div className="relative">
                                    <Input
                                        id="confirmPassword"
                                        type="password"
                                        placeholder="••••••••"
                                        value={confirmPassword}
                                        onChange={(e) => setConfirmPassword(e.target.value)}
                                        required
                                        disabled={loading}
                                        className="transition-all duration-200 focus:scale-[1.02] focus:shadow-lg"
                                    />
                                    {passwordsMatch && (
                                        <motion.div
                                            initial={{ scale: 0, opacity: 0 }}
                                            animate={{ scale: 1, opacity: 1 }}
                                            className="absolute right-3 top-1/2 -translate-y-1/2"
                                        >
                                            <CheckCircle2 className="h-5 w-5 text-green-500" />
                                        </motion.div>
                                    )}
                                </div>
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
                                    className="w-full relative overflow-hidden group bg-gradient-to-r from-green-600 to-green-500 hover:from-green-700 hover:to-green-600"
                                    disabled={loading}
                                >
                                    <span className="absolute inset-0 bg-gradient-to-r from-green-500/0 via-white/20 to-green-500/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-1000" />
                                    {loading ? (
                                        <>
                                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                            Creating account...
                                        </>
                                    ) : (
                                        <>
                                            <Sparkles className="mr-2 h-4 w-4" />
                                            Create Account
                                        </>
                                    )}
                                </Button>
                            </motion.div>

                            <motion.p
                                variants={itemVariants}
                                className="text-sm text-center text-muted-foreground"
                            >
                                Already have an account?{' '}
                                <Link
                                    to="/login"
                                    className="text-primary hover:underline font-semibold inline-flex items-center gap-1 transition-all hover:gap-2"
                                >
                                    Sign in
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
                        className="px-4 py-2 rounded-full bg-green-500/10 border border-green-500/20 backdrop-blur-sm text-sm flex items-center gap-2"
                    >
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                        <span className="text-muted-foreground">Free Forever</span>
                    </motion.div>
                    <motion.div
                        whileHover={{ scale: 1.05, y: -2 }}
                        className="px-4 py-2 rounded-full bg-blue-500/10 border border-blue-500/20 backdrop-blur-sm text-sm flex items-center gap-2"
                    >
                        <Shield className="h-4 w-4 text-blue-500" />
                        <span className="text-muted-foreground">Secure</span>
                    </motion.div>
                    <motion.div
                        whileHover={{ scale: 1.05, y: -2 }}
                        className="px-4 py-2 rounded-full bg-purple-500/10 border border-purple-500/20 backdrop-blur-sm text-sm flex items-center gap-2"
                    >
                        <Database className="h-4 w-4 text-purple-500" />
                        <span className="text-muted-foreground">Scalable</span>
                    </motion.div>
                </motion.div>
            </motion.div>
        </div>
    );
}

