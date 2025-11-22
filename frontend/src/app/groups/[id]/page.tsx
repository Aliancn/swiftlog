'use client';

import { use, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useGroup, useGroupRuns } from '@/lib/hooks';
import { api } from '@/lib/api';
import { RunStatus } from '@/types';

export default function GroupPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const { id } = use(params);
  const [limit] = useState(50);
  const [offset] = useState(0);

  const { data: group, error: groupError, isLoading: groupLoading, mutate: mutateGroup } = useGroup(id);
  const { data: runsData, error: runsError, isLoading: runsLoading, mutate: mutateRuns } = useGroupRuns(id, {
    limit,
    offset,
  });

  const [deletingRun, setDeletingRun] = useState<{ id: string; startTime: string } | null>(null);
  const [editingGroup, setEditingGroup] = useState<{ id: string; name: string } | null>(null);
  const [deletingGroup, setDeletingGroup] = useState<{ id: string; name: string } | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const isLoading = groupLoading || runsLoading;
  const error = groupError || runsError;

  const handleDeleteRun = async () => {
    if (!deletingRun) return;
    setActionLoading(true);
    setActionError(null);
    try {
      await api.deleteRun(deletingRun.id);
      setDeletingRun(null);
      mutateRuns();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to delete run');
    } finally {
      setActionLoading(false);
    }
  };

  const handleUpdateGroup = async () => {
    if (!editingGroup) return;
    setActionLoading(true);
    setActionError(null);
    try {
      await api.updateGroup(editingGroup.id, editingGroup.name);
      setEditingGroup(null);
      mutateGroup();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to update group');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDeleteGroup = async () => {
    if (!deletingGroup) return;
    setActionLoading(true);
    setActionError(null);
    try {
      await api.deleteGroup(deletingGroup.id);
      setDeletingGroup(null);
      router.back();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to delete group');
    } finally {
      setActionLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading runs...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md">
          <h2 className="text-red-800 font-semibold text-lg mb-2">Error</h2>
          <p className="text-red-600">{error.message}</p>
        </div>
      </div>
    );
  }

  const runs = runsData?.data || [];

  const getStatusBadge = (status: RunStatus) => {
    const styles = {
      [RunStatus.Running]: 'bg-blue-100 text-blue-800',
      [RunStatus.Completed]: 'bg-green-100 text-green-800',
      [RunStatus.Failed]: 'bg-red-100 text-red-800',
      [RunStatus.Aborted]: 'bg-gray-100 text-gray-800',
    };

    return (
      <span
        className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
          styles[status] || 'bg-gray-100 text-gray-800'
        }`}
      >
        {status}
      </span>
    );
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{group?.name}</h1>
            <p className="mt-2 text-gray-600">Script execution runs in this group</p>
          </div>
          {group && (
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setEditingGroup({ id: group.id, name: group.name })}
                className="px-3 py-1.5 text-sm text-gray-600 hover:text-blue-600 border border-gray-300 rounded-md hover:border-blue-300"
                title="Edit group"
              >
                Edit
              </button>
              <button
                onClick={() => setDeletingGroup({ id: group.id, name: group.name })}
                className="px-3 py-1.5 text-sm text-red-600 hover:text-red-700 border border-red-300 rounded-md hover:border-red-400"
                title="Delete group"
              >
                Delete
              </button>
            </div>
          )}
        </div>

        {runs.length === 0 ? (
          <div className="bg-white shadow rounded-lg p-12 text-center">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No runs yet</h3>
            <p className="mt-1 text-sm text-gray-500">
              Runs will appear here when you execute scripts with this group.
            </p>
          </div>
        ) : (
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Start Time
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Exit Code
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Duration
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    AI Report
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {runs.map((run) => {
                  const duration = run.end_time
                    ? Math.round(
                        (new Date(run.end_time).getTime() -
                          new Date(run.start_time).getTime()) /
                          1000
                      )
                    : null;

                  return (
                    <tr key={run.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {new Date(run.start_time).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getStatusBadge(run.status)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {run.exit_code !== undefined && run.exit_code !== null
                          ? run.exit_code
                          : '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {duration !== null ? `${duration}s` : 'Running...'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {run.ai_status === 'completed' ? (
                          <span className="text-green-600">Available</span>
                        ) : run.ai_status === 'processing' ? (
                          <span className="text-blue-600">Processing...</span>
                        ) : (
                          <span className="text-gray-400">Pending</span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <div className="flex items-center justify-end space-x-3">
                          <Link
                            href={`/runs/${run.id}`}
                            className="text-blue-600 hover:text-blue-900"
                          >
                            View
                          </Link>
                          <button
                            onClick={() => setDeletingRun({ id: run.id, startTime: run.start_time })}
                            className="text-red-600 hover:text-red-900"
                          >
                            Delete
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Delete Run Confirmation Modal */}
      {deletingRun && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Delete Run</h3>
            {actionError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
                {actionError}
              </div>
            )}
            <p className="text-gray-600 mb-4">
              Are you sure you want to delete this run from <strong>{new Date(deletingRun.startTime).toLocaleString()}</strong>? This action cannot be undone.
            </p>
            <div className="mt-4 flex justify-end space-x-3">
              <button
                onClick={() => { setDeletingRun(null); setActionError(null); }}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                disabled={actionLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteRun}
                disabled={actionLoading}
                className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
              >
                {actionLoading ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Group Modal */}
      {editingGroup && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Edit Group</h3>
            {actionError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
                {actionError}
              </div>
            )}
            <input
              type="text"
              value={editingGroup.name}
              onChange={(e) => setEditingGroup({ ...editingGroup, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Group name"
            />
            <div className="mt-4 flex justify-end space-x-3">
              <button
                onClick={() => { setEditingGroup(null); setActionError(null); }}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                disabled={actionLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleUpdateGroup}
                disabled={actionLoading || !editingGroup.name.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {actionLoading ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Group Confirmation Modal */}
      {deletingGroup && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Delete Group</h3>
            {actionError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
                {actionError}
              </div>
            )}
            <p className="text-gray-600 mb-4">
              Are you sure you want to delete <strong>{deletingGroup.name}</strong>? This will also delete all runs in this group. This action cannot be undone.
            </p>
            <div className="mt-4 flex justify-end space-x-3">
              <button
                onClick={() => { setDeletingGroup(null); setActionError(null); }}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                disabled={actionLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteGroup}
                disabled={actionLoading}
                className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
              >
                {actionLoading ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
