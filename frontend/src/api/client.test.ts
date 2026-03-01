import { describe, expect, it } from 'vitest'
import { buildApiUrl, resolveDoiUrl } from './client'

describe('api client helpers', () => {
  it('resolves raw doi to full doi url', () => {
    expect(resolveDoiUrl('10.1000/xyz123')).toBe('https://doi.org/10.1000/xyz123')
  })

  it('keeps full doi urls unchanged', () => {
    expect(resolveDoiUrl('https://doi.org/10.1000/xyz123')).toBe('https://doi.org/10.1000/xyz123')
  })

  it('builds full api url with normalized slashes', () => {
    expect(buildApiUrl('papers/search')).toMatch(/\/api\/v1\/papers\/search$/)
    expect(buildApiUrl('/health')).toMatch(/\/api\/v1\/health$/)
  })
})
