import React, { useState } from 'react'
import SearchBox from '@/components/SearchBox'
import PaperCard from '@/components/PaperCard'
import Health from '@/components/Health'
import apiService, { Paper } from '@/api/client'
import { useNavigate } from 'react-router-dom'

export const SearchPage: React.FC = () => {
  const [results, setResults] = useState<Paper[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [searched, setSearched] = useState(false)
  const [query, setQuery] = useState('')
  const navigate = useNavigate()

  const handleSearch = async (searchQuery: string) => {
    try {
      setLoading(true)
      setError(null)
      setQuery(searchQuery)
      const response = await apiService.searchPapers(searchQuery)

      // Normalize results to the frontend `Paper` shape to avoid runtime errors
      const rawResults: any[] = response.results || []
      const papers = rawResults.map((r: any) => {
        // authors may be in different shapes depending on the API
        const rawAuthors = r.authors || r.authorships || r.authorships || []
        const authors = rawAuthors.map((a: any) => {
          // OpenAlex nested shape: { author: { display_name: '...' } }
          if (a && a.author && a.author.display_name) return { name: a.author.display_name }
          // sometimes author objects are flat { name }
          if (a && (a.name || a.display_name)) return { name: a.name || a.display_name }
          // fallback: string or unknown
          if (typeof a === 'string') return { name: a }
          return { name: 'Unknown' }
        })

        return {
          id: r.id || r.doi || (typeof r === 'string' ? r : ''),
          title: r.title || r.display_name || '',
          abstract: r.abstract || r.summary || '',
          doi: r.doi || '',
          authors,
          cited_by_count: r.cited_by_count || 0,
          source: r.source || 'OpenAlex',
        }
      })

      setResults(papers)
      setSearched(true)
    } catch (err: any) {
      setError(err.message || 'Failed to search papers')
      setResults([])
      setSearched(true)
    } finally {
      setLoading(false)
    }
  }

  const handlePaperClick = async (paperId: string) => {
    // navigate to the paper detail route
    navigate(`/papers/${encodeURIComponent(paperId)}`)
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">ScholarWeave</h1>
          <p className="text-gray-600">Search and discover academic papers from OpenAlex</p>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-12">
        {/* Health Status */}
        <div className="mb-8">
          <Health />
        </div>

        {/* Search Box */}
        <div className="mb-12">
          <SearchBox onSearch={handleSearch} loading={loading} />
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-8 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700 font-semibold">Error</p>
            <p className="text-red-600 text-sm">{error}</p>
          </div>
        )}

        {/* Loading State */}
        {loading && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="bg-gray-200 rounded-lg h-64 animate-pulse"></div>
            ))}
          </div>
        )}

        {/* Results */}
        {!loading && searched && results.length > 0 && (
          <div>
            <h2 className="text-2xl font-bold text-gray-900 mb-6">
              Results for "{query}" ({results.length} papers)
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {results.map((paper) => (
                <div
                  key={paper.id}
                  onClick={() => handlePaperClick(paper.id)}
                  className="cursor-pointer"
                >
                  <PaperCard paper={paper} />
                </div>
              ))}
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
          <div className="text-center py-12">
            <p className="text-gray-600 text-lg">Enter a search term to get started</p>
            <p className="text-gray-500 text-sm mt-2">Search by title, author, topic, or keywords</p>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-6 py-6 text-center text-gray-600 text-sm">
          <p>Powered by <a href="https://openalex.org" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">OpenAlex</a></p>
        </div>
      </footer>
    </div>
  )
}

export default SearchPage
