const BASE = ''

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export interface Collection {
  id: string
  name: string
  icon: string
  parent_id: string | null
  sort_order: number
  created_at: string
  updated_at: string
  path_count?: number
  file_count?: number
  prefix_type?: string
}

export interface CollectionPath {
  id: string
  collection_id: string
  path: string
  auto_scan: boolean
  sort_order: number
  created_at: string
}

export interface VirtualFile {
  id: string
  collection_id: string
  path: string
  display_name: string | null
  size: number
  sort_order: number
  created_at: string
  file_name?: string
  mime_type?: string
  mod_time?: string
}

export interface FileEntry {
  id: string
  name: string
  path: string
  size: number
  mod_time: string
  mime_type: string
  source: 'scan' | 'virtual'
  is_dir: boolean
  prefix?: string
  prefix_type?: string
  icon?: string
}

export interface Tag {
  id: string
  name: string
  color: string
}

export interface Prefix {
  prefix: string
  type: string
  map_path: string
  url_template: string
  created_at: string
}

export interface OverviewStats {
  stats: Record<string, number>
}

// Collections
export const api = {
  collections: {
    list: () => request<Collection[]>('/api/collections'),
    get: (id: string) => request<Collection>(`/api/collections/${id}`),
    create: (data: Partial<Collection>) =>
      request<Collection>('/api/collections', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Partial<Collection>) =>
      request<Collection>(`/api/collections/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request<{ status: string }>(`/api/collections/${id}`, { method: 'DELETE' }),
    children: (id: string) =>
      request<Collection[]>(`/api/collections/${id}/children`),
    reorder: (items: { id: string; sort_order: number }[]) =>
      request<{ status: string }>('/api/collections/reorder', {
        method: 'PUT', body: JSON.stringify({ items }),
      }),
  },

  paths: {
    list: (collectionId: string) =>
      request<CollectionPath[]>(`/api/collections/${collectionId}/paths`),
    create: (collectionId: string, data: { path: string; auto_scan?: boolean }) =>
      request<CollectionPath>(`/api/collections/${collectionId}/paths`, {
        method: 'POST', body: JSON.stringify(data),
      }),
    update: (id: string, data: Partial<CollectionPath>) =>
      request<{ status: string }>(`/api/paths/${id}`, {
        method: 'PUT', body: JSON.stringify(data),
      }),
    delete: (id: string) =>
      request<{ status: string }>(`/api/paths/${id}`, { method: 'DELETE' }),
    scan: (id: string) =>
      request<{ path_id: string; count: number }>(`/api/paths/${id}/scan`, { method: 'POST' }),
  },

  vfiles: {
    list: (collectionId: string) =>
      request<VirtualFile[]>(`/api/collections/${collectionId}/vfiles`),
    create: (collectionId: string, data: Partial<VirtualFile>) =>
      request<VirtualFile>(`/api/collections/${collectionId}/vfiles`, {
        method: 'POST', body: JSON.stringify(data),
      }),
    update: (id: string, data: Partial<VirtualFile>) =>
      request<{ status: string }>(`/api/vfiles/${id}`, {
        method: 'PUT', body: JSON.stringify(data),
      }),
    delete: (id: string) =>
      request<{ status: string }>(`/api/vfiles/${id}`, { method: 'DELETE' }),
  },

  browse: {
    list: (collectionId: string) =>
      request<FileEntry[]>(`/api/collections/${collectionId}/browse`),
  },

  preview: {
    content: (path: string) => fetch(`/api/preview/content?path=${encodeURIComponent(path)}`).then(r => r.text()),
    thumbnail: (path: string) => `/api/preview/thumbnail?path=${encodeURIComponent(path)}`,
    openExternal: (path: string) =>
      request<{ status: string }>('/api/open-external', {
        method: 'POST', body: JSON.stringify({ path }),
      }),
  },

  tags: {
    search: (q: string) => request<Tag[]>(`/api/tags/search?q=${encodeURIComponent(q)}`),
    create: (data: { name: string; color?: string }) =>
      request<Tag>('/api/tags', { method: 'POST', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request<{ status: string }>(`/api/tags/${id}`, { method: 'DELETE' }),
    getFileTags: (fileId: string) =>
      request<Tag[]>(`/api/files/${fileId}/tags`),
    setFileTags: (fileId: string, tagIds: string[]) =>
      request<{ status: string }>(`/api/files/${fileId}/tags`, {
        method: 'PUT', body: JSON.stringify({ tag_ids: tagIds }),
      }),
    getCollectionTags: (collectionId: string) =>
      request<Tag[]>(`/api/collections/${collectionId}/tags`),
    setCollectionTags: (collectionId: string, tagIds: string[]) =>
      request<{ status: string }>(`/api/collections/${collectionId}/tags`, {
        method: 'PUT', body: JSON.stringify({ tag_ids: tagIds }),
      }),
  },

  settings: {
    get: (key: string) => request<{ key: string; value: string | null }>(`/api/settings/${key}`),
    set: (key: string, value: string) =>
      request<{ key: string; value: string }>(`/api/settings/${key}`, {
        method: 'PUT', body: JSON.stringify({ value }),
      }),
  },

  prefixes: {
    list: () => request<Prefix[]>('/api/prefixes'),
    update: (prefix: string, data: { type: string; map_path?: string; url_template?: string }) =>
      request<{ status: string }>(`/api/prefixes/${prefix}`, {
        method: 'PUT', body: JSON.stringify(data),
      }),
  },

  overview: {
    stats: () => request<OverviewStats>('/api/overview/stats'),
  },
}
