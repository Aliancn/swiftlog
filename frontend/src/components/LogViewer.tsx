'use client';

import { useEffect, useRef } from 'react';
import type { LogLine } from '@/types';

interface LogViewerProps {
  logs: LogLine[];
  isLive?: boolean;
}

export default function LogViewer({ logs, isLive = false }: LogViewerProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new logs arrive in live mode
  useEffect(() => {
    if (isLive && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs, isLive]);

  const getLogColor = (level: string, content: string) => {
    const lowerContent = content.toLowerCase();

    // Check for actual error indicators in content
    const errorKeywords = ['error', 'failed', 'failure', 'fatal', 'panic', 'exception'];
    const hasError = errorKeywords.some(keyword => lowerContent.includes(keyword));

    // Only color as red if it's STDERR AND contains actual error keywords
    if ((level === 'STDERR' || level.toUpperCase() === 'STDERR') && hasError) {
      return 'text-red-400';
    }

    // Check for warning indicators in content
    if (lowerContent.includes('warning') || lowerContent.includes('warn')) {
      return 'text-yellow-400';
    }

    // Default color for normal output (including STDERR status messages)
    return 'text-gray-300';
  };

  const formatLogContent = (content: string) => {
    // Remove [stdout] or [stderr] prefix if present at the start
    let cleaned = content.replace(/^\[(stdout|stderr|STDOUT|STDERR)\]\s*/, '');

    // If the line is now empty or only whitespace, return a single space to preserve the line
    if (!cleaned.trim()) {
      return '';
    }

    return cleaned;
  };

  return (
    <div className="bg-gray-900 rounded-lg overflow-hidden">
      <div className="bg-gray-800 px-4 py-2 flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <svg
            className="h-4 w-4 text-gray-400"
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
          <span className="text-gray-300 text-sm font-medium">Logs</span>
          {isLive && (
            <span className="flex items-center">
              <span className="animate-pulse inline-block w-2 h-2 bg-green-500 rounded-full mr-2"></span>
              <span className="text-green-400 text-xs">Live</span>
            </span>
          )}
        </div>
        <span className="text-gray-400 text-xs">{logs.length} lines</span>
      </div>

      <div className="overflow-auto max-h-[600px] p-4 font-mono text-sm">
        {logs.length === 0 ? (
          <div className="text-gray-500 text-center py-8">
            No logs available yet...
          </div>
        ) : (
          <div className="space-y-1">
            {logs.map((log, index) => {
              const formattedContent = formatLogContent(log.content);
              // Display empty content as a single space to preserve empty lines
              const displayContent = formattedContent || '\u00A0';

              return (
                <div
                  key={index}
                  className="flex items-start hover:bg-gray-800 px-2 py-1 rounded"
                >
                  <span className="text-gray-500 select-none mr-4 flex-shrink-0 w-8 text-right">
                    {index + 1}
                  </span>
                  <span className="text-gray-400 select-none mr-4 flex-shrink-0">
                    {new Date(log.timestamp).toLocaleTimeString()}
                  </span>
                  <span className={`flex-1 whitespace-pre-wrap break-all ${getLogColor(log.level, formattedContent)}`}>
                    {displayContent}
                  </span>
                </div>
              );
            })}
            <div ref={bottomRef} />
          </div>
        )}
      </div>
    </div>
  );
}
