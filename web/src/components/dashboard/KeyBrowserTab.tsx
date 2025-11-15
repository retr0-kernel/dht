import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { kvAPI } from '@/lib/api';
import { Search, Plus, Trash2, Loader2, Key as KeyIcon, RefreshCw, List, ChevronDown, ChevronRight } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { formatDate } from '@/lib/utils';

export function KeyBrowserTab() {
    const [apiKeyInput, setApiKeyInput] = useState('');
    const [isApiKeySet, setIsApiKeySet] = useState(false);
    const [searchKey, setSearchKey] = useState('');
    const [keyValue, setKeyValue] = useState<any>(null);
    const [loading, setLoading] = useState(false);
    const [showPutDialog, setShowPutDialog] = useState(false);
    const [newKey, setNewKey] = useState('');
    const [newValue, setNewValue] = useState('');
    const [newTTL, setNewTTL] = useState('');
    const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set());
    const [viewingKeyData, setViewingKeyData] = useState<{[key: string]: any}>({});

    // New state for list view
    const [keysList, setKeysList] = useState<any[]>([]);
    const [loadingList, setLoadingList] = useState(false);

    // Check if we have a stored API key
    useEffect(() => {
        const storedKey = localStorage.getItem('browser_api_key');
        if (storedKey) {
            setApiKeyInput(storedKey);
            setIsApiKeySet(true);
            // Auto-load keys on mount
            loadKeysList(storedKey);
        }
    }, []);

    const loadKeysList = async (apiKey?: string) => {
        const keyToUse = apiKey || apiKeyInput;
        if (!keyToUse) return;

        setLoadingList(true);
        try {
            const data = await kvAPI.list(keyToUse);
            setKeysList(data.keys || []);
        } catch (error: any) {
            console.error('Failed to load keys:', error);
            if (error.response?.status === 401) {
                handleClearApiKey();
            }
        } finally {
            setLoadingList(false);
        }
    };

    const handleSetApiKey = () => {
        if (!apiKeyInput.trim()) return;

        if (!apiKeyInput.startsWith('ydht_')) {
            alert('Invalid API key format. Key should start with "ydht_"');
            return;
        }

        localStorage.setItem('browser_api_key', apiKeyInput);
        setIsApiKeySet(true);
        loadKeysList(apiKeyInput);
    };

    const handleClearApiKey = () => {
        localStorage.removeItem('browser_api_key');
        setApiKeyInput('');
        setIsApiKeySet(false);
        setKeyValue(null);
        setSearchKey('');
        setKeysList([]);
    };

    const handleSearch = async () => {
        if (!searchKey.trim() || !apiKeyInput) return;

        setLoading(true);
        setKeyValue(null);
        try {
            const data = await kvAPI.get(searchKey, apiKeyInput);
            setKeyValue(data);
        } catch (error: any) {
            if (error.response?.status === 404) {
                setKeyValue({ error: 'Key not found' });
            } else if (error.response?.status === 401) {
                setKeyValue({ error: 'Invalid API key or unauthorized' });
                handleClearApiKey();
            } else {
                setKeyValue({ error: 'Failed to fetch key' });
            }
        } finally {
            setLoading(false);
        }
    };

    const handleKeyClick = async (key: string) => {
        setSearchKey(key);
        setLoading(true);
        setKeyValue(null);
        try {
            const data = await kvAPI.get(key, apiKeyInput);
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
        if (!newKey.trim() || !newValue.trim() || !apiKeyInput) return;

        setLoading(true);
        try {
            let parsedValue;
            try {
                parsedValue = JSON.parse(newValue);
            } catch {
                parsedValue = newValue;
            }

            await kvAPI.put(newKey, parsedValue, apiKeyInput, newTTL || undefined);
            setShowPutDialog(false);
            setNewKey('');
            setNewValue('');
            setNewTTL('');
            setSearchKey(newKey);

            // Reload the list
            await loadKeysList();

            // Automatically search for the newly created key
            setTimeout(() => {
                handleKeyClick(newKey);
            }, 500);
        } catch (error: any) {
            if (error.response?.status === 401) {
                alert('Invalid API key or unauthorized');
                handleClearApiKey();
            } else {
                alert('Failed to create key: ' + (error.response?.data?.error || error.message));
            }
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!searchKey.trim() || !apiKeyInput) return;
        if (!confirm(`Are you sure you want to delete key "${searchKey}"?`)) return;

        setLoading(true);
        try {
            await kvAPI.delete(searchKey, apiKeyInput);
            setKeyValue(null);
            setSearchKey('');

            // Reload the list
            await loadKeysList();
        } catch (error: any) {
            if (error.response?.status === 401) {
                alert('Invalid API key or unauthorized');
                handleClearApiKey();
            } else {
                alert('Failed to delete key: ' + (error.response?.data?.error || error.message));
            }
        } finally {
            setLoading(false);
        }
    };

    const toggleKeyExpansion = async (key: string) => {
        const newExpanded = new Set(expandedKeys);
        if (newExpanded.has(key)) {
            newExpanded.delete(key);
        } else {
            newExpanded.add(key);
            // Load the key data if not already loaded
            if (!viewingKeyData[key]) {
                try {
                    const data = await kvAPI.get(key, apiKeyInput);
                    setViewingKeyData(prev => ({ ...prev, [key]: data }));
                } catch (error: any) {
                    if (error.response?.status === 404) {
                        setViewingKeyData(prev => ({ ...prev, [key]: { error: 'Key not found' } }));
                    } else {
                        setViewingKeyData(prev => ({ ...prev, [key]: { error: 'Failed to fetch key' } }));
                    }
                }
            }
        }
        setExpandedKeys(newExpanded);
    };

    const handleDeleteKey = async (key: string) => {
        if (!apiKeyInput) return;

        setLoading(true);
        try {
            await kvAPI.delete(key, apiKeyInput);

            // Remove from expanded keys and viewing data
            const newExpanded = new Set(expandedKeys);
            newExpanded.delete(key);
            setExpandedKeys(newExpanded);

            const newViewingData = { ...viewingKeyData };
            delete newViewingData[key];
            setViewingKeyData(newViewingData);

            // Reload the list
            await loadKeysList();
        } catch (error: any) {
            if (error.response?.status === 401) {
                alert('Invalid API key or unauthorized');
                handleClearApiKey();
            } else {
                alert('Failed to delete key: ' + (error.response?.data?.error || error.message));
            }
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
                    {!isApiKeySet ? (
                        <div className="space-y-4 p-6 border rounded-lg bg-muted/50">
                            <div className="flex items-center justify-center mb-4">
                                <KeyIcon className="h-12 w-12 text-muted-foreground" />
                            </div>
                            <div className="text-center space-y-2 mb-4">
                                <h3 className="font-semibold">Enter Your API Key</h3>
                                <p className="text-sm text-muted-foreground">
                                    You need to enter your API key to browse and manage keys
                                </p>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="apiKeyInput">API Key</Label>
                                <Input
                                    id="apiKeyInput"
                                    type="password"
                                    placeholder="ydht_..."
                                    value={apiKeyInput}
                                    onChange={(e) => setApiKeyInput(e.target.value)}
                                    onKeyDown={(e) => e.key === 'Enter' && handleSetApiKey()}
                                />
                                <p className="text-xs text-muted-foreground">
                                    Go to the API Keys tab to create a new key if you don't have one
                                </p>
                            </div>
                            <Button onClick={handleSetApiKey} className="w-full" disabled={!apiKeyInput.trim()}>
                                Set API Key
                            </Button>
                        </div>
                    ) : (
                        <>
                            {/* API Key Display */}
                            <div className="flex items-center justify-between p-3 bg-muted rounded-lg">
                                <div className="flex items-center gap-2">
                                    <KeyIcon className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm font-mono">
                                        {apiKeyInput.substring(0, 15)}...
                                    </span>
                                </div>
                                <div className="flex gap-2">
                                    <Button variant="outline" size="sm" onClick={() => loadKeysList()}>
                                        <RefreshCw className="h-4 w-4 mr-2" />
                                        Refresh
                                    </Button>
                                    <Button variant="outline" size="sm" onClick={handleClearApiKey}>
                                        Change Key
                                    </Button>
                                </div>
                            </div>

                            <Tabs defaultValue="list" className="w-full">
                                <TabsList className="grid w-full grid-cols-2">
                                    <TabsTrigger value="list">
                                        <List className="h-4 w-4 mr-2" />
                                        All Keys
                                    </TabsTrigger>
                                    <TabsTrigger value="search">
                                        <Search className="h-4 w-4 mr-2" />
                                        Search
                                    </TabsTrigger>
                                </TabsList>

                                <TabsContent value="list" className="space-y-4">
                                    <div className="flex justify-between items-center">
                                        <p className="text-sm text-muted-foreground">
                                            {keysList.length} key{keysList.length !== 1 ? 's' : ''} found
                                        </p>
                                        <Button onClick={() => setShowPutDialog(true)}>
                                            <Plus className="h-4 w-4 mr-2" />
                                            New Key
                                        </Button>
                                    </div>

                                    {loadingList ? (
                                        <div className="flex items-center justify-center py-8">
                                            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                                        </div>
                                    ) : keysList.length === 0 ? (
                                        <div className="text-center py-12 border rounded-lg">
                                            <p className="text-muted-foreground mb-4">No keys found</p>
                                            <Button onClick={() => setShowPutDialog(true)}>
                                                <Plus className="h-4 w-4 mr-2" />
                                                Create your first key
                                            </Button>
                                        </div>
                                    ) : (
                                        <div className="border rounded-lg">
                                            <Table>
                                                <TableHeader>
                                                    <TableRow>
                                                        <TableHead className="w-[40px]"></TableHead>
                                                        <TableHead>Key</TableHead>
                                                        <TableHead>Created</TableHead>
                                                        <TableHead>TTL</TableHead>
                                                        <TableHead className="w-[100px]">Actions</TableHead>
                                                    </TableRow>
                                                </TableHeader>
                                                <TableBody>
                                                    {keysList.map((keyInfo) => (
                                                        <>
                                                            <TableRow key={keyInfo.key} className="hover:bg-muted/50">
                                                                <TableCell
                                                                    className="cursor-pointer"
                                                                    onClick={() => toggleKeyExpansion(keyInfo.key)}
                                                                >
                                                                    {expandedKeys.has(keyInfo.key) ? (
                                                                        <ChevronDown className="h-4 w-4" />
                                                                    ) : (
                                                                        <ChevronRight className="h-4 w-4" />
                                                                    )}
                                                                </TableCell>
                                                                <TableCell
                                                                    className="font-mono text-sm cursor-pointer"
                                                                    onClick={() => toggleKeyExpansion(keyInfo.key)}
                                                                >
                                                                    {keyInfo.key}
                                                                </TableCell>
                                                                <TableCell
                                                                    className="text-sm text-muted-foreground cursor-pointer"
                                                                    onClick={() => toggleKeyExpansion(keyInfo.key)}
                                                                >
                                                                    {formatDate(keyInfo.created_at)}
                                                                </TableCell>
                                                                <TableCell
                                                                    className="text-sm cursor-pointer"
                                                                    onClick={() => toggleKeyExpansion(keyInfo.key)}
                                                                >
                                                                    {keyInfo.has_ttl ? (
                                                                        <span className="text-orange-600 dark:text-orange-400">
                                                                            Expires
                                                                        </span>
                                                                    ) : (
                                                                        <span className="text-muted-foreground">
                                                                            No expiry
                                                                        </span>
                                                                    )}
                                                                </TableCell>
                                                                <TableCell>
                                                                    <Button
                                                                        variant="ghost"
                                                                        size="sm"
                                                                        onClick={(e) => {
                                                                            e.stopPropagation();
                                                                            if (confirm(`Are you sure you want to delete key "${keyInfo.key}"?`)) {
                                                                                handleDeleteKey(keyInfo.key);
                                                                            }
                                                                        }}
                                                                    >
                                                                        <Trash2 className="h-4 w-4 text-destructive" />
                                                                    </Button>
                                                                </TableCell>
                                                            </TableRow>
                                                            {expandedKeys.has(keyInfo.key) && (
                                                                <TableRow key={`${keyInfo.key}-expanded`}>
                                                                    <TableCell colSpan={5} className="bg-muted/30 p-4">
                                                                        <div className="space-y-3">
                                                                            <div className="flex justify-between items-center">
                                                                                <h4 className="font-semibold text-sm">Key Details</h4>
                                                                                {!viewingKeyData[keyInfo.key] && (
                                                                                    <span className="text-xs text-muted-foreground">
                                                                                        Loading value...
                                                                                    </span>
                                                                                )}
                                                                            </div>
                                                                            <div className="grid grid-cols-2 gap-2 text-sm">
                                                                                <div>
                                                                                    <span className="font-medium">Key:</span>{' '}
                                                                                    <span className="font-mono">{keyInfo.key}</span>
                                                                                </div>
                                                                                <div>
                                                                                    <span className="font-medium">Created:</span>{' '}
                                                                                    {formatDate(keyInfo.created_at)}
                                                                                </div>
                                                                                <div>
                                                                                    <span className="font-medium">TTL:</span>{' '}
                                                                                    {keyInfo.has_ttl ? (
                                                                                        <span className="text-orange-600 dark:text-orange-400">
                                                                                            Expires
                                                                                        </span>
                                                                                    ) : (
                                                                                        'No expiry'
                                                                                    )}
                                                                                </div>
                                                                            </div>
                                                                            {viewingKeyData[keyInfo.key] && (
                                                                                <div className="mt-3 border-t pt-3">
                                                                                    <h5 className="font-medium text-sm mb-2">Value:</h5>
                                                                                    {viewingKeyData[keyInfo.key].error ? (
                                                                                        <div className="text-sm text-muted-foreground p-3 bg-background rounded-md border">
                                                                                            {viewingKeyData[keyInfo.key].error}
                                                                                        </div>
                                                                                    ) : (
                                                                                        <pre className="bg-background p-3 rounded-md overflow-auto max-h-64 text-xs border font-mono">
{JSON.stringify(viewingKeyData[keyInfo.key], null, 2)}
                                                                                        </pre>
                                                                                    )}
                                                                                </div>
                                                                            )}
                                                                        </div>
                                                                    </TableCell>
                                                                </TableRow>
                                                            )}
                                                        </>
                                                    ))}
                                                </TableBody>
                                            </Table>
                                        </div>
                                    )}
                                </TabsContent>

                                <TabsContent value="search" className="space-y-4">
                                    <div className="flex gap-2">
                                        <div className="flex-1">
                                            <Input
                                                placeholder="Enter key to search (e.g., user:123)..."
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

                                    {!keyValue && !loading && (
                                        <div className="text-center py-8 text-sm text-muted-foreground">
                                            <p>Enter a key name above and click search to view its value</p>
                                        </div>
                                    )}
                                </TabsContent>
                            </Tabs>

                            {/* Results (shown in both tabs) */}
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
                                                    <Button variant="destructive" size="icon" onClick={handleDelete}>
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            )}
                                        </div>
                                    </CardHeader>
                                    <CardContent>
                                        {keyValue.error ? (
                                            <div className="text-sm text-muted-foreground p-4 bg-muted rounded-md">
                                                {keyValue.error}
                                            </div>
                                        ) : (
                                            <pre className="bg-muted p-4 rounded-md overflow-auto max-h-96 text-sm">
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
                        <DialogTitle>Create New Key-Value Pair</DialogTitle>
                        <DialogDescription>
                            Store a new key-value pair in the distributed store
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="newKey">Key</Label>
                            <Input
                                id="newKey"
                                placeholder="user:123 or session:abc or any-key-name"
                                value={newKey}
                                onChange={(e) => setNewKey(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                Use descriptive names like "user:123" or "config:app"
                            </p>
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="newValue">Value (JSON or text)</Label>
                            <textarea
                                id="newValue"
                                className="w-full h-32 px-3 py-2 text-sm rounded-md border border-input bg-background resize-none font-mono"
                                placeholder='{"name": "John", "age": 30}'
                                value={newValue}
                                onChange={(e) => setNewValue(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                Enter JSON object or plain text
                            </p>
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="ttl">TTL - Time to Live (optional)</Label>
                            <Input
                                id="ttl"
                                placeholder="1h, 30m, 24h, 7d"
                                value={newTTL}
                                onChange={(e) => setNewTTL(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                Leave empty for no expiration. Examples: 1h (1 hour), 30m (30 minutes), 24h (1 day)
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
                                'Create Key'
                            )}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}