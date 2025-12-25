import React from 'react'
import { Paper } from '@/api/client'

interface PaperCardProps {
  paper: Paper
  onClick?: () => void
}

export const PaperCard: React.FC<PaperCardProps> = ({ paper, onClick }) => {
  const authors = paper.authors ?? []

  return (
    <div
      onClick={onClick}
      className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow cursor-pointer border border-gray-200"
    >
      <h3 className="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
        {paper.title ?? 'Untitled'}
      </h3>
      <p className="text-sm text-gray-600 mb-3 line-clamp-3">
        {paper.abstract ? (paper.abstract.length > 160 ? paper.abstract.slice(0, 157) + '...' : paper.abstract) : 'No abstract available'}
      </p>
      <div className="flex items-center justify-between text-xs text-gray-500">
        <span className="font-medium text-blue-600">{paper.source ?? 'Unknown'}</span>
        <span>Cited by: {paper.cited_by_count ?? 0}</span>
      </div>
      {authors.length > 0 && (
        <div className="mt-3 pt-3 border-t border-gray-100">
          <p className="text-xs text-gray-600">
            <span className="font-medium">Authors: </span>
            {authors.slice(0, 2).map(a => a?.name || '').join(', ')}
            {authors.length > 2 && ` +${authors.length - 2}`}
          </p>
        </div>
      )}
      <div className="mt-3 pt-3 border-t border-gray-100 text-xs text-gray-600 flex items-center justify-between">
        <div>DOI: {paper.doi ? <a className="text-blue-600 hover:underline" href={`https://doi.org/${paper.doi}`} target="_blank" rel="noreferrer">{paper.doi}</a> : 'N/A'}</div>
        <div>Cit: {paper.cited_by_count ?? 0}</div>
      </div>
    </div>
  )
}

export default PaperCard
