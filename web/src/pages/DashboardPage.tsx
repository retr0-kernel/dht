import { useEffect, useState } from 'react';
import { Layout } from '@/components/Layout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { APIKeysTab } from '@/components/dashboard/APIKeysTab';
import { UsageAnalyticsTab } from '@/components/dashboard/UsageAnalyticsTab';
import { KeyBrowserTab } from '@/components/dashboard/KeyBrowserTab';
import { OverviewTab } from '@/components/dashboard/OverviewTab';
import { useAuthStore } from '@/store/authStore';
import { apiKeyAPI, usageAPI } from '@/lib/api';
import { Activity, Key, Database, BarChart3 } from 'lucide-react';

export function DashboardPage() {
    const { user } = useAuthStore();
    const [stats, setStats] = useState({
        apiKeyCount: 0,
        totalRequests: 0,
        activeKeys: 0,
    });

    useEffect(() => {
        loadStats();
    }, []);

    const loadStats = async () => {
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
                apiKeyCount: keysResponse.count,
                totalRequests: usageStats.total_requests,
                activeKeys: keysResponse.api_keys.filter(k => k.is_active).length,
            });
        } catch (error) {
            console.error('Failed to load stats:', error);
        }
    };

    return (
        <Layout>
            <div className="space-y-8">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
                    <p className="text-muted-foreground mt-2">
                        Welcome back, {user?.username}
                    </p>
                </div>

                {/* Quick Stats */}
                <div className="grid gap-4 md:grid-cols-3">
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Total API Keys</CardTitle>
                            <Key className="h-4 w-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{stats.apiKeyCount}</div>
                            <p className="text-xs text-muted-foreground">
                                {stats.activeKeys} active
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
                            <Activity className="h-4 w-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{stats.totalRequests.toLocaleString()}</div>
                            <p className="text-xs text-muted-foreground">
                                All time
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Storage</CardTitle>
                            <Database className="h-4 w-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">Active</div>
                            <p className="text-xs text-muted-foreground">
                                All nodes operational
                            </p>
                        </CardContent>
                    </Card>
                </div>

                {/* Main Content Tabs */}
                <Tabs defaultValue="overview" className="space-y-4">
                    <TabsList>
                        <TabsTrigger value="overview">
                            <BarChart3 className="h-4 w-4 mr-2" />
                            Overview
                        </TabsTrigger>
                        <TabsTrigger value="keys">
                            <Key className="h-4 w-4 mr-2" />
                            API Keys
                        </TabsTrigger>
                        <TabsTrigger value="analytics">
                            <Activity className="h-4 w-4 mr-2" />
                            Analytics
                        </TabsTrigger>
                        <TabsTrigger value="browser">
                            <Database className="h-4 w-4 mr-2" />
                            Key Browser
                        </TabsTrigger>
                    </TabsList>

                    <TabsContent value="overview" className="space-y-4">
                        <OverviewTab />
                    </TabsContent>

                    <TabsContent value="keys" className="space-y-4">
                        <APIKeysTab onUpdate={loadStats} />
                    </TabsContent>

                    <TabsContent value="analytics" className="space-y-4">
                        <UsageAnalyticsTab />
                    </TabsContent>

                    <TabsContent value="browser" className="space-y-4">
                        <KeyBrowserTab />
                    </TabsContent>
                </Tabs>
            </div>
        </Layout>
    );
}