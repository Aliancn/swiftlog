// SwiftLog Frontend Types

export enum RunStatus {
  Running = 'running',
  Completed = 'completed',
  Failed = 'failed',
  Aborted = 'aborted',
}

export enum AIStatus {
  None = 'none',
  Pending = 'pending',
  Processing = 'processing',
  Completed = 'completed',
  Failed = 'failed',
}

export enum LogLevel {
  Stdout = 'STDOUT',
  Stderr = 'STDERR',
}

export interface Project {
  id: string;
  user_id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface LogGroup {
  id: string;
  project_id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface LogRun {
  id: string;
  group_id: string;
  start_time: string;
  end_time?: string;
  status: RunStatus;
  exit_code?: number;
  ai_report?: string;
  ai_status: AIStatus;
  created_at: string;
  updated_at: string;
}

export interface LogLine {
  timestamp: string;
  level: string;
  content: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface APIError {
  error: string;
  message?: string;
}

export enum TruncateStrategy {
  Head = 'head',
  Tail = 'tail',
  Smart = 'smart',
}

export interface UserSettings {
  id: string;
  user_id: string;
  ai_enabled: boolean;
  ai_base_url: string;
  ai_model: string;
  ai_max_tokens: number;
  ai_auto_analyze: boolean;
  ai_max_log_lines: number;
  ai_log_truncate_strategy: TruncateStrategy;
  ai_system_prompt: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectSettings {
  id: string;
  project_id: string;
  ai_enabled?: boolean;
  ai_base_url?: string;
  ai_model?: string;
  ai_max_tokens?: number;
  ai_auto_analyze?: boolean;
  ai_max_log_lines?: number;
  ai_log_truncate_strategy?: TruncateStrategy;
  ai_system_prompt?: string;
  created_at: string;
  updated_at: string;
}

export interface EffectiveSettings {
  ai_enabled: boolean;
  ai_base_url: string;
  ai_model: string;
  ai_max_tokens: number;
  ai_auto_analyze: boolean;
  ai_max_log_lines: number;
  ai_log_truncate_strategy: TruncateStrategy;
  ai_system_prompt: string;
  source: 'user' | 'project' | 'merged';
}
