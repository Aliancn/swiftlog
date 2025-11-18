'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { useRouter } from 'next/navigation';

interface Token {
  id: string;
  name: string;
  created_at: string;
}

export default function TokensPage() {
  const router = useRouter();
  const [tokens, setTokens] = useState<Token[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [newTokenName, setNewTokenName] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [newTokenValue, setNewTokenValue] = useState('');

  useEffect(() => {
    loadTokens();
  }, []);

  const loadTokens = async () => {
    try {
      const response = await api.listTokens();
      setTokens(response.tokens || []);
    } catch (err) {
      if (err instanceof Error && err.message.includes('401')) {
        router.push('/login');
      } else {
        setError('Failed to load tokens');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateToken = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsCreating(true);
    setError('');

    try {
      const response = await api.createToken(newTokenName);
      setNewTokenValue(response.token);
      setNewTokenName('');
      await loadTokens();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create token');
    } finally {
      setIsCreating(false);
    }
  };

  const handleDeleteToken = async (id: string) => {
    if (!confirm('Are you sure you want to delete this token?')) {
      return;
    }

    try {
      await api.deleteToken(id);
      await loadTokens();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete token');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading tokens...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">API Tokens</h1>
          <p className="mt-2 text-sm text-gray-600">
            Manage your API tokens for CLI and programmatic access
          </p>
        </div>

        {error && (
          <div className="mb-4 bg-red-50 border border-red-200 rounded-md p-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {newTokenValue && (
          <div className="mb-6 bg-green-50 border border-green-200 rounded-md p-4">
            <h3 className="text-sm font-medium text-green-900 mb-2">
              Token created successfully!
            </h3>
            <p className="text-xs text-green-700 mb-2">
              Make sure to copy your token now. You won't be able to see it again!
            </p>
            <div className="flex items-center space-x-2">
              <code className="flex-1 bg-white p-2 rounded border border-green-300 text-sm font-mono overflow-x-auto">
                {newTokenValue}
              </code>
              <button
                onClick={() => copyToClipboard(newTokenValue)}
                className="px-3 py-2 bg-green-600 text-white text-sm rounded hover:bg-green-700"
              >
                Copy
              </button>
              <button
                onClick={() => setNewTokenValue('')}
                className="px-3 py-2 bg-gray-600 text-white text-sm rounded hover:bg-gray-700"
              >
                Close
              </button>
            </div>
          </div>
        )}

        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Create New Token</h2>
          <form onSubmit={handleCreateToken} className="flex gap-4">
            <input
              type="text"
              placeholder="Token name (e.g., 'Production Server', 'My Laptop')"
              value={newTokenName}
              onChange={(e) => setNewTokenName(e.target.value)}
              required
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              type="submit"
              disabled={isCreating}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {isCreating ? 'Creating...' : 'Create Token'}
            </button>
          </form>
        </div>

        <div className="bg-white shadow rounded-lg overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {tokens.length === 0 ? (
                <tr>
                  <td colSpan={3} className="px-6 py-4 text-center text-sm text-gray-500">
                    No tokens yet. Create one to get started.
                  </td>
                </tr>
              ) : (
                tokens.map((token) => (
                  <tr key={token.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">{token.name}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-500">
                        {new Date(token.created_at).toLocaleString()}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => handleDeleteToken(token.id)}
                        className="text-red-600 hover:text-red-900"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
