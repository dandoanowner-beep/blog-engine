import api from './client'

export interface Project {
  id: string
  title: string
  description: string
  tech_stack: string
  repo_url: string
  demo_url: string
  thumbnail_url: string
  sort_order: number
}

export interface ProjectInput {
  title: string
  description?: string
  tech_stack?: string
  repo_url?: string
  demo_url?: string
  thumbnail_url?: string
  sort_order?: number
}

export const projectsApi = {
  list: () => api.get<{ projects: Project[] }>('/projects'),

  create: (input: ProjectInput) => api.post<Project>('/projects', input),

  update: (id: string, input: Partial<ProjectInput>) =>
    api.patch<Project>(`/projects/${id}`, input),

  remove: (id: string) => api.delete(`/projects/${id}`),
}
