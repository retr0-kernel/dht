import { useEffect, useState, useCallback } from 'react';
import { Layout } from '@/components/Layout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { APIKeysTab } from '@/components/dashboard/APIKeysTab';
import { UsageAnalyticsTab } from '@/components/dashboard/UsageAnalyticsTab';
import { KeyBrowserTab } from '@/components/dashboard/KeyBrowserTab';
import { OverviewTab } from '@/components/dashboard/OverviewTab';
import { useAuthStore } from '@/store/authStore';
import { apiKeyAPI, usageAPI } from '@/lib/api';
import { Activity, Key, Database, BarChart3, TrendingUp, Sparkles } from 'lucide-react';
import { motion } from 'framer-motion';

export function DashboardPage() {
    const { user } = useAuthStore();
    const [stats, setStats] = useState({
        apiKeyCount: 0,
        totalRequests: 0,
        activeKeys: 0,
    });

    const loadStats = useCallback(async () => {
        try {
            const [keysResponse, usageStats] = await Promise.all([
                apiKeyAPI.list(),
                usageAPI.stats().catch(() => ({
                    total_requests: 0,
                    successful_requests: 0,
                    failed_requests: 0,
                    total_bytes_transferred: 0,
                    average_latency_ms: 0,
                    requests_by_operation: {}
                })),
            ]);

            setStats({
                apiKeyCount: keysResponse?.count || 0,
                totalRequests: usageStats?.total_requests || 0,
                activeKeys: keysResponse?.api_keys?.filter(k => k.is_active).length || 0,
            });
        } catch (error) {
            console.error('Failed to load stats:', error);
        }
    }, []);

    useEffect(() => {
        loadStats();
    }, [loadStats]);

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

    const floatVariants = {
        animate: {
            y: [-5, 5],
            transition: {
                y: {
                    repeat: Infinity,
                    repeatType: 'reverse' as const,
                    duration: 2,
                    ease: 'easeInOut' as const
                }
            }
        }
    };

    const pulseVariants = {
        animate: {
            scale: [1, 1.2, 1],
            opacity: [1, 0.5, 1],
            transition: {
                repeat: Infinity,
                duration: 1.5,
                ease: 'easeInOut' as const
            }
        }
    };

    return (
        <Layout>
            <div className="space-y-8 pb-8">
                {/* Header with gradient */}
                <motion.div
                    initial={{ opacity: 0, y: -30 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.8, ease: 'easeOut' }}
                    className="relative"
                >
                    <div className="absolute inset-0 bg-gradient-to-r from-primary/10 via-primary/5 to-transparent rounded-3xl blur-3xl -z-10" />
                    <div className="flex items-center justify-between">
                        <div>
                            <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
                                Dashboard
                            </h1>
                            <p className="text-muted-foreground mt-2 flex items-center gap-2">
                                <Sparkles className="h-4 w-4 text-primary animate-pulse" />
                                Welcome back, <span className="font-semibold text-foreground">{user?.username}</span>
                            </p>
                        </div>
                        <div className="hidden md:flex items-center gap-2 px-4 py-2 bg-primary/10 rounded-full border border-primary/20">
                            <motion.div
                                variants={pulseVariants}
                                animate="animate"
                                className="w-2 h-2 bg-green-500 rounded-full"
                            />
                            <span className="text-sm font-medium">All Systems Operational</span>
                        </div>
                    </div>
                </motion.div>

                {/* Quick Stats with enhanced cards */}
                <motion.div
                    variants={containerVariants}
                    initial="hidden"
                    animate="visible"
                    className="grid gap-6 md:grid-cols-3"
                >
                    <motion.div variants={itemVariants}>
                        <Card className="relative overflow-hidden border-2 transition-all duration-300 hover:shadow-xl hover:scale-105 hover:border-primary/50 group">
                            <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-transparent" />
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Total API Keys</CardTitle>
                                <motion.div
                                    variants={floatVariants}
                                    animate="animate"
                                    className="p-2 rounded-lg bg-blue-500/10 group-hover:bg-blue-500/20 transition-colors"
                                >
                                    <Key className="h-5 w-5 text-blue-500" />
                                </motion.div>
                            </CardHeader>
                            <CardContent>
                                <motion.div
                                    initial={{ scale: 0.5, opacity: 0 }}
                                    animate={{ scale: 1, opacity: 1 }}
                                    transition={{ delay: 0.5, type: 'spring', stiffness: 200 }}
                                    className="text-3xl font-bold bg-gradient-to-r from-blue-600 to-blue-400 bg-clip-text text-transparent"
                                >
                                    {stats.apiKeyCount}
                                </motion.div>
                                <div className="flex items-center gap-2 mt-2">
                                    <div className="flex items-center gap-1 text-xs">
                                        <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                                        <span className="text-muted-foreground">
                                            {stats.activeKeys} active
                                        </span>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </motion.div>

                    <motion.div variants={itemVariants}>
                        <Card className="relative overflow-hidden border-2 transition-all duration-300 hover:shadow-xl hover:scale-105 hover:border-primary/50 group">
                            <div className="absolute inset-0 bg-gradient-to-br from-purple-500/10 to-transparent" />
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
                                <motion.div
                                    variants={floatVariants}
                                    animate="animate"
                                    className="p-2 rounded-lg bg-purple-500/10 group-hover:bg-purple-500/20 transition-colors"
                                >
                                    <Activity className="h-5 w-5 text-purple-500" />
                                </motion.div>
                            </CardHeader>
                            <CardContent>
                                <motion.div
                                    initial={{ scale: 0.5, opacity: 0 }}
                                    animate={{ scale: 1, opacity: 1 }}
                                    transition={{ delay: 0.6, type: 'spring', stiffness: 200 }}
                                    className="text-3xl font-bold bg-gradient-to-r from-purple-600 to-purple-400 bg-clip-text text-transparent"
                                >
                                    {stats.totalRequests.toLocaleString()}
                                </motion.div>
                                <div className="flex items-center gap-2 mt-2">
                                    <TrendingUp className="h-3 w-3 text-green-500" />
                                    <span className="text-xs text-muted-foreground">All time</span>
                                </div>
                            </CardContent>
                        </Card>
                    </motion.div>

                    <motion.div variants={itemVariants}>
                        <Card className="relative overflow-hidden border-2 transition-all duration-300 hover:shadow-xl hover:scale-105 hover:border-primary/50 group">
                            <div className="absolute inset-0 bg-gradient-to-br from-green-500/10 to-transparent" />
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Storage Status</CardTitle>
                                <motion.div
                                    variants={floatVariants}
                                    animate="animate"
                                    className="p-2 rounded-lg bg-green-500/10 group-hover:bg-green-500/20 transition-colors"
                                >
                                    <Database className="h-5 w-5 text-green-500" />
                                </motion.div>
                            </CardHeader>
                            <CardContent>
                                <motion.div
                                    initial={{ scale: 0.5, opacity: 0 }}
                                    animate={{ scale: 1, opacity: 1 }}
                                    transition={{ delay: 0.7, type: 'spring', stiffness: 200 }}
                                    className="text-3xl font-bold bg-gradient-to-r from-green-600 to-green-400 bg-clip-text text-transparent"
                                >
                                    Active
                                </motion.div>
                                <div className="flex items-center gap-2 mt-2">
                                    <div className="flex items-center gap-1">
                                        <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                                        <span className="text-xs text-muted-foreground">
                                            All nodes operational
                                        </span>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </motion.div>
                </motion.div>

                {/* Main Content Tabs with animations */}
                <motion.div
                    initial={{ opacity: 0, y: 40 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 1, delay: 0.4, ease: 'easeOut' }}
                >
                    <Tabs defaultValue="overview" className="space-y-6">
                        <TabsList className="grid w-full grid-cols-4 h-auto p-1 bg-muted/50 backdrop-blur-sm">
                            <TabsTrigger
                                value="overview"
                                className="data-[state=active]:bg-background data-[state=active]:shadow-md transition-all duration-200"
                            >
                                <BarChart3 className="h-4 w-4 mr-2" />
                                <span className="hidden sm:inline">Overview</span>
                            </TabsTrigger>
                            <TabsTrigger
                                value="keys"
                                className="data-[state=active]:bg-background data-[state=active]:shadow-md transition-all duration-200"
                            >
                                <Key className="h-4 w-4 mr-2" />
                                <span className="hidden sm:inline">API Keys</span>
                            </TabsTrigger>
                            <TabsTrigger
                                value="analytics"
                                className="data-[state=active]:bg-background data-[state=active]:shadow-md transition-all duration-200"
                            >
                                <Activity className="h-4 w-4 mr-2" />
                                <span className="hidden sm:inline">Analytics</span>
                            </TabsTrigger>
                            <TabsTrigger
                                value="browser"
                                className="data-[state=active]:bg-background data-[state=active]:shadow-md transition-all duration-200"
                            >
                                <Database className="h-4 w-4 mr-2" />
                                <span className="hidden sm:inline">Key Browser</span>
                            </TabsTrigger>
                        </TabsList>

                        <TabsContent value="overview" className="space-y-4">
                            <motion.div
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ duration: 0.5 }}
                            >
                                <OverviewTab />
                            </motion.div>
                        </TabsContent>

                        <TabsContent value="keys" className="space-y-4">
                            <motion.div
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ duration: 0.5 }}
                            >
                                <APIKeysTab onUpdate={loadStats} />
                            </motion.div>
                        </TabsContent>

                        <TabsContent value="analytics" className="space-y-4">
                            <motion.div
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ duration: 0.5 }}
                            >
                                <UsageAnalyticsTab />
                            </motion.div>
                        </TabsContent>

                        <TabsContent value="browser" className="space-y-4">
                            <motion.div
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ duration: 0.5 }}
                            >
                                <KeyBrowserTab />
                            </motion.div>
                        </TabsContent>
                    </Tabs>
                </motion.div>
            </div>
        </Layout>
    );
}

