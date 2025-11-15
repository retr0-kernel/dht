import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { usageAPI, type UsageStats } from '@/lib/api.ts';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';
import { Loader2 } from 'lucide-react';

export function OverviewTab() {
    const [stats, setStats] = useState<UsageStats | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadStats();
    }, []);

    const loadStats = async () => {
        try {
            const data = await usageAPI.stats();
            setStats(data);
        } catch (error) {
            console.error('Failed to load usage stats:', error);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (!stats) {
        return (
            <Card>
                <CardContent className="pt-6">
                    <p className="text-center text-muted-foreground">No usage data available</p>
                </CardContent>
            </Card>
        );
    }

    const operationData = Object.entries(stats.requests_by_operation).map(([name, value]) => ({
        name: name.toUpperCase(),
        value,
    }));

    const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

    const successRate = stats.total_requests > 0
        ? ((stats.successful_requests / stats.total_requests) * 100).toFixed(1)
        : '0';

    return (
        <div className="space-y-4">
            {/* Stats Grid */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Total Requests</CardDescription>
                        <CardTitle className="text-3xl">{stats.total_requests.toLocaleString()}</CardTitle>
                    </CardHeader>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Success Rate</CardDescription>
                        <CardTitle className="text-3xl">{successRate}%</CardTitle>
                    </CardHeader>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Avg Latency</CardDescription>
                        <CardTitle className="text-3xl">{stats.average_latency_ms.toFixed(0)}ms</CardTitle>
                    </CardHeader>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Data Transferred</CardDescription>
                        <CardTitle className="text-3xl">
                            {(stats.total_bytes_transferred / 1024 / 1024).toFixed(2)}MB
                        </CardTitle>
                    </CardHeader>
                </Card>
            </div>

            {/* Charts */}
            <div className="grid gap-4 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle>Requests by Operation</CardTitle>
                        <CardDescription>Distribution of API operations</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {operationData.length > 0 ? (
                            <ResponsiveContainer width="100%" height={300}>
                                <PieChart>
                                    <Pie
                                        data={operationData}
                                        cx="50%"
                                        cy="50%"
                                        labelLine={false}
                                        label={({ name, percent }) => `${name} ${(percent as any * 100).toFixed(0)}%`}
                                        outerRadius={80}
                                        fill="#8884d8"
                                        dataKey="value"
                                    >
                                        {operationData.map((_entry, index) => (
                                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                        ))}
                                    </Pie>
                                    <Tooltip />
                                    <Legend />
                                </PieChart>
                            </ResponsiveContainer>
                        ) : (
                            <div className="h-[300px] flex items-center justify-center text-muted-foreground">
                                No operation data available
                            </div>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Request Status</CardTitle>
                        <CardDescription>Success vs Failed requests</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <ResponsiveContainer width="100%" height={300}>
                            <BarChart
                                data={[
                                    { name: 'Successful', count: stats.successful_requests },
                                    { name: 'Failed', count: stats.failed_requests },
                                ]}
                            >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis />
                                <Tooltip />
                                <Bar dataKey="count" fill="#8884d8" />
                            </BarChart>
                        </ResponsiveContainer>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}