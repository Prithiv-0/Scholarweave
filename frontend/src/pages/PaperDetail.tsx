import React, { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import apiService, { Paper } from '@/api/client'

const PaperDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const [paper, setPaper] = useState<Paper | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    const fetchPaper = async () => {
      try {
        setLoading(true)
        const data = await apiService.getPaperById(id)
        setPaper(data)
        setError(null)
      } catch (err: any) {
        setError(err.message || 'Failed to load paper')
        setPaper(null)
      } finally {
        setLoading(false)
      }
    }
    fetchPaper()
  }, [id])

  if (loading) return <div className="p-8">Loading paper...</div>
  if (error) return (
    <div className="p-8">
      <p className="text-red-600">{error}</p>
      <Link to="/" className="text-blue-600 underline">Back to search</Link>
    </div>
  )

  if (!paper) return <div className="p-8">Paper not found.</div>

  return (
    <div className="max-w-4xl mx-auto p-8">
      <Link to="/" className="text-sm text-blue-600 hover:underline">← Back to search</Link>
      <h1 className="text-3xl font-bold mt-4 mb-2">{paper.title}</h1>
      <p className="text-sm text-gray-600 mb-4">{paper.authors?.map(a => a.name).join(', ')}</p>

      {paper.doi && (
        <p className="mb-3"><strong>DOI:</strong> <a className="text-blue-600" href={`https://doi.org/${paper.doi}`} target="_blank" rel="noreferrer">{paper.doi}</a></p>
      )}

      <p className="text-sm text-gray-700 mb-6 whitespace-pre-wrap">{paper.abstract || 'No abstract available'}</p>

      <div className="text-sm text-gray-600">
        <p><strong>Cited by:</strong> {paper.cited_by_count ?? 0}</p>
        <p><strong>Source:</strong> {paper.source}</p>
      </div>

      <div className="mt-6">
        <a className="inline-block px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700" href={paper.doi ? `https://doi.org/${paper.doi}` : '#'} target="_blank" rel="noreferrer">Open DOI</a>
      </div>
    </div>
  )
}

export default PaperDetail
