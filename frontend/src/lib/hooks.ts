'use client';

import { useState, useEffect } from 'react';
import { api } from './api';
import type { Project, LogGroup, LogRun, LogLine, PaginatedResponse } from '@/types';

interface UseDataResult<T> {
  data: T | undefined;
  error: Error | undefined;
  isLoading: boolean;
  mutate: () => Promise<void>;
}

function useData<T>(key: string | null, fetcher: () => Promise<T>): UseDataResult<T> {
  const [data, setData] = useState<T | undefined>(undefined);
  const [error, setError] = useState<Error | undefined>(undefined);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(
    function effectCallback() {
      if (!key) {
        setIsLoading(false);
        return;
      }

      let cancelled = false;

      async function fetchData() {
        try {
          setIsLoading(true);
          const result = await fetcher();
          if (!cancelled) {
            setData(result);
            setError(undefined);
          }
        } catch (err) {
          if (!cancelled) {
            setError(err instanceof Error ? err : new Error('Unknown error'));
          }
        } finally {
          if (!cancelled) {
            setIsLoading(false);
          }
        }
      }

      fetchData();

      return function cleanup() {
        cancelled = true;
      };
    },
    [key]
  );

  async function mutate() {
    if (!key) return;
    try {
      const result = await fetcher();
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'));
    }
  }

  return { data, error, isLoading, mutate };
}

export function useProjects(): UseDataResult<Project[]> {
  return useData<Project[]>('projects', function fetchProjects() {
    return api.getProjects();
  });
}

export function useProject(id: string | null): UseDataResult<Project> {
  return useData<Project>(
    id ? `project-${id}` : null,
    function fetchProject() {
      return api.getProject(id!);
    }
  );
}

export function useProjectGroups(projectId: string | null): UseDataResult<LogGroup[]> {
  return useData<LogGroup[]>(
    projectId ? `project-${projectId}-groups` : null,
    function fetchProjectGroups() {
      return api.getProjectGroups(projectId!);
    }
  );
}

export function useGroup(id: string | null): UseDataResult<LogGroup> {
  return useData<LogGroup>(
    id ? `group-${id}` : null,
    function fetchGroup() {
      return api.getGroup(id!);
    }
  );
}

export function useGroupRuns(
  groupId: string | null,
  params?: { limit?: number; offset?: number }
): UseDataResult<PaginatedResponse<LogRun>> {
  const key = groupId
    ? `group-${groupId}-runs-${params?.limit || 50}-${params?.offset || 0}`
    : null;

  return useData<PaginatedResponse<LogRun>>(
    key,
    function fetchGroupRuns() {
      return api.getGroupRuns(groupId!, params);
    }
  );
}

export function useRun(id: string | null): UseDataResult<LogRun> {
  return useData<LogRun>(
    id ? `run-${id}` : null,
    function fetchRun() {
      return api.getRun(id!);
    }
  );
}

export function useRunLogs(id: string | null): UseDataResult<LogLine[]> {
  return useData<LogLine[]>(
    id ? `run-${id}-logs` : null,
    function fetchRunLogs() {
      return api.getRunLogs(id!);
    }
  );
}

export function useAIReport(id: string | null): UseDataResult<{ report: string; status: string }> {
  return useData<{ report: string; status: string }>(
    id ? `run-${id}-ai-report` : null,
    function fetchAIReport() {
      return api.getAIReport(id!);
    }
  );
}

export function useStatistics(): UseDataResult<{
  run_statistics: {
    running: number;
    completed: number;
    failed: number;
    aborted: number;
    total: number;
  };
  ai_statistics: {
    pending: number;
    processing: number;
    completed: number;
    failed: number;
    total: number;
  };
  queue_length: number;
}> {
  return useData('statistics', function fetchStatistics() {
    return api.getStatistics();
  });
}

export function useRecentRuns(limit?: number): UseDataResult<PaginatedResponse<LogRun>> {
  return useData(`recent-runs-${limit || 20}`, function fetchRecentRuns() {
    return api.getRecentRuns(limit);
  });
}
