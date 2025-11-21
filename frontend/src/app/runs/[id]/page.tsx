'use client';

import { use, useEffect, useRef } from 'react';
import Link from 'next/link';
import { useRun, useRunLogs } from '@/lib/hooks';
import LogViewer from '@/components/LogViewer';
import RealtimeLog from '@/components/RealtimeLog';
import RunStatusSubscriber from '@/components/RunStatusSubscriber';
import AIReport from '@/components/AIReport';
import { RunStatus, AIStatus } from '@/types';

export default function RunDetailsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { data: run, error: runError, isLoading: runLoading, mutate: refreshRun } = useRun(id);
  const { data: logs, error: logsError, isLoading: logsLoading, mutate: refreshLogs } = useRunLogs(id);
  const previousStatusRef = useRef<RunStatus | null>(null);

  const isLoading = runLoading || logsLoading;
  const error = runError || logsError;

  // Handle run status update from WebSocket
  const handleRunUpdate = () => {
    console.log('Refreshing run data due to WebSocket update');
    refreshRun();

    // If run just completed, also refresh logs to get complete set
    if (run && run.status === RunStatus.Running) {
      refreshLogs();
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading run details...</p>
        </div>
      </div>
    );
  }

  if (error || !run) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md">
          <h2 className="text-red-800 font-semibold text-lg mb-2">Error</h2>
          <p className="text-red-600">{error?.message || 'Run not found'}</p>
        </div>
      </div>
    );
  }

  const isRunning = run.status === RunStatus.Running;
  const duration = run.end_time
    ? Math.round((new Date(run.end_time).getTime() - new Date(run.start_time).getTime()) / 1000)
    : null;

  const getStatusBadge = (status: RunStatus) => {
    const styles = {
      [RunStatus.Running]: 'bg-blue-100 text-blue-800',
      [RunStatus.Completed]: 'bg-green-100 text-green-800',
      [RunStatus.Failed]: 'bg-red-100 text-red-800',
      [RunStatus.Aborted]: 'bg-gray-100 text-gray-800',
    };

    return (
      <span
        className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
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
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-3xl font-bold text-gray-900">Run Details</h1>
            {getStatusBadge(run.status)}
          </div>

          {/* Metadata */}
          <div className="bg-white shadow rounded-lg p-6">
            <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-3">
              <div>
                <dt className="text-sm font-medium text-gray-500">Run ID</dt>
                <dd className="mt-1 text-sm text-gray-900 font-mono">{run.id}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Start Time</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {new Date(run.start_time).toLocaleString()}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">End Time</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {run.end_time ? new Date(run.end_time).toLocaleString() : 'Running...'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Duration</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {duration !== null ? `${duration} seconds` : 'In progress...'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Exit Code</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {run.exit_code !== undefined && run.exit_code !== null ? (
                    <span
                      className={
                        run.exit_code === 0
                          ? 'text-green-600 font-semibold'
                          : 'text-red-600 font-semibold'
                      }
                    >
                      {run.exit_code}
                    </span>
                  ) : (
                    '-'
                  )}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Status</dt>
                <dd className="mt-1 text-sm">{getStatusBadge(run.status)}</dd>
              </div>
            </dl>
          </div>
        </div>

        {/* Logs Section */}
        <div className="mb-8">
          {isRunning ? (
            <RealtimeLog
              runId={id}
              initialLogs={logs || []}
              isRunning={isRunning}
              onRunUpdate={handleRunUpdate}
            />
          ) : (
            <>
              <h3 className="text-lg font-medium text-gray-900 mb-2">Logs</h3>
              <LogViewer logs={logs || []} />
              {/* Subscribe to status updates for completed runs */}
              <RunStatusSubscriber runId={id} onRunUpdate={handleRunUpdate} />
            </>
          )}
        </div>

        {/* AI Report Section */}
        {!isRunning && (
          <div className="mb-8">
            <AIReport
              runId={id}
              report={run.ai_report}
              status={run.ai_status}
              onReportGenerated={refreshRun}
            />
          </div>
        )}
      </div>
    </div>
  );
}
