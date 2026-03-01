import React from 'react'
import { Link } from 'react-router-dom'
import { GraphNode } from '@/api/client'

interface GraphNodeDetailPanelProps {
  node: GraphNode | null
  title?: string
  className?: string
  onPivotPaper?: (paperId: string) => void | Promise<void>
  onPivotAuthor?: (authorId: string) => void | Promise<void>
  onPivotConcept?: (conceptId: string) => void | Promise<void>
}

const GraphNodeDetailPanel: React.FC<GraphNodeDetailPanelProps> = ({
  node,
  title = 'Node Details',
  className = '',
  onPivotPaper,
  onPivotAuthor,
  onPivotConcept,
}) => {
  return (
    <div className={`bg-white border border-slate-200 rounded-lg p-4 ${className}`.trim()}>
      <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500 mb-3">{title}</h2>

      {!node && <p className="text-sm text-slate-600">Select a node in the graph to inspect metadata.</p>}

      {node && (
        <div className="space-y-3 text-sm">
          <p className="font-semibold text-slate-900 break-words">{node.label}</p>
          <p className="text-slate-600">Type: {node.type}</p>
          <p className="text-slate-600 break-all">ID: {node.id}</p>

          {node.properties && (
            <div className="rounded border border-slate-200 bg-slate-50 p-3 text-xs text-slate-700 space-y-1">
              {Object.entries(node.properties).map(([key, value]) => (
                <p key={key}>
                  <span className="font-semibold">{key}</span>: {String(value)}
                </p>
              ))}
            </div>
          )}

          <div className="flex flex-wrap gap-2">
            {node.type === 'paper' && (
              <>
                <Link
                  to={`/papers/${encodeURIComponent(node.id)}`}
                  className="inline-flex rounded bg-slate-900 px-3 py-2 text-white hover:bg-slate-800"
                >
                  Open Paper
                </Link>
                <button
                  type="button"
                  className="inline-flex rounded border border-slate-300 px-3 py-2 text-slate-700 hover:bg-slate-100"
                  onClick={() => onPivotPaper?.(node.id)}
                >
                  Pivot Paper Graph
                </button>
              </>
            )}

            {node.type === 'author' && (
              <button
                type="button"
                className="inline-flex rounded bg-slate-900 px-3 py-2 text-white hover:bg-slate-800"
                onClick={() => onPivotAuthor?.(node.id)}
              >
                Pivot Author Graph
              </button>
            )}

            {node.type === 'concept' && (
              <button
                type="button"
                className="inline-flex rounded bg-slate-900 px-3 py-2 text-white hover:bg-slate-800"
                onClick={() => onPivotConcept?.(node.id)}
              >
                Pivot Concept Graph
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default GraphNodeDetailPanel
