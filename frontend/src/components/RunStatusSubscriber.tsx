'use client';

import { useEffect, useRef } from 'react';

interface RunStatusSubscriberProps {
  runId: string;
  onRunUpdate: () => void;
}

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8081';

/**
 * RunStatusSubscriber creates a WebSocket connection to receive run status updates.
 * This is used for non-running runs to get real-time updates when AI analysis completes.
 */
export default function RunStatusSubscriber({ runId, onRunUpdate }: RunStatusSubscriberProps) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5; // Fewer attempts for non-running runs

  useEffect(() => {
    const connect = () => {
      try {
        // Get token from localStorage
        const token = typeof window !== 'undefined' ? localStorage.getItem('swiftlog_token') : null;
        const wsUrl = `${WS_URL}/ws/runs/${runId}${token ? `?token=${token}` : ''}`;

        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          console.log('RunStatusSubscriber connected for run:', runId);
          reconnectAttempts.current = 0;
        };

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'run_update') {
              console.log('Run status update received:', data);
              onRunUpdate();
            }
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        ws.onerror = () => {
          console.error('RunStatusSubscriber connection error');
        };

        ws.onclose = () => {
          console.log('RunStatusSubscriber disconnected');

          // Attempt to reconnect with exponential backoff
          if (reconnectAttempts.current < maxReconnectAttempts) {
            const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 10000);
            reconnectAttempts.current += 1;

            reconnectTimeoutRef.current = setTimeout(() => {
              connect();
            }, delay);
          }
        };
      } catch (error) {
        console.error('Failed to connect RunStatusSubscriber:', error);
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
  }, [runId, onRunUpdate]);

  // This component doesn't render anything
  return null;
}
