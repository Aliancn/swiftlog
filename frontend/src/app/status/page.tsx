'use client';

import { useEffect } from 'react';
import { useStatistics, useRecentRuns } from '@/lib/hooks';
import { RunStatus, AIStatus } from '@/types';
import Link from 'next/link';

export default function StatusPage() {
  const { data: stats, error: statsError, isLoading: statsLoading, mutate: refreshStats } = useStatistics();
  const { data: recentRuns, error: runsError, isLoading: runsLoading, mutate: refreshRuns } = useRecentRuns(20);

  // Auto-refresh every 5 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      refreshStats();
      refreshRuns();
    }, 5000);

    return () => clearInterval(interval);
  }, [refreshStats, refreshRuns]);

  if (statsLoading || runsLoading) {
    return (
      <div className="min-h-screen bg-gray-900 text-white p-8">
        <div className="max-w-7xl mx-auto">
          <div className="flex justify-center items-center h-64">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
          </div>
        </div>
      </div>
    );
  }

  if (statsError || runsError) {
    return (
      <div className="min-h-screen bg-gray-900 text-white p-8">
        <div className="max-w-7xl mx-auto">
          <div className="bg-red-900/50 border border-red-700 rounded p-4">
            <p className="text-red-200">Error: {statsError?.message || runsError?.message}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold">System Status</h1>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-sm text-gray-400">Live (auto-refresh 5s)</span>
            </div>
            <button
              onClick={() => {
                refreshStats();
                refreshRuns();
              }}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded transition-colors"
            >
              Refresh Now
            </button>
          </div>
        </div>

        {/* Statistics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          {/* Run Statistics Card */}
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h2 className="text-xl font-semibold mb-4 text-blue-400">Log Run Status</h2>
            <div className="space-y-3">
              <StatRow label="Running" value={stats?.run_statistics.running || 0} color="blue" />
              <StatRow label="Completed" value={stats?.run_statistics.completed || 0} color="green" />
              <StatRow label="Failed" value={stats?.run_statistics.failed || 0} color="red" />
              <StatRow label="Aborted" value={stats?.run_statistics.aborted || 0} color="gray" />
              <div className="pt-3 border-t border-gray-700">
                <StatRow label="Total" value={stats?.run_statistics.total || 0} color="white" bold />
              </div>
            </div>
          </div>

          {/* AI Analysis Statistics Card */}
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h2 className="text-xl font-semibold mb-4 text-purple-400">AI Analysis Status</h2>
            <div className="space-y-3">
              <StatRow label="Pending" value={stats?.ai_statistics.pending || 0} color="yellow" />
              <StatRow label="Processing" value={stats?.ai_statistics.processing || 0} color="blue" />
              <StatRow label="Completed" value={stats?.ai_statistics.completed || 0} color="green" />
              <StatRow label="Failed" value={stats?.ai_statistics.failed || 0} color="red" />
              <div className="pt-3 border-t border-gray-700">
                <StatRow label="Total" value={stats?.ai_statistics.total || 0} color="white" bold />
              </div>
            </div>
          </div>

          {/* Queue Status Card */}
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h2 className="text-xl font-semibold mb-4 text-green-400">Task Queue</h2>
            <div className="space-y-3">
              <div className="text-center py-8">
                <div className="text-5xl font-bold text-green-400 mb-2">
                  {stats?.queue_length || 0}
                </div>
                <div className="text-gray-400">Tasks in Queue</div>
              </div>
              <div className="text-sm text-gray-500 text-center">
                {stats?.queue_length === 0 ? '✓ Queue is empty' : `${stats?.queue_length} tasks waiting`}
              </div>
            </div>
          </div>
        </div>

        {/* Recent Runs Table */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-6 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white">Recent Runs</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-750 border-b border-gray-700">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Run ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Start Time
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Exit Code
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    AI Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {recentRuns?.data.map((run) => (
                  <tr key={run.id} className="hover:bg-gray-750 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-300">
                      {run.id.slice(0, 8)}...
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                      {new Date(run.start_time).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <StatusBadge status={run.status} />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                      {run.exit_code !== null && run.exit_code !== undefined ? run.exit_code : 'N/A'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <AIStatusBadge status={run.ai_status} />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <Link
                        href={`/runs/${run.id}`}
                        className="text-blue-400 hover:text-blue-300 transition-colors"
                      >
                        View Details →
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {recentRuns?.data.length === 0 && (
            <div className="p-8 text-center text-gray-400">No recent runs found</div>
          )}
        </div>
      </div>
    </div>
  );
}

// Helper Components
function StatRow({
  label,
  value,
  color,
  bold,
}: {
  label: string;
  value: number;
  color: string;
  bold?: boolean;
}) {
  const colorClasses = {
    blue: 'text-blue-400',
    green: 'text-green-400',
    red: 'text-red-400',
    yellow: 'text-yellow-400',
    gray: 'text-gray-400',
    purple: 'text-purple-400',
    white: 'text-white',
  };

  return (
    <div className="flex justify-between items-center">
      <span className={`text-sm ${bold ? 'font-semibold' : ''}`}>{label}</span>
      <span className={`text-lg ${bold ? 'font-bold' : 'font-semibold'} ${colorClasses[color as keyof typeof colorClasses]}`}>
        {value}
      </span>
    </div>
  );
}

function StatusBadge({ status }: { status: RunStatus }) {
  const styles = {
    [RunStatus.Running]: 'bg-blue-900/50 text-blue-300 border-blue-700',
    [RunStatus.Completed]: 'bg-green-900/50 text-green-300 border-green-700',
    [RunStatus.Failed]: 'bg-red-900/50 text-red-300 border-red-700',
    [RunStatus.Aborted]: 'bg-gray-700/50 text-gray-300 border-gray-600',
  };

  return (
    <span className={`px-3 py-1 rounded-full text-xs font-medium border ${styles[status]}`}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
}

function AIStatusBadge({ status }: { status: AIStatus }) {
  const styles = {
    [AIStatus.None]: 'bg-gray-700/50 text-gray-400 border-gray-600',
    [AIStatus.Pending]: 'bg-yellow-900/50 text-yellow-300 border-yellow-700',
    [AIStatus.Processing]: 'bg-blue-900/50 text-blue-300 border-blue-700',
    [AIStatus.Completed]: 'bg-green-900/50 text-green-300 border-green-700',
    [AIStatus.Failed]: 'bg-red-900/50 text-red-300 border-red-700',
  };

  return (
    <span className={`px-3 py-1 rounded-full text-xs font-medium border ${styles[status]}`}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
}
