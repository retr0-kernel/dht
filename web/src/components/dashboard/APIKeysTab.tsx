import { useEffect, useState, useCallback } from 'react';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { apiKeyAPI, type APIKey } from '../../lib/api';
import { formatDate } from '../../lib/utils';
import { Plus, Copy, Trash2, Loader2, Check, AlertCircle } from 'lucide-react';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '../ui/dialog';

interface APIKeysTabProps {
    onUpdate?: () => void;
}

export function APIKeysTab({ onUpdate }: APIKeysTabProps) {
    const [keys, setKeys] = useState<APIKey[]>([]);
    const [loading, setLoading] = useState(true);
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [showKeyDialog, setShowKeyDialog] = useState(false);
    const [newKeyName, setNewKeyName] = useState('');
    const [newKeyExpiry, setNewKeyExpiry] = useState('90');
    const [createdKey, setCreatedKey] = useState<APIKey | null>(null);
    const [creating, setCreating] = useState(false);
    const [copiedKey, setCopiedKey] = useState(false);

    const loadKeys = useCallback(async () => {
        try {
            const response = await apiKeyAPI.list();
            setKeys(response?.api_keys || []);
            onUpdate?.();
        } catch (error) {
            console.error('Failed to load API keys:', error);
        } finally {
            setLoading(false);
        }
    }, [onUpdate]);

    useEffect(() => {
        loadKeys();
    }, [loadKeys]);

    const handleCreateKey = async () => {
        if (!newKeyName.trim()) return;

        setCreating(true);
        try {
            const key = await apiKeyAPI.create({
                name: newKeyName,
                scopes: ['read', 'write'],
                expires_in_days: parseInt(newKeyExpiry) || undefined,
            });
            setCreatedKey(key);
            setShowCreateDialog(false);
            setShowKeyDialog(true);
            setNewKeyName('');
            setNewKeyExpiry('90');
            loadKeys();
        } catch (error) {
            console.error('Failed to create API key:', error);
        } finally {
            setCreating(false);
        }
    };

    const handleCopyKey = () => {
        if (createdKey?.key) {
            navigator.clipboard.writeText(createdKey.key);
            setCopiedKey(true);
            setTimeout(() => setCopiedKey(false), 2000);
        }
    };

    const handleDeleteKey = async (id: number) => {
        if (!confirm('Are you sure you want to delete this API key?')) return;

        try {
            await apiKeyAPI.delete(id);
            loadKeys();
        } catch (error) {
            console.error('Failed to delete API key:', error);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <>
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>API Keys</CardTitle>
                            <CardDescription>
                                Manage your API keys for programmatic access
                            </CardDescription>
                        </div>
                        <Button onClick={() => setShowCreateDialog(true)}>
                            <Plus className="h-4 w-4 mr-2" />
                            Create Key
                        </Button>
                    </div>
                </CardHeader>
                <CardContent>
                    {keys.length === 0 ? (
                        <div className="text-center py-12">
                            <p className="text-muted-foreground mb-4">No API keys yet</p>
                            <Button onClick={() => setShowCreateDialog(true)}>
                                <Plus className="h-4 w-4 mr-2" />
                                Create your first API key
                            </Button>
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Key Prefix</TableHead>
                                    <TableHead>Scopes</TableHead>
                                    <TableHead>Last Used</TableHead>
                                    <TableHead>Expires</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead className="w-[100px]">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {keys.map((key) => (
                                    <TableRow key={key.id}>
                                        <TableCell className="font-medium">{key.name}</TableCell>
                                        <TableCell>
                                            <code className="text-xs bg-muted px-2 py-1 rounded">
                                                {key.key_prefix}...
                                            </code>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex gap-1">
                                                {key.scopes.map((scope) => (
                                                    <span
                                                        key={scope}
                                                        className="text-xs bg-secondary px-2 py-1 rounded"
                                                    >
                            {scope}
                          </span>
                                                ))}
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-sm text-muted-foreground">
                                            {key.last_used_at ? formatDate(key.last_used_at) : 'Never'}
                                        </TableCell>
                                        <TableCell className="text-sm text-muted-foreground">
                                            {key.expires_at ? formatDate(key.expires_at) : 'Never'}
                                        </TableCell>
                                        <TableCell>
                      <span
                          className={`text-xs px-2 py-1 rounded ${
                              key.is_active
                                  ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
                                  : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-100'
                          }`}
                      >
                        {key.is_active ? 'Active' : 'Inactive'}
                      </span>
                                        </TableCell>
                                        <TableCell>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                onClick={() => handleDeleteKey(key.id)}
                                            >
                                                <Trash2 className="h-4 w-4 text-destructive" />
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>

            {/* Create Key Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create API Key</DialogTitle>
                        <DialogDescription>
                            Create a new API key for programmatic access to your data
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="keyName">Key Name</Label>
                            <Input
                                id="keyName"
                                placeholder="Production Key"
                                value={newKeyName}
                                onChange={(e) => setNewKeyName(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="expiry">Expires In (days)</Label>
                            <Input
                                id="expiry"
                                type="number"
                                placeholder="90"
                                value={newKeyExpiry}
                                onChange={(e) => setNewKeyExpiry(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                Leave empty for no expiration
                            </p>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleCreateKey} disabled={creating || !newKeyName.trim()}>
                            {creating ? (
                                <>
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                    Creating...
                                </>
                            ) : (
                                'Create Key'
                            )}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Show Created Key Dialog */}
            <Dialog open={showKeyDialog} onOpenChange={setShowKeyDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>API Key Created</DialogTitle>
                        <DialogDescription>
                            Save this key securely. You won't be able to see it again!
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="flex items-center gap-2 p-3 bg-muted rounded-md">
                            <code className="flex-1 text-sm break-all">
                                {createdKey?.key}
                            </code>
                            <Button
                                variant="ghost"
                                size="icon"
                                onClick={handleCopyKey}
                            >
                                {copiedKey ? (
                                    <Check className="h-4 w-4 text-green-600" />
                                ) : (
                                    <Copy className="h-4 w-4" />
                                )}
                            </Button>
                        </div>
                        <div className="flex items-start gap-2 p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-md">
                            <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-500 mt-0.5" />
                            <p className="text-sm text-yellow-800 dark:text-yellow-200">
                                Make sure to copy your API key now. You won't be able to see it again!
                            </p>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button onClick={() => setShowKeyDialog(false)}>Done</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}