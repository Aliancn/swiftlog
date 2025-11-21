'use client';

import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { api } from '@/lib/api';
import { AIStatus } from '@/types';

interface AIReportProps {
  runId: string;
  report?: string;
  status: AIStatus;
  onReportGenerated?: () => void;
}

export default function AIReport({ runId, report, status, onReportGenerated }: AIReportProps) {
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleGenerateReport = async () => {
    try {
      setIsGenerating(true);
      setError(null);
      await api.triggerAIAnalysis(runId);
      // Wait a bit then refresh
      setTimeout(() => {
        if (onReportGenerated) {
          onReportGenerated();
        }
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate report');
    } finally {
      setIsGenerating(false);
    }
  };

  if (status === AIStatus.Pending && !report) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-medium text-gray-900">AI Analysis</h3>
            <p className="mt-1 text-sm text-gray-500">
              Generate an AI-powered analysis of this log run
            </p>
          </div>
          <button
            onClick={handleGenerateReport}
            disabled={isGenerating}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isGenerating ? (
              <>
                <svg
                  className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                Generating...
              </>
            ) : (
              'Generate AI Report'
            )}
          </button>
        </div>
        {error && (
          <div className="mt-4 bg-red-50 border border-red-200 rounded-md p-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}
      </div>
    );
  }

  if (status === AIStatus.Processing) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-center">
          <svg
            className="animate-spin h-5 w-5 text-blue-600 mr-3"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          <div>
            <h3 className="text-lg font-medium text-gray-900">
              AI Analysis in Progress
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              Your report is being generated...
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (status === AIStatus.Failed) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-start">
          <svg
            className="h-5 w-5 text-red-600 mr-3 mt-0.5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <div className="flex-1">
            <h3 className="text-lg font-medium text-red-900">
              AI Analysis Failed
            </h3>
            <p className="mt-1 text-sm text-red-700">
              {report || 'The AI analysis encountered an error'}
            </p>
            <button
              onClick={handleGenerateReport}
              disabled={isGenerating}
              className="mt-3 inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (status === AIStatus.Completed && report) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-start justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">AI Analysis Report</h3>
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
            ✓ Completed
          </span>
        </div>
        <div className="prose prose-sm max-w-none text-gray-700">
          <ReactMarkdown
            components={{
              h1: ({ node, ...props }) => <h1 className="text-2xl font-bold mt-6 mb-4" {...props} />,
              h2: ({ node, ...props }) => <h2 className="text-xl font-bold mt-5 mb-3" {...props} />,
              h3: ({ node, ...props }) => <h3 className="text-lg font-semibold mt-4 mb-2" {...props} />,
              p: ({ node, ...props }) => <p className="mb-3 leading-relaxed" {...props} />,
              ul: ({ node, ...props }) => <ul className="list-disc list-inside mb-3 space-y-1" {...props} />,
              ol: ({ node, ...props }) => <ol className="list-decimal list-inside mb-3 space-y-1" {...props} />,
              li: ({ node, ...props }) => <li className="ml-4" {...props} />,
              pre: ({ node, children, ...props }: any) => {
                // Pre tag for code blocks
                return (
                  <pre className="bg-gray-100 p-3 rounded my-2 overflow-x-auto" {...props}>
                    {children}
                  </pre>
                );
              },
              code: ({ node, className, children, ...props }: any) => {
                // Check if this code element is inside a pre tag by checking node parent
                const isInPre = node?.parent?.tagName === 'pre';

                if (isInPre) {
                  // Code block inside pre - minimal styling
                  return (
                    <code className="text-sm font-mono text-gray-800" {...props}>
                      {children}
                    </code>
                  );
                }

                // Inline code - add background and padding
                return (
                  <code className="bg-gray-100 px-1.5 py-0.5 rounded text-sm font-mono text-gray-800" {...props}>
                    {children}
                  </code>
                );
              },
              strong: ({ node, ...props }) => <strong className="font-bold text-gray-900" {...props} />,
              em: ({ node, ...props }) => <em className="italic" {...props} />,
              a: ({ node, ...props }) => <a className="text-blue-600 hover:underline" {...props} />,
            }}
          >
            {report}
          </ReactMarkdown>
        </div>
      </div>
    );
  }

  return null;
}
