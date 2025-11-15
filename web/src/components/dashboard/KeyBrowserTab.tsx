import { useState, useEffect, useCallback } from 'react';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { kvAPI, apiKeyAPI, type APIKey } from '../../lib/api';
import { Search, Plus, Trash2, Edit2, Loader2, Database } from 'lucide-react';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '../ui/dialog';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '../ui/select';

export function KeyBrowserTab() {
    const [selectedApiKey, setSelectedApiKey] = useState<string>('');
    const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
    const [searchKey, setSearchKey] = useState('');
    const [keyValue, setKeyValue] = useState<any>(null);
    const [loading, setLoading] = useState(false);
    const [showPutDialog, setShowPutDialog] = useState(false);
    const [newKey, setNewKey] = useState('');
    const [newValue, setNewValue] = useState('');
    const [newTTL, setNewTTL] = useState('');

    const loadApiKeys = useCallback(async () => {
        try {
            const response = await apiKeyAPI.list();
            const activeKeys = (response?.api_keys || [])
                .filter(k => k.is_active && k.key); // Only include keys that have a valid key value
            setApiKeys(activeKeys);
            if (activeKeys.length > 0 && activeKeys[0].key) {
                setSelectedApiKey(activeKeys[0].key);
            }
        } catch (error) {
            console.error('Failed to load API keys:', error);
            setApiKeys([]);
        }
    }, []);

    useEffect(() => {
        loadApiKeys();
    }, [loadApiKeys]);

    const handleSearch = async () => {
        if (!searchKey.trim() || !selectedApiKey) return;

        setLoading(true);
        setKeyValue(null);
        try {
            const data = await kvAPI.get(searchKey, selectedApiKey);
            setKeyValue(data);
        } catch (error: any) {
            if (error.response?.status === 404) {
                setKeyValue({ error: 'Key not found' });
            } else {
                setKeyValue({ error: 'Failed to fetch key' });
            }
        } finally {
            setLoading(false);
        }
    };

    const handlePut = async () => {
        if (!newKey.trim() || !newValue.trim() || !selectedApiKey) return;

        setLoading(true);
        try {
            let parsedValue;
            try {
                parsedValue = JSON.parse(newValue);
            } catch {
                parsedValue = newValue;
            }

            await kvAPI.put(newKey, parsedValue, selectedApiKey, newTTL || undefined);
            setShowPutDialog(false);
            setNewKey('');
            setNewValue('');
            setNewTTL('');
            setSearchKey(newKey);
            handleSearch();
        } catch (error) {
            console.error('Failed to put key:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!searchKey.trim() || !selectedApiKey) return;
        if (!confirm(`Are you sure you want to delete key "${searchKey}"?`)) return;

        setLoading(true);
        try {
            await kvAPI.delete(searchKey, selectedApiKey);
            setKeyValue(null);
            setSearchKey('');
        } catch (error) {
            console.error('Failed to delete key:', error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <>
            <Card>
                <CardHeader>
                    <CardTitle>Key Browser</CardTitle>
                    <CardDescription>
                        Browse and manage your key-value pairs
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    {/* API Key Selection */}
                    {apiKeys.length === 0 ? (
                        <div className="text-center py-8">
                            <Database className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                            <p className="text-muted-foreground mb-4">
                                You need an active API key to browse keys
                            </p>
                            <p className="text-sm text-muted-foreground">
                                Go to the API Keys tab to create one
                            </p>
                        </div>
                    ) : (
                        <>
                            <div className="space-y-2">
                                <Label>Select API Key</Label>
                                <Select value={selectedApiKey} onValueChange={setSelectedApiKey}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select an API key" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {apiKeys.map((key) => (
                                            <SelectItem key={key.id} value={key.key!}>
                                                {key.name} ({key.key_prefix}...)
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Search Bar */}
                            <div className="flex gap-2">
                                <div className="flex-1">
                                    <Input
                                        placeholder="Enter key to search..."
                                        value={searchKey}
                                        onChange={(e) => setSearchKey(e.target.value)}
                                        onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                                    />
                                </div>
                                <Button onClick={handleSearch} disabled={loading || !searchKey.trim()}>
                                    {loading ? (
                                        <Loader2 className="h-4 w-4 animate-spin" />
                                    ) : (
                                        <Search className="h-4 w-4" />
                                    )}
                                </Button>
                                <Button onClick={() => setShowPutDialog(true)}>
                                    <Plus className="h-4 w-4" />
                                </Button>
                            </div>

                            {/* Results */}
                            {keyValue && (
                                <Card>
                                    <CardHeader>
                                        <div className="flex items-center justify-between">
                                            <div>
                                                <CardTitle className="text-lg font-mono">{searchKey}</CardTitle>
                                                <CardDescription>Key value</CardDescription>
                                            </div>
                                            {!keyValue.error && (
                                                <div className="flex gap-2">
                                                    <Button variant="outline" size="icon" disabled>
                                                        <Edit2 className="h-4 w-4" />
                                                    </Button>
                                                    <Button variant="destructive" size="icon" onClick={handleDelete}>
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            )}
                                        </div>
                                    </CardHeader>
                                    <CardContent>
                                        {keyValue.error ? (
                                            <p className="text-muted-foreground">{keyValue.error}</p>
                                        ) : (
                                            <pre className="bg-muted p-4 rounded-md overflow-auto max-h-96">
                        <code>{JSON.stringify(keyValue, null, 2)}</code>
                      </pre>
                                        )}
                                    </CardContent>
                                </Card>
                            )}
                        </>
                    )}
                </CardContent>
            </Card>

            {/* Put Key Dialog */}
            <Dialog open={showPutDialog} onOpenChange={setShowPutDialog}>
                <DialogContent className="max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>Add New Key</DialogTitle>
                        <DialogDescription>
                            Create a new key-value pair in the distributed store
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="newKey">Key</Label>
                            <Input
                                id="newKey"
                                placeholder="user:123"
                                value={newKey}
                                onChange={(e) => setNewKey(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="newValue">Value (JSON)</Label>
                            <textarea
                                id="newValue"
                                className="w-full h-32 px-3 py-2 text-sm rounded-md border border-input bg-background"
                                placeholder='{"name": "John", "age": 30}'
                                value={newValue}
                                onChange={(e) => setNewValue(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="ttl">TTL (optional)</Label>
                            <Input
                                id="ttl"
                                placeholder="1h, 30m, 24h"
                                value={newTTL}
                                onChange={(e) => setNewTTL(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                Leave empty for no expiration
                            </p>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowPutDialog(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handlePut} disabled={loading || !newKey.trim() || !newValue.trim()}>
                            {loading ? (
                                <>
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                    Saving...
                                </>
                            ) : (
                                'Save'
                            )}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}