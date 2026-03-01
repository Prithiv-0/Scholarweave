import { useState, useMemo, useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import apiService, { Paper, SearchParams } from '@/api/client'

const SEARCH_HISTORY_KEY = 'scholarweave_search_history'

const loadSearchHistory = (): string[] => {
    try {
        const raw = localStorage.getItem(SEARCH_HISTORY_KEY)
        if (!raw) return []
        const parsed = JSON.parse(raw)
        if (!Array.isArray(parsed)) return []
        return parsed.filter((entry) => typeof entry === 'string').slice(0, 8)
    } catch {
        return []
    }
}

const saveSearchHistory = (history: string[]) => {
    localStorage.setItem(SEARCH_HISTORY_KEY, JSON.stringify(history.slice(0, 8)))
}

export function useSearch() {
    const [searchParams, setSearchParams] = useSearchParams()

    const [results, setResults] = useState<Paper[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [searched, setSearched] = useState(false)

    const [history, setHistory] = useState<string[]>(loadSearchHistory())
    const [totalPages, setTotalPages] = useState(1)
    const [totalCount, setTotalCount] = useState(0)

    // Form state initialized from URL
    const [query, setQuery] = useState(searchParams.get('q') || '')
    const [sort, setSort] = useState(searchParams.get('sort') || 'relevance')
    const [fromYear, setFromYear] = useState(searchParams.get('from_year') || '')
    const [toYear, setToYear] = useState(searchParams.get('to_year') || '')
    const [paperType, setPaperType] = useState(searchParams.get('type') || '')
    const [subject, setSubject] = useState(searchParams.get('subject') || '')
    const [currentPage, setCurrentPage] = useState(Number(searchParams.get('page')) || 1)
    const [perPage, setPerPage] = useState(Number(searchParams.get('per_page')) || 10)
    const abortControllerRef = useRef<AbortController | null>(null)
    const searchRequestIdRef = useRef(0)

    const avgCitations = useMemo(() => {
        if (results.length === 0) return 0
        const total = results.reduce((sum, paper) => sum + (paper.cited_by_count || 0), 0)
        return Math.round(total / results.length)
    }, [results])

    const runSearch = async (
        overrideParams: Partial<{
            q: string
            page: number
            per_page: number
            sort: string
            from_year: string
            to_year: string
            type: string
            subject: string
        }> = {}
    ) => {
        // Current URL state merged with overrides
        const fetchQ = overrideParams.q !== undefined ? overrideParams.q : query
        if (!fetchQ.trim()) return

        const fetchPage = overrideParams.page !== undefined ? overrideParams.page : currentPage
        const fetchPerPage = overrideParams.per_page !== undefined ? overrideParams.per_page : perPage
        const fetchSort = overrideParams.sort !== undefined ? overrideParams.sort : sort
        const fetchFromYear = overrideParams.from_year !== undefined ? overrideParams.from_year : fromYear
        const fetchToYear = overrideParams.to_year !== undefined ? overrideParams.to_year : toYear
        const fetchType = overrideParams.type !== undefined ? overrideParams.type : paperType
        const fetchSubject = overrideParams.subject !== undefined ? overrideParams.subject : subject

        // Update URL first
        const newParams = new URLSearchParams()
        newParams.set('q', fetchQ.trim())
        if (fetchPage > 1) newParams.set('page', String(fetchPage))
        if (fetchPerPage !== 10) newParams.set('per_page', String(fetchPerPage))
        if (fetchSort !== 'relevance') newParams.set('sort', fetchSort)
        if (fetchFromYear.trim()) newParams.set('from_year', fetchFromYear.trim())
        if (fetchToYear.trim()) newParams.set('to_year', fetchToYear.trim())
        if (fetchType.trim()) newParams.set('type', fetchType.trim())
        if (fetchSubject.trim()) newParams.set('subject', fetchSubject.trim())

        setSearchParams(newParams)

        abortControllerRef.current?.abort()
        const requestId = searchRequestIdRef.current + 1
        searchRequestIdRef.current = requestId
        const controller = new AbortController()
        abortControllerRef.current = controller

        try {
            setLoading(true)
            setError(null)
            const params: SearchParams = {
                q: fetchQ,
                page: fetchPage,
                per_page: fetchPerPage,
                sort: fetchSort as any,
            }

            if (fetchFromYear.trim()) {
                const parsedFromYear = Number(fetchFromYear)
                if (!Number.isNaN(parsedFromYear)) params.from_year = parsedFromYear
            }

            if (fetchToYear.trim()) {
                const parsedToYear = Number(fetchToYear)
                if (!Number.isNaN(parsedToYear)) params.to_year = parsedToYear
            }

            if (fetchType.trim()) params.type = fetchType.trim()
            if (fetchSubject.trim()) params.subject = fetchSubject.trim()

            const response = await apiService.searchPapers(params, { signal: controller.signal })

            if (searchRequestIdRef.current !== requestId) return

            setResults(response.results || [])
            setTotalPages(response.meta?.total_pages || 1)
            setTotalCount(response.meta?.count || response.results.length)
            setSearched(true)

            const normalizedQuery = fetchQ.trim()
            if (normalizedQuery) {
                setHistory((prevHistory) => {
                    const nextHistory = [normalizedQuery, ...prevHistory.filter((item) => item.toLowerCase() !== normalizedQuery.toLowerCase())].slice(0, 8)
                    saveSearchHistory(nextHistory)
                    return nextHistory
                })
            }
        } catch (err: unknown) {
            if ((err as { code?: string })?.code === 'ERR_CANCELED') {
                return
            }
            if (searchRequestIdRef.current !== requestId) return
            const message = err instanceof Error ? err.message : 'Failed to search papers'
            setError(message)
            setResults([])
            setTotalPages(1)
            setTotalCount(0)
            setSearched(true)
        } finally {
            if (searchRequestIdRef.current === requestId) {
                setLoading(false)
                abortControllerRef.current = null
            }
        }
    }

    useEffect(() => {
        return () => {
            abortControllerRef.current?.abort()
        }
    }, [])

    // Load initial search if URL has query
    useEffect(() => {
        if (query && !searched && !loading && results.length === 0) {
            runSearch()
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []) // only on mount

    return {
        state: {
            query,
            sort,
            fromYear,
            toYear,
            paperType,
            subject,
            currentPage,
            perPage,
            results,
            loading,
            error,
            searched,
            history,
            totalPages,
            totalCount,
            avgCitations
        },
        setters: {
            setQuery,
            setSort,
            setFromYear,
            setToYear,
            setPaperType,
            setSubject,
            setCurrentPage,
            setPerPage
        },
        runSearch,
    }
}
