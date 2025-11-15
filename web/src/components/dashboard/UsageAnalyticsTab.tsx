import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table';
import { usageAPI, type UsageRecord } from '../../lib/api';
import { formatDate, formatBytes } from '../../lib/utils';
import { Loader2, TrendingUp, TrendingDown, Activity } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { format } from 'date-fns';

export function UsageAnalyticsTab() {
    const [records, setRecords] = useState<UsageRecord[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadUsageRecords();
    }, []);

    const loadUsageRecords = async () => {
        try {
            const data = await usageAPI.list({ limit: 100 });
            setRecords(data);
        } catch (error) {
            console.error('Failed to load usage records:', error);
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

    // Prepare chart data - group by hour
    const chartData = records
        .slice(0, 50)
        .reverse()
        .map((record) => ({
            time: format(new Date(record.created_at), 'HH:mm'),
            latency: record.duration_ms,
            success: record.status_code >= 200 && record.status_code < 300 ? 1 : 0,
        }));

    const totalRequests = records.length;
    const successfulRequests = records.filter(r => r.status_code >= 200 && r.status_code < 300).length;
    const avgLatency = records.length > 0
        ? records.reduce((sum, r) => sum + r.duration_ms, 0) / records.length
        : 0;

    return (
        <div className="space-y-4">
            {/* Summary Cards */}
            <div className="grid gap-4 md:grid-cols-3">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
                        <Activity className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{totalRequests}</div>
                        <p className="text-xs text-muted-foreground">Last 100 requests</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
                        <TrendingUp className="h-4 w-4 text-green-600" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {totalRequests > 0 ? ((successfulRequests / totalRequests) * 100).toFixed(1) : 0}%
                        </div>
                        <p className="text-xs text-muted-foreground">
                            {successfulRequests} / {totalRequests} successful
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
                        <TrendingDown className="h-4 w-4 text-blue-600" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{avgLatency.toFixed(0)}ms</div>
                        <p className="text-xs text-muted-foreground">Average response time</p>
                    </CardContent>
                </Card>
            </div>

            {/* Latency Chart */}
            <Card>
                <CardHeader>
                    <CardTitle>Response Time</CardTitle>
                    <CardDescription>Request latency over time</CardDescription>
                </CardHeader>
                <CardContent>
                    <ResponsiveContainer width="100%" height={300}>
                        <LineChart data={chartData}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="time" />
                            <YAxis />
                            <Tooltip />
                            <Line
                                type="monotone"
                                dataKey="latency"
                                stroke="#8884d8"
                                strokeWidth={2}
                                dot={false}
                            />
                        </LineChart>
                    </ResponsiveContainer>
                </CardContent>
            </Card>

            {/* Recent Requests Table */}
            <Card>
                <CardHeader>
                    <CardTitle>Recent Requests</CardTitle>
                    <CardDescription>Your latest API requests</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Timestamp</TableHead>
                                <TableHead>Operation</TableHead>
                                <TableHead>Key</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Latency</TableHead>
                                <TableHead>Size</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {records.slice(0, 20).map((record) => (
                                <TableRow key={record.id}>
                                    <TableCell className="text-sm">
                                        {formatDate(record.created_at)}
                                    </TableCell>
                                    <TableCell>
                    <span className="text-xs font-mono bg-secondary px-2 py-1 rounded">
                      {record.operation}
                    </span>
                                    </TableCell>
                                    <TableCell className="font-mono text-sm">
                                        {record.key_accessed || '-'}
                                    </TableCell>
                                    <TableCell>
                    <span
                        className={`text-xs px-2 py-1 rounded ${
                            record.status_code >= 200 && record.status_code < 300
                                ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
                                : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-100'
                        }`}
                    >
                      {record.status_code}
                    </span>
                                    </TableCell>
                                    <TableCell className="text-sm">
                                        {record.duration_ms}ms
                                    </TableCell>
                                    <TableCell className="text-sm text-muted-foreground">
                                        {formatBytes(record.request_size_bytes + record.response_size_bytes)}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}