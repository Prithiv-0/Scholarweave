import React, { useEffect, useMemo, useRef, useState } from 'react'
import axios from 'axios'
import { useSearchParams } from 'react-router-dom'
import AppShell from '@/components/AppShell'
import KnowledgeGraph from '@/components/KnowledgeGraph'
import GraphNodeDetailPanel from '@/components/GraphNodeDetailPanel'
import apiService, { GraphNode, KnowledgeGraph as KnowledgeGraphData } from '@/api/client'

type GraphContextType = 'paper' | 'author' | 'concept'

interface GraphBreadcrumb {
  type: GraphContextType
  id: string
}

const KnowledgeGraphPage: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const initialPaperId = searchParams.get('paper') || ''
  const initialAuthorId = searchParams.get('author') || ''
  const initialConceptId = searchParams.get('concept') || ''

  const initialContextType: GraphContextType = initialAuthorId ? 'author' : initialConceptId ? 'concept' : 'paper'
  const initialTargetId = initialAuthorId || initialConceptId || initialPaperId

  const [contextType, setContextType] = useState<GraphContextType>(initialContextType)
  const [targetId, setTargetId] = useState(initialTargetId)
  const [depth, setDepth] = useState<1 | 2>(1)
  const [graph, setGraph] = useState<KnowledgeGraphData | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null)
  const [breadcrumbs, setBreadcrumbs] = useState<GraphBreadcrumb[]>([])
  const loadRequestIdRef = useRef(0)

  const loadGraph = async (targetType: GraphContextType, targetValue: string, targetDepth: 1 | 2) => {
    const requestId = loadRequestIdRef.current + 1
    loadRequestIdRef.current = requestId
    const normalized = targetValue.trim()
    if (!normalized) {
      const labels: Record<GraphContextType, string> = {
        paper: 'paper ID',
        author: 'author ID',
        concept: 'concept ID',
      }
      setError(`Enter a ${labels[targetType]} to load the knowledge graph.`)
      setGraph(null)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const data =
        targetType === 'paper'
          ? await apiService.getPaperGraph(normalized, targetDepth)
          : targetType === 'author'
            ? await apiService.getAuthorGraph(normalized)
            : await apiService.getConceptGraph(normalized)
      if (loadRequestIdRef.current !== requestId) return
      setGraph(data)
      setSelectedNode(null)
      setBreadcrumbs((prev) => {
        const next = [...prev]
        const latest = next[next.length - 1]
        if (!latest || latest.type !== targetType || latest.id !== normalized) {
          next.push({ type: targetType, id: normalized })
        }
        return next.slice(-8)
      })
    } catch (err: unknown) {
      if (loadRequestIdRef.current !== requestId) return
      setGraph(null)
      const message = axios.isAxiosError<{ error?: string }>(err)
        ? err.response?.data?.error || err.message || 'Failed to load knowledge graph'
        : err instanceof Error
          ? err.message
          : 'Failed to load knowledge graph'
      setError(message)
    } finally {
      if (loadRequestIdRef.current === requestId) {
        setLoading(false)
      }
    }
  }

  useEffect(() => {
    if (!initialTargetId) return
    loadGraph(initialContextType, initialTargetId, depth)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const depthDisabled = contextType !== 'paper'

  const contextLabel = useMemo(() => {
    if (contextType === 'author') return 'Author ID'
    if (contextType === 'concept') return 'Concept ID'
    return 'Paper ID'
  }, [contextType])

  const contextPlaceholder = useMemo(() => {
    if (contextType === 'author') return 'e.g. A1969205032'
    if (contextType === 'concept') return 'e.g. C119857082'
    return 'e.g. W2741809807'
  }, [contextType])

  const graphStats = useMemo(() => {
    return {
      nodes: graph?.nodes.length || 0,
      edges: graph?.edges.length || 0,
    }
  }, [graph])

  const buildSearchParams = (type: GraphContextType, id: string) => ({ [type]: id })

  const copyShareLink = async () => {
    const currentParams = new URLSearchParams()
    const normalizedTarget = targetId.trim()
    if (normalizedTarget) {
      currentParams.set(contextType, normalizedTarget)
    }
    if (contextType === 'paper') {
      currentParams.set('depth', String(depth))
    }
    const url = `${window.location.origin}/graph${currentParams.toString() ? `?${currentParams.toString()}` : ''}`
    try {
      await navigator.clipboard.writeText(url)
    } catch {
      // Fallback: prompt user with the URL
      window.prompt('Copy this link:', url)
    }
  }

  return (
    <AppShell
      title="Knowledge Graph Explorer"
      subtitle="Visualize connections among papers, authors, institutions, and concepts"
      rightSlot={
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-white border border-slate-200 rounded-lg p-4">
            <p className="text-xs text-slate-500 uppercase font-semibold tracking-wide">Nodes</p>
            <p className="text-2xl font-bold text-slate-900 mt-1">{graphStats.nodes}</p>
          </div>
          <div className="bg-white border border-slate-200 rounded-lg p-4">
            <p className="text-xs text-slate-500 uppercase font-semibold tracking-wide">Edges</p>
            <p className="text-2xl font-bold text-slate-900 mt-1">{graphStats.edges}</p>
          </div>
          <div className="bg-white border border-slate-200 rounded-lg p-4">
            <p className="text-xs text-slate-500 uppercase font-semibold tracking-wide">Depth</p>
            <p className="text-2xl font-bold text-slate-900 mt-1">{depth}</p>
          </div>
        </div>
      }
    >
      <div className="space-y-6">
        <div className="bg-white border border-slate-200 rounded-lg p-5">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-3 items-end">
            <div className="lg:col-span-7">
              <label className="text-sm font-medium text-slate-700">{contextLabel}</label>
              <input
                type="text"
                value={targetId}
                onChange={(e) => setTargetId(e.target.value)}
                placeholder={contextPlaceholder}
                className="mt-1 w-full rounded border border-slate-300 px-3 py-2"
              />
            </div>

            <div className="lg:col-span-2">
              <label className="text-sm font-medium text-slate-700">Context</label>
              <select
                value={contextType}
                onChange={(e) => setContextType(e.target.value as GraphContextType)}
                className="mt-1 w-full rounded border border-slate-300 px-3 py-2 bg-white"
              >
                <option value="paper">Paper</option>
                <option value="author">Author</option>
                <option value="concept">Concept</option>
              </select>
            </div>

            <div className="lg:col-span-2">
              <label className="text-sm font-medium text-slate-700">Depth</label>
              <select
                value={depth}
                onChange={(e) => setDepth(Number(e.target.value) as 1 | 2)}
                className="mt-1 w-full rounded border border-slate-300 px-3 py-2 bg-white"
                disabled={depthDisabled}
              >
                <option value={1}>Depth 1</option>
                <option value={2}>Depth 2</option>
              </select>
            </div>

            <div className="lg:col-span-1 flex gap-2">
              <button
                type="button"
                className="w-full rounded bg-slate-900 px-4 py-2 text-white hover:bg-slate-800"
                onClick={async () => {
                  const normalized = targetId.trim()
                  if (!normalized) {
                    setSearchParams({})
                    await loadGraph(contextType, normalized, depth)
                    return
                  }

                  const nextParams: Record<string, string> = { [contextType]: normalized }
                  setSearchParams(nextParams)
                  await loadGraph(contextType, normalized, depth)
                }}
              >
                Load Graph
              </button>
            </div>
          </div>
          <p className="mt-3 text-xs text-slate-500">
            Tip: Open a paper detail and use “Explore Knowledge Graph” for one-click graph loading.
          </p>

          <div className="mt-3 flex flex-wrap items-center gap-2">
            <button
              type="button"
              className="rounded border border-slate-300 px-3 py-1.5 text-sm hover:bg-slate-100"
              onClick={copyShareLink}
            >
              Copy Share Link
            </button>
            {breadcrumbs.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {breadcrumbs.map((item, idx) => (
                  <button
                    key={`${item.type}-${item.id}-${idx}`}
                    type="button"
                    className="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-700 hover:bg-slate-200"
                    onClick={async () => {
                      setContextType(item.type)
                      setTargetId(item.id)
                      setSearchParams(buildSearchParams(item.type, item.id))
                      await loadGraph(item.type, item.id, depth)
                    }}
                  >
                    {item.type}:{item.id}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="grid grid-cols-1 xl:grid-cols-12 gap-6">
          <div className="xl:col-span-9">
            <KnowledgeGraph
              graph={graph}
              loading={loading}
              error={error}
              selectedNodeId={selectedNode?.id}
              onNodeSelect={setSelectedNode}
              height={680}
            />
          </div>

          <aside className="xl:col-span-3">
            <GraphNodeDetailPanel
              node={selectedNode}
              onPivotPaper={async (paperId) => {
                setContextType('paper')
                setTargetId(paperId)
                setSearchParams(buildSearchParams('paper', paperId))
                await loadGraph('paper', paperId, depth)
              }}
              onPivotAuthor={async (authorId) => {
                setContextType('author')
                setTargetId(authorId)
                setSearchParams(buildSearchParams('author', authorId))
                await loadGraph('author', authorId, depth)
              }}
              onPivotConcept={async (conceptId) => {
                setContextType('concept')
                setTargetId(conceptId)
                setSearchParams(buildSearchParams('concept', conceptId))
                await loadGraph('concept', conceptId, depth)
              }}
            />
          </aside>
        </div>
      </div>
    </AppShell>
  )
}

export default KnowledgeGraphPage
