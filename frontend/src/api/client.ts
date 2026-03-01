import axios, { AxiosInstance } from 'axios'

const trimTrailingSlash = (value: string): string => value.replace(/\/+$/, '')

const normalizeApiBaseUrl = (value?: string): string => {
  const raw = (value || '/api/v1').trim()
  const normalized = trimTrailingSlash(raw)

  if (/\/api$/i.test(normalized)) {
    return `${normalized}/v1`
  }

  return normalized
}

const API_BASE_URL = normalizeApiBaseUrl(import.meta.env.VITE_API_BASE_URL)
export const AUTH_TOKEN_KEY = 'scholarweave_auth_token'

const client: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
})

export const buildApiUrl = (path: string): string => {
  const normalizedBase = trimTrailingSlash(API_BASE_URL)
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizedBase}${normalizedPath}`
}

export const resolveDoiUrl = (doi?: string): string | null => {
  if (!doi) return null
  if (/^https?:\/\//i.test(doi)) return doi
  const normalizedDoi = doi.replace(/^doi:\s*/i, '')
  return `https://doi.org/${normalizedDoi}`
}

export interface Author {
  name: string
  id?: string
  orcid?: string
}

export interface Paper {
  id: string
  title: string
  abstract: string
  doi: string
  authors: Author[]
  cited_by_count: number
  source: string
  publication_date?: string
  type?: string
}

export interface SearchResponse {
  results: Paper[]
  meta: {
    count: number
    page: number
    per_page: number
    total_pages: number
    sort: string
  }
}

export interface SearchParams {
  q: string
  page?: number
  per_page?: number
  sort?: 'relevance' | 'citations' | 'date'
  from_year?: number
  to_year?: number
  type?: string
  subject?: string
}

interface RequestOptions {
  signal?: AbortSignal
}

export interface HealthResponse {
  status: string
  version: string
  timestamp: string
  services: {
    api: string
    openalex: string
  }
}

export interface User {
  id: string
  name: string
  email: string
  created_at: string
}

export interface AuthRequest {
  name?: string
  email: string
  password: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface SavedPaper {
  paper_id: string
  title: string
  doi: string
  saved_at: string
  cited_by_count: number
  source: string
}

export interface ReadingList {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
  papers: SavedPaper[]
}

export type GraphNodeType = 'paper' | 'author' | 'concept' | 'institution'
export type GraphEdgeType = 'authored' | 'has_concept' | 'cites' | 'related' | 'affiliated'

export interface GraphNode {
  id: string
  label: string
  type: GraphNodeType
  properties?: Record<string, any>
}

export interface GraphEdge {
  source: string
  target: string
  relationship: GraphEdgeType
  weight: number
}

export interface KnowledgeGraph {
  nodes: GraphNode[]
  edges: GraphEdge[]
  center_node_id: string
}

export const authStorage = {
  getToken: (): string | null => localStorage.getItem(AUTH_TOKEN_KEY),
  setToken: (token: string) => localStorage.setItem(AUTH_TOKEN_KEY, token),
  clearToken: () => localStorage.removeItem(AUTH_TOKEN_KEY),
}

const authHeaders = (): Record<string, string> => {
  const token = authStorage.getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

const normalizeId = (id: string): string => {
  try {
    return encodeURIComponent(decodeURIComponent(id))
  } catch {
    return encodeURIComponent(id)
  }
}

export const apiService = {
  health: async (): Promise<HealthResponse> => {
    const { data } = await client.get<HealthResponse>('/health')
    return data
  },

  searchPapers: async (searchParams: SearchParams, options?: RequestOptions): Promise<SearchResponse> => {
    const params = Object.fromEntries(
      Object.entries(searchParams).filter(([, value]) => value !== undefined && value !== null && value !== '')
    )
    const { data } = await client.get<SearchResponse>('/papers/search', {
      params,
      signal: options?.signal,
    })
    return data
  },

  getPaperById: async (id: string): Promise<Paper> => {
    const { data } = await client.get<Paper>(`/papers/${normalizeId(id)}`)
    return data
  },

  register: async (payload: AuthRequest): Promise<AuthResponse> => {
    const { data } = await client.post<AuthResponse>('/auth/register', payload)
    return data
  },

  login: async (payload: AuthRequest): Promise<AuthResponse> => {
    const { data } = await client.post<AuthResponse>('/auth/login', payload)
    return data
  },

  me: async (): Promise<User> => {
    const { data } = await client.get<User>('/users/me', {
      headers: authHeaders(),
    })
    return data
  },

  listFavorites: async (): Promise<SavedPaper[]> => {
    const { data } = await client.get<{ favorites: SavedPaper[] }>('/users/me/favorites', {
      headers: authHeaders(),
    })
    return data.favorites || []
  },

  addFavorite: async (paper: Paper): Promise<SavedPaper> => {
    const payload = {
      paper_id: paper.id,
      title: paper.title,
      doi: paper.doi,
      cited_by_count: paper.cited_by_count,
      source: paper.source,
    }
    const { data } = await client.post<SavedPaper>('/users/me/favorites', payload, {
      headers: authHeaders(),
    })
    return data
  },

  removeFavorite: async (paperId: string): Promise<void> => {
    await client.delete(`/users/me/favorites/${encodeURIComponent(paperId)}`, {
      headers: authHeaders(),
    })
  },

  listReadingLists: async (): Promise<ReadingList[]> => {
    const { data } = await client.get<{ lists: ReadingList[] }>('/users/me/lists', {
      headers: authHeaders(),
    })
    return data.lists || []
  },

  createReadingList: async (name: string, description: string): Promise<ReadingList> => {
    const { data } = await client.post<ReadingList>('/users/me/lists', { name, description }, {
      headers: authHeaders(),
    })
    return data
  },

  addPaperToReadingList: async (listId: string, paper: SavedPaper): Promise<ReadingList> => {
    const payload = {
      paper_id: paper.paper_id,
      title: paper.title,
      doi: paper.doi,
      cited_by_count: paper.cited_by_count,
      source: paper.source,
    }
    const { data } = await client.post<ReadingList>(`/users/me/lists/${encodeURIComponent(listId)}/items`, payload, {
      headers: authHeaders(),
    })
    return data
  },

  removePaperFromReadingList: async (listId: string, paperId: string): Promise<ReadingList> => {
    const { data } = await client.delete<ReadingList>(`/users/me/lists/${encodeURIComponent(listId)}/items/${encodeURIComponent(paperId)}`, {
      headers: authHeaders(),
    })
    return data
  },

  getPaperGraph: async (id: string, depth: 1 | 2 = 1): Promise<KnowledgeGraph> => {
    const { data } = await client.get<KnowledgeGraph>(`/graph/paper/${normalizeId(id)}`, {
      params: { depth },
    })
    return data
  },

  getAuthorGraph: async (id: string): Promise<KnowledgeGraph> => {
    const { data } = await client.get<KnowledgeGraph>(`/graph/author/${normalizeId(id)}`)
    return data
  },

  getConceptGraph: async (id: string): Promise<KnowledgeGraph> => {
    const { data } = await client.get<KnowledgeGraph>(`/graph/concept/${normalizeId(id)}`)
    return data
  },

  getSearchGraph: async (paperIds: string[]): Promise<KnowledgeGraph> => {
    const normalizedPaperIds = paperIds.map((paperId) => {
      try {
        return decodeURIComponent(paperId)
      } catch {
        return paperId
      }
    }).filter(Boolean)

    const { data } = await client.post<KnowledgeGraph>('/graph/search', {
      paper_ids: normalizedPaperIds,
    })
    return data
  }
}

export default apiService
