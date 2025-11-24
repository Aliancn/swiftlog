import { AIStatus } from '@/types';

export default function AIStatusBadge({ status }: { status: AIStatus }) {
  const styles = {
    [AIStatus.None]: 'bg-gray-100 text-gray-600 border-gray-300',
    [AIStatus.Pending]: 'bg-yellow-50 text-yellow-700 border-yellow-200',
    [AIStatus.Processing]: 'bg-blue-50 text-blue-700 border-blue-200',
    [AIStatus.Completed]: 'bg-green-50 text-green-700 border-green-200',
    [AIStatus.Failed]: 'bg-red-50 text-red-700 border-red-200',
  };

  return (
    <span className={`px-3 py-1 rounded-full text-xs font-medium border ${styles[status]}`}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
}