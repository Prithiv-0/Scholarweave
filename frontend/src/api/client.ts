import axios, { AxiosInstance } from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3000/api/v1'

const client: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
})

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
}

export interface SearchResponse {
  results: Paper[]
  meta?: {
    count: number
  }
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

export const apiService = {
  health: async (): Promise<HealthResponse> => {
    const { data } = await client.get<HealthResponse>('/health')
    return data
  },

  searchPapers: async (query: string): Promise<SearchResponse> => {
    const { data } = await client.get<SearchResponse>('/papers/search', {
      params: { q: query }
    })
    return data
  },

  getPaperById: async (id: string): Promise<Paper> => {
    // encode ID so slashes or special characters don't break the backend route
    const safeId = encodeURIComponent(id)
    const { data } = await client.get<Paper>(`/papers/${safeId}`)
    return data
  }
}

export default apiService
