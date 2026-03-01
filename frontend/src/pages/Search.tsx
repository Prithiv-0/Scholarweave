import React, { useEffect, useMemo, useState } from 'react'
import SearchBox from '@/components/SearchBox'
import PaperCard from '@/components/PaperCard'
import KnowledgeGraph from '@/components/KnowledgeGraph'
import GraphNodeDetailPanel from '@/components/GraphNodeDetailPanel'
import AppShell from '@/components/AppShell'
import apiService, { GraphNode, KnowledgeGraph as KnowledgeGraphData } from '@/api/client'
import axios from 'axios'
import { useNavigate } from 'react-router-dom'

import { useSearch } from '@/hooks/useSearch'

export const SearchPage: React.FC = () => {
  const navigate = useNavigate()
  const { state, setters, runSearch } = useSearch()

  const {
    query, sort, fromYear, toYear, paperType, subject,
    currentPage, perPage, results, loading, error, searched,
    history, totalPages, totalCount, avgCitations
  } = state

  const {
    setQuery, setSort, setFromYear, setToYear, setPaperType,
    setSubject, setCurrentPage, setPerPage
  } = setters

  const [viewMode, setViewMode] = useState<'grid' | 'list' | 'graph'>('grid')
  const [graphData, setGraphData] = useState<KnowledgeGraphData | null>(null)
  const [graphLoading, setGraphLoading] = useState(false)
  const [graphError, setGraphError] = useState<string | null>(null)
  const [selectedGraphNode, setSelectedGraphNode] = useState<GraphNode | null>(null)

  useEffect(() => {
    if (viewMode !== 'graph') return

    if (!searched || results.length === 0) {
      setGraphData(null)
      setGraphError(null)
      setGraphLoading(false)
      return
    }

    let cancelled = false

    const loadGraph = async () => {
      try {
        setGraphLoading(true)
        setGraphError(null)
        const graph = await apiService.getSearchGraph(results.map((paper) => paper.id))
        if (cancelled) return
        setGraphData(graph)
        setSelectedGraphNode(null)
      } catch (err: unknown) {
        if (cancelled) return
        setGraphData(null)
        const message = axios.isAxiosError<{ error?: string }>(err)
          ? err.response?.data?.error || err.message || 'Failed to build search graph'
          : err instanceof Error
            ? err.message
            : 'Failed to build search graph'
        setGraphError(message)
        setSelectedGraphNode(null)
      } finally {
        if (!cancelled) {
          setGraphLoading(false)
        }
      }
    }

    loadGraph()

    return () => {
      cancelled = true
    }
  }, [viewMode, results, searched])

  const publicationTypeOptions = [
    { value: '', label: 'Publication type (all)' },
    { value: 'journal-article', label: 'Journal Article' },
    { value: 'proceedings-article', label: 'Conference / Proceedings' },
    { value: 'book-chapter', label: 'Book Chapter' },
    { value: 'preprint', label: 'Preprint' },
    { value: 'report', label: 'Report' },
    { value: 'dataset', label: 'Dataset' },
    { value: 'review', label: 'Review' },
  ]

  const dynamicTopics = useMemo(() => {
    const fallbackTopics = ['Machine Learning', 'Computer Vision', 'NLP', 'Healthcare AI', 'Climate Science']
    const trimmedQuery = query.trim().toLowerCase()

    if (!trimmedQuery && !searched) {
      return fallbackTopics
    }

    const stopWords = new Set([
      'the', 'and', 'for', 'with', 'from', 'this', 'that', 'are', 'was', 'were', 'have', 'has', 'had', 'into', 'over', 'under',
      'using', 'use', 'used', 'via', 'based', 'between', 'across', 'their', 'these', 'those', 'also', 'such', 'than', 'then',
      'paper', 'study', 'method', 'methods', 'approach', 'approaches', 'model', 'models', 'analysis', 'new', 'towards'
    ])

    const tokenize = (text: string): string[] =>
      text
        .toLowerCase()
        .split(/[^a-z0-9+\-]+/)
        .filter((word) => word.length > 2 && !stopWords.has(word))

    const queryWords = new Set(tokenize(trimmedQuery))
    const frequencies = new Map<string, number>()

    results.slice(0, 30).forEach((paper) => {
      tokenize(paper.title || '').forEach((word) => {
        if (queryWords.has(word)) return
        frequencies.set(word, (frequencies.get(word) || 0) + 1)
      })
    })

    const formatPhrase = (value: string) =>
      value
        .split(' ')
        .filter(Boolean)
        .map((part) => (part.length <= 3 ? part.toUpperCase() : part[0].toUpperCase() + part.slice(1)))
        .join(' ')

    const suggestions: string[] = []

    if (trimmedQuery) {
      suggestions.push(formatPhrase(trimmedQuery))
    }

    const topWords = [...frequencies.entries()]
      .sort((a, b) => b[1] - a[1])
      .slice(0, 8)
      .map(([word]) => word)

    if (trimmedQuery && topWords.length > 0) {
      topWords.forEach((word) => {
        suggestions.push(formatPhrase(`${trimmedQuery} ${word}`))
      })
    } else if (topWords.length > 0) {
      topWords.forEach((word) => suggestions.push(formatPhrase(word)))
    }

    if (trimmedQuery && suggestions.length < 5) {
      suggestions.push(
        formatPhrase(`${trimmedQuery} review`),
        formatPhrase(`${trimmedQuery} applications`),
        formatPhrase(`${trimmedQuery} benchmark`)
      )
    }

    const uniqueSuggestions = [...new Set(suggestions)].filter(Boolean)
    return uniqueSuggestions.slice(0, 5).length > 0 ? uniqueSuggestions.slice(0, 5) : fallbackTopics
  }, [query, results, searched])

  const handleSearch = async (searchQuery: string) => {
    setQuery(searchQuery)
    setCurrentPage(1)
    await runSearch({ q: searchQuery, page: 1 })
  }

  const handleApplyFilters = async () => {
    if (!query.trim()) return
    setCurrentPage(1)
    await runSearch({ page: 1 })
  }

  const handlePageChange = async (nextPage: number) => {
    if (!query.trim()) return
    if (nextPage < 1 || nextPage > totalPages || nextPage === currentPage) return
    setCurrentPage(nextPage)
    await runSearch({ page: nextPage })
  }

  const handlePaperClick = async (paperId: string) => {
    navigate(`/papers/${encodeURIComponent(paperId)}`)
  }

  const handleClearFilters = () => {
    setSort('relevance')
    setFromYear('')
    setToYear('')
    setPaperType('')
    setSubject('')
    setPerPage(10)
  }

  const handleExportCSV = () => {
    if (results.length === 0) return
    const lines = [
      'title,doi,cited_by_count,source,id',
      ...results.map((paper) => {
        const safeTitle = `"${(paper.title || '').replace(/"/g, '""')}"`
        const safeDoi = `"${(paper.doi || '').replace(/"/g, '""')}"`
        const safeSource = `"${(paper.source || '').replace(/"/g, '""')}"`
        const safeID = `"${(paper.id || '').replace(/"/g, '""')}"`
        return [safeTitle, safeDoi, paper.cited_by_count ?? 0, safeSource, safeID].join(',')
      }),
    ]
    const blob = new Blob([lines.join('\n')], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `scholarweave-results-${Date.now()}.csv`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }

  return (
    <AppShell
      title="Research Intelligence Workspace"
      subtitle="Enterprise-grade literature discovery, monitoring, and curation"
      rightSlot={
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-white border border-slate-200 rounded-lg p-4">
            <p className="text-xs text-slate-500 uppercase font-semibold tracking-wide">Total Matches</p>
            <p className="text-2xl font-bold text-slate-900 mt-1">{totalCount || 0}</p>
          </div>
          <div className="bg-white border border-slate-200 rounded-lg p-4">
            <p className="text-xs text-slate-500 uppercase font-semibold tracking-wide">Avg Citations (Page)</p>
            <p className="text-2xl font-bold text-slate-900 mt-1">{avgCitations}</p>
          </div>
          <div className="bg-white border border-slate-200 rounded-lg p-4">
            <p className="text-xs text-slate-500 uppercase font-semibold tracking-wide">Current Page</p>
            <p className="text-2xl font-bold text-slate-900 mt-1">{currentPage} / {totalPages}</p>
          </div>
        </div>
      }
    >
      <div className="space-y-8">
        <div className="bg-white border border-slate-200 rounded-lg p-6">
          <SearchBox onSearch={handleSearch} loading={loading} />

          <div className="max-w-5xl mx-auto border-t border-slate-100 pt-5 mt-2">
            <div className="flex flex-wrap gap-2 mb-4">
              {dynamicTopics.map((topic) => (
                <button
                  key={topic}
                  type="button"
                  className="px-3 py-1.5 rounded-full text-xs font-medium bg-slate-100 text-slate-700 hover:bg-slate-200"
                  onClick={async () => {
                    if (query.trim()) {
                      setSubject(topic)
                      await runSearch({ page: 1, subject: topic })
                      return
                    }

                    setQuery(topic)
                    setSubject('')
                    await runSearch({ q: topic, page: 1 })
                  }}
                >
                  {topic}
                </button>
              ))}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              <select
                value={sort}
                onChange={(e) => setSort(e.target.value as 'relevance' | 'citations' | 'date')}
                className="px-3 py-2 rounded border border-slate-300 bg-white"
                disabled={loading}
              >
                <option value="relevance">Sort: Relevance</option>
                <option value="citations">Sort: Citations</option>
                <option value="date">Sort: Date</option>
              </select>

              <select
                value={perPage}
                onChange={(e) => setPerPage(Number(e.target.value))}
                className="px-3 py-2 rounded border border-slate-300 bg-white"
                disabled={loading}
              >
                <option value={10}>10 per page</option>
                <option value={20}>20 per page</option>
                <option value={30}>30 per page</option>
              </select>

              <input
                type="number"
                placeholder="From year"
                aria-label="Filter from year"
                value={fromYear}
                onChange={(e) => setFromYear(e.target.value)}
                className="px-3 py-2 rounded border border-slate-300"
                disabled={loading}
              />

              <input
                type="number"
                placeholder="To year"
                aria-label="Filter to year"
                value={toYear}
                onChange={(e) => setToYear(e.target.value)}
                className="px-3 py-2 rounded border border-slate-300"
                disabled={loading}
              />

              <select
                value={paperType}
                onChange={(e) => setPaperType(e.target.value)}
                className="px-3 py-2 rounded border border-slate-300 bg-white"
                disabled={loading}
              >
                {publicationTypeOptions.map((option) => (
                  <option key={option.value || 'all'} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>

              <input
                type="text"
                placeholder="Subject area"
                aria-label="Filter by subject area"
                value={subject}
                onChange={(e) => setSubject(e.target.value)}
                className="px-3 py-2 rounded border border-slate-300"
                disabled={loading}
              />

              <button
                type="button"
                onClick={handleApplyFilters}
                disabled={loading || !query.trim()}
                className="px-4 py-2 bg-slate-900 text-white rounded hover:bg-slate-800 disabled:bg-slate-400"
              >
                Apply Filters
              </button>
            </div>

            <div className="mt-4 flex flex-wrap gap-2">
              <button
                type="button"
                className="px-3 py-1.5 rounded border border-slate-300 text-sm hover:bg-slate-100"
                onClick={handleClearFilters}
              >
                Clear Filters
              </button>
              <button
                type="button"
                className="px-3 py-1.5 rounded border border-slate-300 text-sm hover:bg-slate-100"
                onClick={handleExportCSV}
                disabled={results.length === 0}
              >
                Export Results (CSV)
              </button>
              <button
                type="button"
                className={`px-3 py-1.5 rounded text-sm ${viewMode === 'grid' ? 'bg-slate-900 text-white' : 'border border-slate-300 hover:bg-slate-100'}`}
                onClick={() => setViewMode('grid')}
              >
                Grid View
              </button>
              <button
                type="button"
                className={`px-3 py-1.5 rounded text-sm ${viewMode === 'list' ? 'bg-slate-900 text-white' : 'border border-slate-300 hover:bg-slate-100'}`}
                onClick={() => setViewMode('list')}
              >
                List View
              </button>
              <button
                type="button"
                className={`px-3 py-1.5 rounded text-sm ${viewMode === 'graph' ? 'bg-slate-900 text-white' : 'border border-slate-300 hover:bg-slate-100'}`}
                onClick={() => setViewMode('graph')}
                disabled={results.length === 0}
              >
                Graph View
              </button>
            </div>

            {history.length > 0 && (
              <div className="mt-4">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500 mb-2">Recent Searches</p>
                <div className="flex flex-wrap gap-2">
                  {history.map((item) => (
                    <button
                      key={item}
                      type="button"
                      onClick={async () => {
                        setQuery(item)
                        await runSearch({ q: item, page: 1 })
                      }}
                      className="px-3 py-1.5 rounded-full text-xs bg-slate-100 text-slate-700 hover:bg-slate-200"
                    >
                      {item}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {error && (
          <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700 font-semibold">Error</p>
            <p className="text-red-600 text-sm">{error}</p>
          </div>
        )}

        {loading && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="bg-gray-200 rounded-lg h-64 animate-pulse"></div>
            ))}
          </div>
        )}

        {!loading && searched && results.length > 0 && (
          <div>
            <h2 className="text-2xl font-bold text-gray-900 mb-6">
              Results for "{query}" ({totalCount} papers)
            </h2>

            {viewMode === 'grid' ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {results.map((paper) => (
                  <PaperCard
                    key={paper.id}
                    paper={paper}
                    onClick={() => handlePaperClick(paper.id)}
                  />
                ))}
              </div>
            ) : viewMode === 'list' ? (
              <div className="space-y-3">
                {results.map((paper) => (
                  <button
                    type="button"
                    key={paper.id}
                    onClick={() => handlePaperClick(paper.id)}
                    className="w-full text-left bg-white border border-slate-200 rounded-lg p-4 hover:border-slate-300"
                  >
                    <p className="font-semibold text-slate-900 mb-1">{paper.title}</p>
                    <p className="text-sm text-slate-600 line-clamp-2">{paper.abstract || 'No abstract available'}</p>
                    <div className="mt-2 text-xs text-slate-500 flex items-center gap-4">
                      <span>Source: {paper.source || 'Unknown'}</span>
                      <span>Citations: {paper.cited_by_count ?? 0}</span>
                    </div>
                  </button>
                ))}
              </div>
            ) : (
              <div className="space-y-4">
                <KnowledgeGraph
                  graph={graphData}
                  loading={graphLoading}
                  error={graphError}
                  height={640}
                  selectedNodeId={selectedGraphNode?.id}
                  onNodeSelect={setSelectedGraphNode}
                />

                {selectedGraphNode && (
                  <GraphNodeDetailPanel
                    node={selectedGraphNode}
                    title="Selected Node"
                    onPivotPaper={(paperId) => navigate(`/graph?paper=${encodeURIComponent(paperId)}`)}
                    onPivotAuthor={(authorId) => navigate(`/graph?author=${encodeURIComponent(authorId)}`)}
                    onPivotConcept={(conceptId) => navigate(`/graph?concept=${encodeURIComponent(conceptId)}`)}
                  />
                )}
              </div>
            )}

            <div className="mt-8 flex items-center justify-center gap-3">
              <button
                type="button"
                onClick={() => handlePageChange(currentPage - 1)}
                disabled={loading || currentPage <= 1}
                className="px-4 py-2 rounded border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                Previous
              </button>
              <span className="text-sm text-gray-700">
                Page {currentPage} of {totalPages}
              </span>
              <button
                type="button"
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={loading || currentPage >= totalPages}
                className="px-4 py-2 rounded border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                Next
              </button>
            </div>
          </div>
        )}

        {/* No Results */}
        {!loading && searched && results.length === 0 && !error && (
          <div className="text-center py-12">
            <p className="text-gray-600 text-lg mb-4">No papers found for "{query}"</p>
            <p className="text-gray-500">Try adjusting your search query</p>
          </div>
        )}

        {/* Initial State */}
        {!searched && (
          <div className="text-center py-14 bg-white border border-slate-200 rounded-lg">
            <p className="text-slate-700 text-lg">Start with a query to explore global research intelligence</p>
            <p className="text-slate-500 text-sm mt-2">Use filters, sorting, and quick topics to narrow high-signal papers fast.</p>
          </div>
        )}
      </div>
    </AppShell>
  )
}

export default SearchPage
