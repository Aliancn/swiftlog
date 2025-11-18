# SwiftLog Frontend

The SwiftLog web interface built with Next.js 14, TypeScript, and Tailwind CSS.

## Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript 5
- **Styling**: Tailwind CSS
- **UI Libraries**:
  - React Markdown (for AI reports)
  - Heroicons (icons)
- **API Client**: Fetch API
- **WebSocket**: Native WebSocket API

## Features

- Project and group management
- Real-time log viewing
- Log run history
- AI-powered analysis reports
- Responsive design
- Dark mode optimized

## Getting Started

### Prerequisites

- Node.js 20+ or npm/yarn/pnpm
- SwiftLog backend services running (see root README.md)

### Development Mode

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Or with other package managers
yarn dev
pnpm dev
bun dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Production Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

### Docker Build

```bash
# Build Docker image
docker build -t swiftlog-frontend .

# Run container
docker run -p 3000:3000 swiftlog-frontend
```

## Project Structure

```
frontend/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── layout.tsx         # Root layout
│   │   ├── page.tsx           # Home page (projects list)
│   │   ├── projects/          # Project routes
│   │   │   └── [id]/         # Dynamic project page
│   │   │       ├── page.tsx  # Project detail
│   │   │       └── groups/   # Group routes
│   │   │           └── [groupId]/
│   │   │               └── page.tsx  # Group detail (runs list)
│   │   └── runs/             # Run routes
│   │       └── [id]/
│   │           └── page.tsx  # Run detail (logs + AI)
│   ├── components/           # React components
│   │   ├── AIReport.tsx     # AI analysis display
│   │   └── LogViewer.tsx    # Log display component
│   ├── lib/                 # Utility libraries
│   │   └── api.ts          # API client functions
│   └── types/              # TypeScript type definitions
│       └── index.ts        # Shared types
├── public/                 # Static assets
├── tailwind.config.ts     # Tailwind configuration
├── tsconfig.json          # TypeScript configuration
├── next.config.js         # Next.js configuration
├── Dockerfile             # Container definition
└── package.json           # Dependencies
```

## Key Components

### AIReport (`src/components/AIReport.tsx`)

Displays AI-generated analysis reports with markdown rendering.

**Features:**
- Multiple states: pending, processing, completed, failed
- Markdown rendering with syntax highlighting
- Retry functionality
- Custom styled markdown elements

**Usage:**
```tsx
<AIReport
  runId={runId}
  report={report}
  status={aiStatus}
  onReportGenerated={refreshData}
/>
```

### LogViewer (`src/components/LogViewer.tsx`)

Displays log entries with proper formatting and color coding.

**Features:**
- Color-coded log levels (STDOUT/STDERR)
- Timestamp display
- Auto-scrolling
- Optimized for large log volumes

**Usage:**
```tsx
<LogViewer logs={logs} />
```

## API Client (`src/lib/api.ts`)

Centralized API client with typed methods.

**Key Functions:**

```typescript
// Projects
api.listProjects(): Promise<Project[]>
api.getProject(id: string): Promise<Project>
api.createProject(name: string): Promise<Project>
api.getProjectGroups(projectId: string): Promise<LogGroup[]>

// Groups
api.getGroup(id: string): Promise<LogGroup>

// Runs
api.getGroupRuns(groupId: string, limit?: number, offset?: number): Promise<RunsResponse>
api.getRun(id: string): Promise<LogRun>
api.getRunLogs(runId: string, limit?: number): Promise<LogEntry[]>

// AI Analysis
api.triggerAIAnalysis(runId: string): Promise<void>
```

**Configuration:**

The API base URL is automatically determined:
- Development: `http://localhost:8080` (from environment variable)
- Production: Set `NEXT_PUBLIC_API_URL` environment variable

## Type Definitions (`src/types/index.ts`)

All TypeScript interfaces matching backend models:

```typescript
interface Project {
  id: string;
  name: string;
  user_id: string;
  created_at: string;
  updated_at: string;
}

interface LogGroup {
  id: string;
  project_id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

interface LogRun {
  id: string;
  group_id: string;
  start_time: string;
  end_time?: string;
  exit_code?: number;
  status: RunStatus;
  ai_status: AIStatus;
  ai_report?: string;
  created_at: string;
}

interface LogEntry {
  timestamp: string;
  level: string;
  content: string;
}

enum RunStatus {
  Running = 'running',
  Success = 'success',
  Failed = 'failed'
}

enum AIStatus {
  Pending = 'pending',
  Processing = 'processing',
  Completed = 'completed',
  Failed = 'failed'
}
```

## Styling

### Tailwind Configuration

Custom theme extensions in `tailwind.config.ts`:

```typescript
theme: {
  extend: {
    colors: {
      // Custom color palette
    }
  }
}
```

### CSS Organization

- Global styles: `src/app/globals.css`
- Component styles: Inline Tailwind classes
- Dark mode: Optimized for dark backgrounds

### Log Color Scheme

- **STDOUT**: White (`text-white`)
- **STDERR**: Red (`text-red-400`)
- **WARN**: Yellow (`text-yellow-400`)
- **INFO**: Blue (`text-blue-400`)

## Environment Variables

Create `.env.local` for development:

```bash
# API URL (optional, defaults to http://localhost:8080)
NEXT_PUBLIC_API_URL=http://localhost:8080

# WebSocket URL (optional, defaults to ws://localhost:8081)
NEXT_PUBLIC_WS_URL=ws://localhost:8081
```

For production, set these in your deployment environment.

## Development Guidelines

### Code Style

- Use TypeScript for all files
- Follow Next.js App Router conventions
- Use `'use client'` directive for client components
- Prefer functional components with hooks
- Use async/await for API calls

### Component Structure

```tsx
'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Project } from '@/types';

export default function ProjectList() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadProjects();
  }, []);

  const loadProjects = async () => {
    try {
      const data = await api.listProjects();
      setProjects(data);
    } catch (error) {
      console.error('Failed to load projects:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      {projects.map(project => (
        <div key={project.id}>{project.name}</div>
      ))}
    </div>
  );
}
```

### Error Handling

Always handle errors gracefully:

```typescript
try {
  const data = await api.getRun(runId);
  setRun(data);
} catch (error) {
  if (error instanceof Error) {
    setError(error.message);
  } else {
    setError('An unexpected error occurred');
  }
}
```

### Loading States

Provide visual feedback for async operations:

```tsx
{loading && <LoadingSpinner />}
{error && <ErrorMessage message={error} />}
{data && <DataDisplay data={data} />}
```

## Testing

### Run Tests

```bash
# Run unit tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run with coverage
npm test -- --coverage
```

### E2E Testing

```bash
# Install Playwright (first time)
npx playwright install

# Run E2E tests
npm run test:e2e
```

## Building

### Local Build

```bash
npm run build
```

Outputs to `.next/` directory.

### Docker Build

```bash
docker build -t swiftlog-frontend:latest .
docker run -p 3000:3000 swiftlog-frontend:latest
```

### Multi-stage Build

The Dockerfile uses multi-stage builds for optimization:

1. **deps**: Install dependencies
2. **builder**: Build application
3. **runner**: Production runtime

## Performance Optimization

### Implemented Optimizations

- Static page generation where possible
- Dynamic imports for large components
- Image optimization with Next.js Image component
- Font optimization with `next/font`
- Code splitting by route

### Future Optimizations

- [ ] Implement React Query for data caching
- [ ] Add service worker for offline support
- [ ] Optimize bundle size with tree shaking
- [ ] Add lazy loading for log entries
- [ ] Implement virtual scrolling for large logs

## Troubleshooting

### Port Already in Use

```bash
# Kill process using port 3000
lsof -ti:3000 | xargs kill -9

# Or use a different port
PORT=3001 npm run dev
```

### API Connection Failed

1. Verify backend services are running:
   ```bash
   curl http://localhost:8080/health
   ```

2. Check `NEXT_PUBLIC_API_URL` environment variable

3. Check browser console for CORS errors

### Styles Not Loading

```bash
# Clear Next.js cache
rm -rf .next

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install

# Rebuild
npm run dev
```

### Type Errors

```bash
# Rebuild TypeScript
npm run build

# Check types manually
npx tsc --noEmit
```

## Deployment

### Vercel (Recommended)

1. Push code to GitHub
2. Import project in Vercel
3. Set environment variables
4. Deploy

### Docker

```bash
docker build -t swiftlog-frontend .
docker push your-registry/swiftlog-frontend:latest
```

### Static Export (if applicable)

```bash
npm run build
npm run export
```

Output in `out/` directory.

## Browser Support

- Chrome/Edge (last 2 versions)
- Firefox (last 2 versions)
- Safari 14+
- Mobile browsers (iOS Safari 14+, Chrome Android)

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) in the project root.

## Support

- **Documentation**: [Main README](../README.md)
- **API Documentation**: [docs/API.md](../docs/API.md)
- **Issues**: GitHub Issues

## License

See [LICENSE](../LICENSE) in the project root.
