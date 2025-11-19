'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { api } from '@/lib/api';

export default function Navigation() {
  const pathname = usePathname();
  const router = useRouter();
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [username, setUsername] = useState('');

  useEffect(() => {
    checkAuth();
  }, [pathname]);

  const checkAuth = async () => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('swiftlog_token') : null;
    if (!token) {
      setIsLoggedIn(false);
      return;
    }

    try {
      const user = await api.getCurrentUser();
      setIsLoggedIn(true);
      setUsername(user.username);
    } catch {
      setIsLoggedIn(false);
      setUsername('');
    }
  };

  const handleLogout = () => {
    api.logout();
    setIsLoggedIn(false);
    setUsername('');
    router.push('/login');
  };

  // Don't show navigation on login page
  if (pathname === '/login') {
    return null;
  }

  return (
    <nav className="bg-white shadow-sm border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center">
            <Link href="/dashboard" className="flex items-center">
              <svg
                className="h-8 w-8 text-blue-600"
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
              <span className="ml-2 text-xl font-bold text-gray-900">
                SwiftLog
              </span>
            </Link>
          </div>
          <div className="flex items-center space-x-4">
            <Link
              href="/dashboard"
              className="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium hover:bg-gray-100 transition-colors"
            >
              Dashboard
            </Link>
            <Link
              href="/status"
              className="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium hover:bg-gray-100 transition-colors"
            >
              Status
            </Link>
            {isLoggedIn && (
              <>
                <Link
                  href="/settings"
                  className="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium hover:bg-gray-100 transition-colors"
                >
                  Settings
                </Link>
                <Link
                  href="/tokens"
                  className="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium hover:bg-gray-100 transition-colors"
                >
                  Tokens
                </Link>
                <div className="flex items-center space-x-3 border-l pl-4">
                  <span className="text-sm text-gray-600">
                    {username}
                  </span>
                  <button
                    onClick={handleLogout}
                    className="text-sm text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md font-medium hover:bg-gray-100 transition-colors"
                  >
                    Logout
                  </button>
                </div>
              </>
            )}
            {!isLoggedIn && (
              <Link
                href="/login"
                className="text-sm text-blue-600 hover:text-blue-800 px-3 py-2 rounded-md font-medium hover:bg-blue-50 transition-colors"
              >
                Login
              </Link>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
