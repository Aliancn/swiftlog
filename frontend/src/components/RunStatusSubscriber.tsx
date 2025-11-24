'use client';

import { useEffect, useRef } from 'react';
import { getRunWebSocketUrl } from '@/lib/url';

interface RunStatusSubscriberProps {
  runId: string;
  onRunUpdate: () => void;
}

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

        // Build WebSocket URL dynamically
        const wsUrl = getRunWebSocketUrl(runId);

        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          // Send authentication message after connection
          if (token) {
            ws.send(JSON.stringify({
              type: 'auth',
              token: token
            }));
          } else {
            console.error('No auth token available');
            ws.close();
            return;
          }
        };

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);

            // Handle authentication response
            if (data.type === 'auth_success') {
              if (process.env.NODE_ENV === 'development') {
                console.log('RunStatusSubscriber authenticated for run:', runId);
              }
              reconnectAttempts.current = 0;
              return;
            }

            // Handle authentication error
            if (data.error) {
              console.error('RunStatusSubscriber auth error:', data.error);
              ws.close();
              return;
            }

            if (data.type === 'run_update') {
              if (process.env.NODE_ENV === 'development') {
                console.log('Run status update received:', data);
              }
              onRunUpdate();
            }
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        ws.onerror = () => {
          if (process.env.NODE_ENV === 'development') {
            console.error('RunStatusSubscriber connection error');
          }
        };

        ws.onclose = () => {
          if (process.env.NODE_ENV === 'development') {
            console.log('RunStatusSubscriber disconnected');
          }

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
