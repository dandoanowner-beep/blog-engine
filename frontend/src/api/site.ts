import api from './client'

export const siteApi = {
  getAbout: () => api.get<{ content: string }>('/about'),

  updateAbout: (content: string) => api.put<{ status: string }>('/about', { content }),
}
