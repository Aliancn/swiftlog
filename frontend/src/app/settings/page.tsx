'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useUserSettings } from '@/lib/hooks';
import { api } from '@/lib/api';
import { TruncateStrategy } from '@/types';

export default function SettingsPage() {
  const router = useRouter();
  const { data, error, isLoading, mutate } = useUserSettings();
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState('');
  const [saveSuccess, setSaveSuccess] = useState(false);

  // Form state
  const [formData, setFormData] = useState({
    ai_enabled: true,
    ai_base_url: '',
    ai_api_key: '',
    ai_model: '',
    ai_max_tokens: 500,
    ai_auto_analyze: false,
    ai_max_log_lines: 1000,
    ai_log_truncate_strategy: TruncateStrategy.Tail,
    ai_system_prompt: '',
  });

  const [showApiKey, setShowApiKey] = useState(false);
  const [apiKeyChanged, setApiKeyChanged] = useState(false);

  // Initialize form when data loads
  useEffect(() => {
    if (data?.settings) {
      setFormData({
        ai_enabled: data.settings.ai_enabled,
        ai_base_url: data.settings.ai_base_url,
        ai_api_key: '',
        ai_model: data.settings.ai_model,
        ai_max_tokens: data.settings.ai_max_tokens,
        ai_auto_analyze: data.settings.ai_auto_analyze,
        ai_max_log_lines: data.settings.ai_max_log_lines,
        ai_log_truncate_strategy: data.settings.ai_log_truncate_strategy,
        ai_system_prompt: data.settings.ai_system_prompt,
      });
    }
  }, [data]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    setSaveError('');
    setSaveSuccess(false);

    try {
      await api.updateUserSettings({
        ...formData,
        ai_api_key: apiKeyChanged ? (formData.ai_api_key || null) : undefined,
      });
      setSaveSuccess(true);
      setApiKeyChanged(false);
      await mutate();
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err: any) {
      setSaveError(err.message || 'Failed to save settings');
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 p-8">
        <div className="max-w-4xl mx-auto">
          <div className="flex justify-center items-center h-64">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 p-8">
        <div className="max-w-4xl mx-auto">
          <div className="bg-red-50 border border-red-200 rounded p-4">
            <p className="text-red-800">Error: {error.message}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">My Settings</h1>
          <p className="text-gray-600 mt-2">
            Configure your personal AI analysis settings. These settings apply to all your projects unless overridden at the project level.
          </p>
        </div>

        {/* Success/Error Messages */}
        {saveSuccess && (
          <div className="mb-6 bg-green-50 border border-green-200 rounded p-4">
            <p className="text-green-800">Settings saved successfully!</p>
          </div>
        )}
        {saveError && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded p-4">
            <p className="text-red-800">{saveError}</p>
          </div>
        )}

        {/* Settings Form */}
        <form onSubmit={handleSave} className="bg-white rounded-lg shadow p-6 space-y-6">
          {/* AI Enabled */}
          <div className="flex items-center justify-between">
            <div>
              <label className="text-sm font-medium text-gray-900">Enable AI Analysis</label>
              <p className="text-sm text-gray-500">Turn AI-powered log analysis on or off globally</p>
            </div>
            <input
              type="checkbox"
              checked={formData.ai_enabled}
              onChange={(e) => setFormData({ ...formData, ai_enabled: e.target.checked })}
              className="h-5 w-5 text-blue-600 rounded focus:ring-blue-500"
            />
          </div>

          <hr className="border-gray-200" />

          {/* AI Base URL */}
          <div>
            <label className="block text-sm font-medium text-gray-900 mb-2">AI API Base URL</label>
            <input
              type="url"
              value={formData.ai_base_url}
              onChange={(e) => setFormData({ ...formData, ai_base_url: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="https://api.openai.com/v1"
              required
            />
            <p className="text-sm text-gray-500 mt-1">OpenAI-compatible API endpoint</p>
          </div>

          {/* AI API Key */}
          <div>
            <label className="block text-sm font-medium text-gray-900 mb-2">
              AI API Key
              {data?.has_api_key && !apiKeyChanged && (
                <span className="ml-2 text-xs text-green-600">● Key configured</span>
              )}
            </label>
            <div className="flex gap-2">
              <input
                type={showApiKey ? 'text' : 'password'}
                value={formData.ai_api_key}
                onChange={(e) => {
                  setFormData({ ...formData, ai_api_key: e.target.value });
                  setApiKeyChanged(true);
                }}
                className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder={data?.has_api_key ? '••••••••••••••••' : 'Enter API key'}
              />
              <button
                type="button"
                onClick={() => setShowApiKey(!showApiKey)}
                className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
              >
                {showApiKey ? 'Hide' : 'Show'}
              </button>
            </div>
            <p className="text-sm text-gray-500 mt-1">Leave empty to keep existing key</p>
          </div>

          {/* AI Model */}
          <div>
            <label className="block text-sm font-medium text-gray-900 mb-2">AI Model</label>
            <input
              type="text"
              value={formData.ai_model}
              onChange={(e) => setFormData({ ...formData, ai_model: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="gpt-4o-mini"
              required
            />
            <p className="text-sm text-gray-500 mt-1">Model name (e.g., gpt-4o-mini, gpt-4)</p>
          </div>

          {/* Max Tokens */}
          <div>
            <label className="block text-sm font-medium text-gray-900 mb-2">Max Tokens</label>
            <input
              type="number"
              value={formData.ai_max_tokens}
              onChange={(e) => setFormData({ ...formData, ai_max_tokens: parseInt(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              min="1"
              required
            />
            <p className="text-sm text-gray-500 mt-1">Maximum tokens for AI response</p>
          </div>

          {/* Auto Analyze */}
          <div className="flex items-center justify-between">
            <div>
              <label className="text-sm font-medium text-gray-900">Auto-Analyze Logs</label>
              <p className="text-sm text-gray-500">Automatically trigger AI analysis when runs complete</p>
            </div>
            <input
              type="checkbox"
              checked={formData.ai_auto_analyze}
              onChange={(e) => setFormData({ ...formData, ai_auto_analyze: e.target.checked })}
              className="h-5 w-5 text-blue-600 rounded focus:ring-blue-500"
            />
          </div>

          <hr className="border-gray-200" />

          {/* Max Log Lines */}
          <div>
            <label className="block text-sm font-medium text-gray-900 mb-2">Max Log Lines</label>
            <input
              type="number"
              value={formData.ai_max_log_lines}
              onChange={(e) => setFormData({ ...formData, ai_max_log_lines: parseInt(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              min="1"
              required
            />
            <p className="text-sm text-gray-500 mt-1">Maximum number of log lines to analyze</p>
          </div>

          {/* Truncate Strategy */}
          <div>
            <label className="block text-sm font-medium text-gray-900 mb-2">Log Truncate Strategy</label>
            <select
              value={formData.ai_log_truncate_strategy}
              onChange={(e) =>
                setFormData({ ...formData, ai_log_truncate_strategy: e.target.value as TruncateStrategy })
              }
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value={TruncateStrategy.Head}>Head - Keep first N lines</option>
              <option value={TruncateStrategy.Tail}>Tail - Keep last N lines</option>
              <option value={TruncateStrategy.Smart}>Smart - Keep head + tail with summary</option>
            </select>
            <p className="text-sm text-gray-500 mt-1">How to handle logs exceeding max lines</p>
          </div>

          {/* System Prompt */}
          <div>
            <label className="block text-sm font-medium text-gray-900 mb-2">AI System Prompt</label>
            <textarea
              value={formData.ai_system_prompt}
              onChange={(e) => setFormData({ ...formData, ai_system_prompt: e.target.value })}
              rows={4}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Instructions for the AI assistant..."
              required
            />
            <p className="text-sm text-gray-500 mt-1">System instructions for AI analysis</p>
          </div>

          {/* Save Button */}
          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={() => router.push('/dashboard')}
              className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSaving ? 'Saving...' : 'Save Settings'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
