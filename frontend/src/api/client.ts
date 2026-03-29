import axios from 'axios'
import type { BrowseItem, Job, FileRecord, CategorySummary, Settings } from '@/types'

const api = axios.create({
  baseURL: '/api',
})

// Unwrap axios data
const unwrap = <T>(promise: Promise<{ data: T }>) => promise.then(r => r.data)

export const browseApi = {
  list: (path: string) => unwrap<BrowseItem[]>(api.get('/browse', { params: { path } })),
}

export const settingsApi = {
  get: () => unwrap<Settings>(api.get('/settings')),
  update: (settings: Partial<Settings>) => unwrap<{ status: string }>(api.put('/settings', settings)),
}

export const modelsApi = {
  list: (provider: string, endpoint?: string, apiKey?: string) =>
    unwrap<string[]>(api.get('/models', { params: { provider, endpoint, api_key: apiKey } })),
}

export const jobsApi = {
  list: () => unwrap<Job[]>(api.get('/jobs')),
  get: (id: string) => unwrap<Job>(api.get(`/jobs/${id}`)),
  create: (params: {
    input_path: string
    provider: string
    model: string
    endpoint: string
    api_key: string
    concurrency: number
    mode: string
    categories: string[]
    custom_prompt: string
    instant_move: boolean
  }) => unwrap<Job>(api.post('/jobs', params)),
  cancel: (id: string) => unwrap<{ status: string }>(api.post(`/jobs/${id}/cancel`)),
  commit: (id: string) => unwrap<{ status: string }>(api.post(`/jobs/${id}/commit`)),
  undo: (id: string) => unwrap<{ status: string }>(api.post(`/jobs/${id}/undo`)),
  delete: (id: string) => unwrap<{ status: string }>(api.delete(`/jobs/${id}`)),
}

export const filesApi = {
  list: (jobId: string, params?: { category?: string; status?: string; page?: number; per_page?: number }) =>
    unwrap<{ files: FileRecord[]; total: number; page: number }>(api.get(`/jobs/${jobId}/files`, { params })),
  summary: (jobId: string) => unwrap<CategorySummary[]>(api.get(`/jobs/${jobId}/files/summary`)),
  updateCategory: (jobId: string, fileId: number, category: string) =>
    unwrap<{ status: string }>(api.put(`/jobs/${jobId}/files/${fileId}/category`, { category })),
  bulkUpdateCategory: (jobId: string, fileIds: number[], category: string) =>
    unwrap<{ status: string; count: number }>(api.put(`/jobs/${jobId}/files/bulk-category`, { file_ids: fileIds, category })),
}
