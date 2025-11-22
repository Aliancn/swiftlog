'use client';

import { use, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useProject, useProjectGroups } from '@/lib/hooks';
import { api } from '@/lib/api';

export default function ProjectPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const { id } = use(params);
  const { data: project, error: projectError, isLoading: projectLoading, mutate: mutateProject } = useProject(id);
  const { data: groups, error: groupsError, isLoading: groupsLoading, mutate: mutateGroups } = useProjectGroups(id);

  const [editingGroup, setEditingGroup] = useState<{ id: string; name: string } | null>(null);
  const [deletingGroup, setDeletingGroup] = useState<{ id: string; name: string } | null>(null);
  const [editingProject, setEditingProject] = useState<{ id: string; name: string } | null>(null);
  const [deletingProject, setDeletingProject] = useState<{ id: string; name: string } | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const isLoading = projectLoading || groupsLoading;
  const error = projectError || groupsError;

  const handleUpdateGroup = async () => {
    if (!editingGroup) return;
    setActionLoading(true);
    setActionError(null);
    try {
      await api.updateGroup(editingGroup.id, editingGroup.name);
      setEditingGroup(null);
      mutateGroups();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to update group');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDeleteGroup = async () => {
    if (!deletingGroup) return;
    setActionLoading(true);
    setActionError(null);
    try {
      await api.deleteGroup(deletingGroup.id);
      setDeletingGroup(null);
      mutateGroups();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to delete group');
    } finally {
      setActionLoading(false);
    }
  };

  const handleUpdateProject = async () => {
    if (!editingProject) return;
    setActionLoading(true);
    setActionError(null);
    try {
      await api.updateProject(editingProject.id, editingProject.name);
      setEditingProject(null);
      mutateProject();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to update project');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDeleteProject = async () => {
    if (!deletingProject) return;
    setActionLoading(true);
    setActionError(null);
    try {
      await api.deleteProject(deletingProject.id);
      setDeletingProject(null);
      router.push('/dashboard');
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to delete project');
    } finally {
      setActionLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading project...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md">
          <h2 className="text-red-800 font-semibold text-lg mb-2">Error</h2>
          <p className="text-red-600">{error.message}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="flex mb-8" aria-label="Breadcrumb">
          <ol className="flex items-center space-x-4">
            <li>
              <Link
                href="/dashboard"
                className="text-gray-400 hover:text-gray-500"
              >
                Projects
              </Link>
            </li>
            <li>
              <div className="flex items-center">
                <svg
                  className="flex-shrink-0 h-5 w-5 text-gray-300"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path d="M5.555 17.776l8-16 .894.448-8 16-.894-.448z" />
                </svg>
                <span className="ml-4 text-sm font-medium text-gray-500">
                  {project?.name}
                </span>
              </div>
            </li>
          </ol>
        </nav>

        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{project?.name}</h1>
            <p className="mt-2 text-gray-600">Log groups in this project</p>
          </div>
          {project && (
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setEditingProject({ id: project.id, name: project.name })}
                className="px-3 py-1.5 text-sm text-gray-600 hover:text-blue-600 border border-gray-300 rounded-md hover:border-blue-300"
                title="Edit project"
              >
                Edit
              </button>
              <button
                onClick={() => setDeletingProject({ id: project.id, name: project.name })}
                className="px-3 py-1.5 text-sm text-red-600 hover:text-red-700 border border-red-300 rounded-md hover:border-red-400"
                title="Delete project"
              >
                Delete
              </button>
            </div>
          )}
        </div>

        {!groups || groups.length === 0 ? (
          <div className="bg-white shadow rounded-lg p-12 text-center">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">
              No log groups yet
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              Groups will appear here when you run commands with this project.
            </p>
          </div>
        ) : (
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <ul className="divide-y divide-gray-200">
              {groups.map((group) => (
                <li key={group.id}>
                  <div className="flex items-center justify-between px-6 py-4 hover:bg-gray-50 transition-colors">
                    <Link
                      href={`/groups/${group.id}`}
                      className="flex-1 flex items-center"
                    >
                      <svg
                        className="h-5 w-5 text-gray-400 mr-3"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                        />
                      </svg>
                      <p className="text-sm font-medium text-gray-900">
                        {group.name}
                      </p>
                    </Link>
                    <div className="flex items-center space-x-4">
                      <span className="text-sm text-gray-500">
                        {new Date(group.created_at).toLocaleDateString()}
                      </span>
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={() => setEditingGroup({ id: group.id, name: group.name })}
                          className="p-1 text-gray-400 hover:text-blue-600"
                          title="Edit group"
                        >
                          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                          </svg>
                        </button>
                        <button
                          onClick={() => setDeletingGroup({ id: group.id, name: group.name })}
                          className="p-1 text-gray-400 hover:text-red-600"
                          title="Delete group"
                        >
                          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                      </div>
                      <Link href={`/groups/${group.id}`}>
                        <svg
                          className="h-5 w-5 text-gray-400"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M9 5l7 7-7 7"
                          />
                        </svg>
                      </Link>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>

      {/* Edit Group Modal */}
      {editingGroup && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Edit Group</h3>
            {actionError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
                {actionError}
              </div>
            )}
            <input
              type="text"
              value={editingGroup.name}
              onChange={(e) => setEditingGroup({ ...editingGroup, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Group name"
            />
            <div className="mt-4 flex justify-end space-x-3">
              <button
                onClick={() => { setEditingGroup(null); setActionError(null); }}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                disabled={actionLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleUpdateGroup}
                disabled={actionLoading || !editingGroup.name.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {actionLoading ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Group Confirmation Modal */}
      {deletingGroup && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Delete Group</h3>
            {actionError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
                {actionError}
              </div>
            )}
            <p className="text-gray-600 mb-4">
              Are you sure you want to delete <strong>{deletingGroup.name}</strong>? This will also delete all runs in this group. This action cannot be undone.
            </p>
            <div className="mt-4 flex justify-end space-x-3">
              <button
                onClick={() => { setDeletingGroup(null); setActionError(null); }}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                disabled={actionLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteGroup}
                disabled={actionLoading}
                className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
              >
                {actionLoading ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Project Modal */}
      {editingProject && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Edit Project</h3>
            {actionError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
                {actionError}
              </div>
            )}
            <input
              type="text"
              value={editingProject.name}
              onChange={(e) => setEditingProject({ ...editingProject, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Project name"
            />
            <div className="mt-4 flex justify-end space-x-3">
              <button
                onClick={() => { setEditingProject(null); setActionError(null); }}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                disabled={actionLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleUpdateProject}
                disabled={actionLoading || !editingProject.name.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {actionLoading ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Project Confirmation Modal */}
      {deletingProject && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Delete Project</h3>
            {actionError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
                {actionError}
              </div>
            )}
            <p className="text-gray-600 mb-4">
              Are you sure you want to delete <strong>{deletingProject.name}</strong>? This will also delete all groups and runs in this project. This action cannot be undone.
            </p>
            <div className="mt-4 flex justify-end space-x-3">
              <button
                onClick={() => { setDeletingProject(null); setActionError(null); }}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                disabled={actionLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteProject}
                disabled={actionLoading}
                className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
              >
                {actionLoading ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
