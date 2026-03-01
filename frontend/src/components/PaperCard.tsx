import React from 'react'
import { Paper, resolveDoiUrl } from '@/api/client'

interface PaperCardProps {
  paper: Paper
  onClick?: () => void
}

export const PaperCard: React.FC<PaperCardProps> = ({ paper, onClick }) => {
  const authors = paper.authors ?? []
  const doiUrl = resolveDoiUrl(paper.doi)

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') onClick?.() }}
      className="bg-white rounded-xl p-6 hover:shadow-md transition-shadow cursor-pointer border border-slate-200 hover:border-slate-300"
    >
      <h3 className="text-lg font-semibold text-slate-900 mb-2 line-clamp-2">
        {paper.title ?? 'Untitled'}
      </h3>
      <p className="text-sm text-slate-600 mb-3 line-clamp-3">
        {paper.abstract || 'No abstract available'}
      </p>
      <div className="text-xs text-slate-500 mb-2 font-medium">
        {paper.source ?? 'Unknown source'}
      </div>
      {authors.length > 0 && (
        <div className="mt-3 pt-3 border-t border-slate-100">
          <p className="text-xs text-slate-600">
            <span className="font-medium">Authors: </span>
            {authors.slice(0, 2).map(a => a?.name || '').join(', ')}
            {authors.length > 2 && ` +${authors.length - 2}`}
          </p>
        </div>
      )}
      <div className="mt-3 pt-3 border-t border-slate-100 text-xs text-slate-600 flex items-center justify-between">
        <div>DOI: {doiUrl ? <a className="text-slate-800 hover:underline" href={doiUrl} target="_blank" rel="noreferrer" onClick={(e) => e.stopPropagation()}>{paper.doi}</a> : 'N/A'}</div>
        <div>Cit: {paper.cited_by_count ?? 0}</div>
      </div>
    </div>
  )
}

export default PaperCard
