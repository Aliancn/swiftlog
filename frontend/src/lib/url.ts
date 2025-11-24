// URL utilities for SwiftLog frontend
// Handles dynamic URL construction for API and WebSocket connections

/**
 * Get the base API URL
 * - In production (Docker): uses relative path /api/v1 through nginx
 * - In development: uses NEXT_PUBLIC_API_URL or defaults to localhost:8080
 */
export function getApiBaseUrl(): string {
  const configuredUrl = process.env.NEXT_PUBLIC_API_URL;

  // If no URL configured, use default
  if (!configuredUrl) {
    return 'http://localhost:8080/api/v1';
  }

  // If it's a relative path, use as-is (browser will resolve it)
  if (configuredUrl.startsWith('/')) {
    return configuredUrl;
  }

  // If it's a full URL, use as-is
  return configuredUrl;
}

/**
 * Get the WebSocket URL for a specific endpoint
 * Uses NEXT_PUBLIC_WS_BASE_URL (base URL) and appends the full path
 * Handles conversion from HTTP/HTTPS URLs to WS/WSS URLs
 *
 * @param path - The WebSocket path (e.g., '/ws/runs/123')
 * @returns Full WebSocket URL (e.g., 'wss://example.com/ws/runs/123')
 */
export function getWebSocketUrl(path: string): string {
  // Ensure path starts with /
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;

  const baseUrl = process.env.NEXT_PUBLIC_WS_BASE_URL;

  // If no base URL configured, use default development URL
  if (!baseUrl) {
    return `ws://localhost:8081${normalizedPath}`;
  }

  // Convert HTTP/HTTPS to WS/WSS
  let wsBaseUrl: string;
  if (baseUrl.startsWith('https://')) {
    wsBaseUrl = baseUrl.replace(/^https:\/\//, 'wss://');
  } else if (baseUrl.startsWith('http://')) {
    wsBaseUrl = baseUrl.replace(/^http:\/\//, 'ws://');
  } else if (baseUrl.startsWith('wss://') || baseUrl.startsWith('ws://')) {
    wsBaseUrl = baseUrl;
  } else {
    // Relative path or invalid format - fall back to window.location
    if (typeof window !== 'undefined') {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      return `${protocol}//${host}${normalizedPath}`;
    }
    // Server-side fallback
    return `ws://localhost:8081${normalizedPath}`;
  }

  // Remove trailing slash from base URL
  const cleanBaseUrl = wsBaseUrl.replace(/\/$/, '');

  // Append the path
  return `${cleanBaseUrl}${normalizedPath}`;
}

/**
 * Get the WebSocket URL for a run's real-time logs
 * @param runId - The run ID to subscribe to
 */
export function getRunWebSocketUrl(runId: string): string {
  return getWebSocketUrl(`/ws/runs/${runId}`);
}
