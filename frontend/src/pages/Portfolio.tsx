import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { projectsApi, type Project, type ProjectInput } from '../api/projects'
import { useAuthStore } from '../store/auth'

const emptyForm: ProjectInput = {
  title: '',
  description: '',
  tech_stack: '',
  repo_url: '',
  demo_url: '',
  thumbnail_url: '',
}

// FR-CR002-001: public portfolio of the owner's projects; owner manages inline.
export default function Portfolio() {
  const { user } = useAuthStore()
  const isOwner = user?.role === 'owner'
  const qc = useQueryClient()

  const [editing, setEditing] = useState<string | 'new' | null>(null)
  const [form, setForm] = useState<ProjectInput>(emptyForm)
  const [error, setError] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectsApi.list().then((r) => r.data),
  })

  const invalidate = () => qc.invalidateQueries({ queryKey: ['projects'] })

  const createMut = useMutation({
    mutationFn: (input: ProjectInput) => projectsApi.create(input),
    onSuccess: () => { invalidate(); setEditing(null); setForm(emptyForm) },
    onError: () => setError('Failed to save project'),
  })
  const updateMut = useMutation({
    mutationFn: ({ id, input }: { id: string; input: ProjectInput }) => projectsApi.update(id, input),
    onSuccess: () => { invalidate(); setEditing(null); setForm(emptyForm) },
    onError: () => setError('Failed to save project'),
  })
  const removeMut = useMutation({
    mutationFn: (id: string) => projectsApi.remove(id),
    onSuccess: invalidate,
  })

  const projects = data?.projects ?? []

  const startEdit = (p: Project) => {
    setEditing(p.id)
    setForm({
      title: p.title,
      description: p.description,
      tech_stack: p.tech_stack,
      repo_url: p.repo_url,
      demo_url: p.demo_url,
      thumbnail_url: p.thumbnail_url,
    })
  }

  const save = () => {
    setError('')
    if (editing === 'new') createMut.mutate(form)
    else if (editing) updateMut.mutate({ id: editing, input: form })
  }

  const field = (key: keyof ProjectInput, label: string, placeholder = '') => (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <input
        value={(form[key] as string) ?? ''}
        onChange={(e) => setForm({ ...form, [key]: e.target.value })}
        placeholder={placeholder}
        className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        data-testid={`project-${key.replace('_', '-')}-input`}
      />
    </div>
  )

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Portfolio</h1>
        {isOwner && editing === null && (
          <button
            onClick={() => { setForm(emptyForm); setEditing('new') }}
            className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
            data-testid="add-project-btn"
          >
            Add project
          </button>
        )}
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 text-sm px-3 py-2 rounded mb-4" role="alert">{error}</div>
      )}

      {editing !== null && (
        <div className="border border-gray-200 rounded-lg p-4 mb-6 space-y-3 bg-white" data-testid="project-form">
          <h2 className="font-semibold">{editing === 'new' ? 'New project' : 'Edit project'}</h2>
          {field('title', 'Title', 'My awesome project')}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              value={form.description ?? ''}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              rows={3}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              data-testid="project-description-input"
            />
          </div>
          {field('tech_stack', 'Tech stack', 'Go, React, PostgreSQL')}
          {field('repo_url', 'Repository URL', 'https://github.com/…')}
          {field('demo_url', 'Demo URL', 'https://…')}
          {field('thumbnail_url', 'Thumbnail URL', 'https://…')}
          <div className="flex gap-3 justify-end">
            <button
              onClick={() => { setEditing(null); setForm(emptyForm) }}
              className="border border-gray-300 px-4 py-2 rounded text-sm hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={save}
              disabled={createMut.isPending || updateMut.isPending}
              className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700 disabled:opacity-60"
              data-testid="project-save-btn"
            >
              Save
            </button>
          </div>
        </div>
      )}

      {isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-48 bg-gray-100 rounded-lg animate-pulse" />
          ))}
        </div>
      )}

      {!isLoading && projects.length === 0 && (
        <div className="text-center py-12 text-gray-400">No projects yet.</div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {projects.map((p) => (
          <div key={p.id} className="bg-white rounded-lg shadow-sm border border-gray-100 overflow-hidden" data-testid="project-card">
            {p.thumbnail_url && (
              <img src={p.thumbnail_url} alt={p.title} className="w-full h-40 object-cover" />
            )}
            <div className="p-4">
              <h2 className="font-semibold text-gray-900 mb-1">{p.title}</h2>
              {/* description is sanitized server-side (bluemonday) */}
              <div className="text-gray-500 text-sm mb-3 prose prose-sm max-w-none" dangerouslySetInnerHTML={{ __html: p.description }} />
              {p.tech_stack && (
                <p className="text-xs text-gray-400 mb-3">{p.tech_stack}</p>
              )}
              <div className="flex items-center gap-4 text-sm">
                {p.repo_url && (
                  <a href={p.repo_url} target="_blank" rel="noreferrer" className="text-blue-600 hover:underline">
                    Code
                  </a>
                )}
                {p.demo_url && (
                  <a href={p.demo_url} target="_blank" rel="noreferrer" className="text-blue-600 hover:underline">
                    Live demo
                  </a>
                )}
                {isOwner && (
                  <span className="ml-auto flex gap-3">
                    <button onClick={() => startEdit(p)} className="text-gray-500 hover:text-blue-600" data-testid={`edit-project-${p.id}`}>
                      Edit
                    </button>
                    <button
                      onClick={() => { if (window.confirm('Delete this project?')) removeMut.mutate(p.id) }}
                      className="text-gray-500 hover:text-red-500"
                      data-testid={`delete-project-${p.id}`}
                    >
                      Delete
                    </button>
                  </span>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
