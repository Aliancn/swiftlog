'use client';

import { useState, useEffect, useRef } from 'react';
import type { LogLine } from '@/types';
import LogViewer from './LogViewer';

interface RealtimeLogProps {
  runId: string;
  initialLogs?: LogLine[];
  isRunning: boolean;
}

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8081';

export default function RealtimeLog({ runId, initialLogs = [], isRunning }: RealtimeLogProps) {
  const [logs, setLogs] = useState<LogLine[]>(initialLogs);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected');
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttempts = useRef(0);

  useEffect(() => {
    // Only connect if the run is still running
    if (!isRunning) {
      setConnectionStatus('disconnected');
      return;
    }

    const connect = () => {
      try {
        setConnectionStatus('connecting');

        // Get token from localStorage
        const token = typeof window !== 'undefined' ? localStorage.getItem('swiftlog_token') : null;
        const wsUrl = `${WS_URL}/ws/runs/${runId}${token ? `?token=${token}` : ''}`;

        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          setConnectionStatus('connected');
          reconnectAttempts.current = 0;
        };

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'log') {
              const newLog: LogLine = {
                timestamp: data.timestamp,
                level: data.level,
                content: data.content,
              };
              setLogs((prev) => [...prev, newLog]);
            }
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        ws.onerror = () => {
          setConnectionStatus('error');
        };

        ws.onclose = () => {
          setConnectionStatus('disconnected');

          // Attempt to reconnect with exponential backoff
          if (isRunning && reconnectAttempts.current < 10) {
            const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000);
            reconnectAttempts.current += 1;

            reconnectTimeoutRef.current = setTimeout(() => {
              connect();
            }, delay);
          }
        };
      } catch (error) {
        console.error('Failed to connect to WebSocket:', error);
        setConnectionStatus('error');
      }
    };

    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [runId, isRunning]);

  const getStatusIndicator = () => {
    switch (connectionStatus) {
      case 'connected':
        return (
          <span className="flex items-center text-green-600 text-sm">
            <span className="animate-pulse inline-block w-2 h-2 bg-green-500 rounded-full mr-2"></span>
            Connected
          </span>
        );
      case 'connecting':
        return (
          <span className="flex items-center text-yellow-600 text-sm">
            <svg className="animate-spin h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            Connecting...
          </span>
        );
      case 'error':
        return (
          <span className="flex items-center text-red-600 text-sm">
            <span className="inline-block w-2 h-2 bg-red-500 rounded-full mr-2"></span>
            Connection Error
          </span>
        );
      default:
        return (
          <span className="flex items-center text-gray-600 text-sm">
            <span className="inline-block w-2 h-2 bg-gray-400 rounded-full mr-2"></span>
            Disconnected
          </span>
        );
    }
  };

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-lg font-medium text-gray-900">Real-time Logs</h3>
        {getStatusIndicator()}
      </div>
      <LogViewer logs={logs} isLive={isRunning && connectionStatus === 'connected'} />
    </div>
  );
}
