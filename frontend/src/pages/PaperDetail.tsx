import React, { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import AppShell from '@/components/AppShell'
import apiService, { authStorage, Paper, buildApiUrl, resolveDoiUrl } from '@/api/client'

const PaperDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const [paper, setPaper] = useState<Paper | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [favoriteMessage, setFavoriteMessage] = useState<string | null>(null)
  const doiUrl = resolveDoiUrl(paper?.doi)

  const handleSaveFavorite = async () => {
    if (!paper) return
    if (!authStorage.getToken()) {
      setFavoriteMessage('Sign in to save favorites')
      return
    }

    try {
      await apiService.addFavorite(paper)
      setFavoriteMessage('Saved to favorites')
      setTimeout(() => setFavoriteMessage(null), 3000)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to save favorite'
      setFavoriteMessage(message)
      setTimeout(() => setFavoriteMessage(null), 5000)
    }
  }

  useEffect(() => {
    if (!id) return
    const fetchPaper = async () => {
      try {
        setLoading(true)
        const data = await apiService.getPaperById(id)
        setPaper(data)
        setError(null)
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : 'Failed to load paper'
        setError(message)
        setPaper(null)
      } finally {
        setLoading(false)
      }
    }
    fetchPaper()
  }, [id])

  if (loading) return <AppShell title="Paper Detail"><div>Loading paper...</div></AppShell>
  if (error) return (
    <AppShell title="Paper Detail">
      <div>
        <p className="text-red-600">{error}</p>
        <Link to="/" className="text-blue-600 underline">Back to search</Link>
      </div>
    </AppShell>
  )

  if (!paper) return <AppShell title="Paper Detail"><div>Paper not found.</div></AppShell>

  return (
    <AppShell title="Paper Detail" subtitle="Review metadata, export semantic data, and save to your library">
      <div className="max-w-4xl">
        {/* JSON-LD Structured Data */}
        <script type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "ScholarlyArticle",
              "headline": paper.title,
              "abstract": paper.abstract,
              "author": paper.authors?.map((opt) => ({
                "@type": "Person",
                "name": opt.name
              })),
              "datePublished": paper.publication_date || new Date().getFullYear().toString(),
              "sameAs": doiUrl || undefined
            }).replace(/<\//g, '<\\/')
          }}
        />

        <Link to="/" className="text-sm text-blue-600 hover:underline">← Back to search</Link>
        <h1 className="text-3xl font-bold mt-4 mb-2">{paper.title}</h1>
        <p className="text-sm text-gray-600 mb-4">{paper.authors?.map((a) => a.name).join(', ')}</p>

        {doiUrl && (
          <p className="mb-3"><strong>DOI:</strong> <a className="text-blue-600" href={doiUrl} target="_blank" rel="noreferrer">{paper.doi}</a></p>
        )}

        <p className="text-sm text-gray-700 mb-6 whitespace-pre-wrap">{paper.abstract || 'No abstract available'}</p>

        <div className="text-sm text-gray-600">
          <p><strong>Cited by:</strong> {paper.cited_by_count ?? 0}</p>
          <p><strong>Source:</strong> {paper.source}</p>
        </div>

        <div className="mt-6 flex gap-4">
          <button
            type="button"
            onClick={handleSaveFavorite}
            className="inline-block px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
          >
            Save to favorites
          </button>
          <Link className="inline-block px-4 py-2 bg-slate-900 text-white rounded hover:bg-slate-800" to={`/graph?paper=${encodeURIComponent(paper.id)}`}>Explore Knowledge Graph</Link>
          {doiUrl && (
            <a className="inline-block px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700" href={doiUrl} target="_blank" rel="noreferrer">Open DOI</a>
          )}
          <a className="inline-block px-4 py-2 border border-gray-300 text-gray-700 rounded hover:bg-gray-50" href={buildApiUrl(`/papers/${encodeURIComponent(paper.id)}/rdf`)} target="_blank" rel="noreferrer">Export RDF</a>
        </div>

        {favoriteMessage && (
          <p className="mt-3 text-sm text-gray-700">{favoriteMessage}</p>
        )}

        <div className="mt-3">
          <Link to="/library" className="text-sm text-indigo-700 hover:underline">Go to library →</Link>
        </div>
      </div>
    </AppShell>
  )
}

export default PaperDetail
