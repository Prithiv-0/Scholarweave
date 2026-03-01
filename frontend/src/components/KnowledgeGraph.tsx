import React, { useEffect, useMemo, useRef, useState } from 'react'
import ForceGraph2D from 'react-force-graph-2d'
import { GraphNode, GraphNodeType, KnowledgeGraph as KnowledgeGraphData } from '@/api/client'

type VisibleNodeTypes = Record<GraphNodeType, boolean>

interface KnowledgeGraphProps {
  graph: KnowledgeGraphData | null
  loading?: boolean
  error?: string | null
  height?: number
  selectedNodeId?: string
  onNodeSelect?: (node: GraphNode) => void
}

const nodeColors: Record<GraphNodeType, string> = {
  paper: '#2563eb',
  author: '#059669',
  concept: '#d97706',
  institution: '#7c3aed',
}

const edgeColors: Record<string, string> = {
  authored: '#10b981',
  has_concept: '#f59e0b',
  cites: '#3b82f6',
  related: '#94a3b8',
  affiliated: '#8b5cf6',
}

const initialVisibility: VisibleNodeTypes = {
  paper: true,
  author: true,
  concept: true,
  institution: true,
}

const nodeSize = (node: GraphNode): number => {
  if (node.type === 'paper') {
    const citations = Number(node.properties?.cited_by_count || 0)
    return Math.max(4, Math.min(12, 4 + Math.log10(citations + 1) * 3))
  }
  if (node.type === 'concept') {
    const score = Number(node.properties?.score || 0)
    return Math.max(4, Math.min(10, 4 + score * 6))
  }
  if (node.type === 'author') return 6
  return 5
}

const downloadFile = (blob: Blob, filename: string) => {
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.setAttribute('download', filename)
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  URL.revokeObjectURL(url)
}

const KnowledgeGraph: React.FC<KnowledgeGraphProps> = ({
  graph,
  loading = false,
  error = null,
  height = 640,
  selectedNodeId,
  onNodeSelect,
}) => {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const fgRef = useRef<any>(null)
  const [width, setWidth] = useState(1000)
  const [visibleTypes, setVisibleTypes] = useState<VisibleNodeTypes>(initialVisibility)

  useEffect(() => {
    const updateWidth = () => {
      if (!containerRef.current) return
      setWidth(Math.max(320, containerRef.current.clientWidth))
    }

    updateWidth()
    window.addEventListener('resize', updateWidth)
    return () => window.removeEventListener('resize', updateWidth)
  }, [])

  const graphData = useMemo(() => {
    if (!graph) {
      return { nodes: [], links: [] }
    }

    const visibleIds = new Set(
      graph.nodes
        .filter((node) => visibleTypes[node.type])
        .map((node) => node.id)
    )

    const nodes = graph.nodes.filter((node) => visibleIds.has(node.id))
    const links = graph.edges
      .filter((edge) => visibleIds.has(edge.source) && visibleIds.has(edge.target))
      .map((edge) => ({ ...edge, source: edge.source, target: edge.target }))

    return { nodes, links }
  }, [graph, visibleTypes])

  const handleExportJSON = () => {
    if (!graph) return
    const payload = JSON.stringify(graph, null, 2)
    downloadFile(new Blob([payload], { type: 'application/json;charset=utf-8;' }), `scholarweave-graph-${Date.now()}.json`)
  }

  const handleExportPNG = () => {
    const canvas = fgRef.current?.canvas?.()
    if (!canvas) return
    canvas.toBlob((blob: Blob | null) => {
      if (!blob) return
      downloadFile(blob, `scholarweave-graph-${Date.now()}.png`)
    }, 'image/png')
  }

  useEffect(() => {
    if (!graphData.nodes.length || !fgRef.current) return
    const timer = window.setTimeout(() => {
      fgRef.current?.zoomToFit?.(600, 50)
    }, 300)
    return () => window.clearTimeout(timer)
  }, [graphData.nodes.length])

  if (loading) {
    return <div className="h-80 rounded-lg border border-slate-200 bg-white p-6">Loading graph...</div>
  }

  if (error) {
    return <div className="h-80 rounded-lg border border-red-200 bg-red-50 p-6 text-red-700">{error}</div>
  }

  if (!graph || graphData.nodes.length === 0) {
    return <div className="h-80 rounded-lg border border-slate-200 bg-white p-6 text-slate-600">No graph data available.</div>
  }

  return (
    <div className="space-y-4" ref={containerRef}>
      <div className="flex flex-wrap items-center gap-3 rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm">
        <span className="font-semibold text-slate-700">Node filters:</span>
        {(Object.keys(initialVisibility) as GraphNodeType[]).map((type) => (
          <label key={type} className="inline-flex items-center gap-2 text-slate-700">
            <input
              type="checkbox"
              checked={visibleTypes[type]}
              onChange={(e) => setVisibleTypes((prev) => ({ ...prev, [type]: e.target.checked }))}
            />
            <span>{type}</span>
          </label>
        ))}
        <button
          type="button"
          className="ml-auto rounded border border-slate-300 px-3 py-1.5 hover:bg-slate-50"
          onClick={() => fgRef.current?.zoomToFit?.(500, 40)}
        >
          Fit Graph
        </button>
        <button
          type="button"
          className="rounded border border-slate-300 px-3 py-1.5 hover:bg-slate-50"
          onClick={handleExportJSON}
        >
          Export JSON
        </button>
        <button
          type="button"
          className="rounded border border-slate-300 px-3 py-1.5 hover:bg-slate-50"
          onClick={handleExportPNG}
        >
          Export PNG
        </button>
      </div>

      <div className="rounded-lg border border-slate-200 bg-white overflow-hidden" aria-label={`Knowledge graph with ${graphData.nodes.length} nodes`}>
        <ForceGraph2D
          ref={fgRef}
          width={width}
          height={height}
          graphData={graphData as any}
          linkWidth={(link: any) => Math.max(1, Number(link.weight || 1))}
          linkColor={(link: any) => edgeColors[String(link.relationship || '')] || '#94a3b8'}
          linkLineDash={(link: any) => {
            if (link.relationship === 'related') return [5, 4]
            if (link.relationship === 'has_concept') return [2, 3]
            return []
          }}
          nodeLabel={(node: any) => {
            const typedNode = node as GraphNode
            const extra =
              typedNode.type === 'paper'
                ? `Citations: ${typedNode.properties?.cited_by_count ?? 0}`
                : typedNode.type === 'concept'
                  ? `Score: ${typedNode.properties?.score ?? 0}`
                  : ''
            return `${typedNode.label} (${typedNode.type}) ${extra}`.trim()
          }}
          nodeRelSize={1}
          nodeCanvasObject={(nodeObj: any, ctx, globalScale) => {
            const node = nodeObj as GraphNode
            const label = node.label || node.id
            const size = nodeSize(node)
            const color = nodeColors[node.type] || '#334155'
            const isSelected = selectedNodeId === node.id

            ctx.beginPath()
            ctx.arc(nodeObj.x, nodeObj.y, isSelected ? size + 2 : size, 0, 2 * Math.PI, false)
            ctx.fillStyle = color
            ctx.fill()

            if (isSelected) {
              ctx.lineWidth = 2
              ctx.strokeStyle = '#0f172a'
              ctx.stroke()
            }

            const fontSize = Math.max(10, 12 / globalScale)
            ctx.font = `${fontSize}px sans-serif`
            ctx.fillStyle = '#0f172a'
            ctx.fillText(label, nodeObj.x + size + 2, nodeObj.y + size + 2)
          }}
          onNodeClick={(node: any) => onNodeSelect?.(node as GraphNode)}
          cooldownTicks={120}
        />
      </div>

      <div className="rounded-lg border border-slate-200 bg-white px-4 py-3 text-xs text-slate-600">
        <div className="flex flex-wrap gap-4">
          {(Object.keys(nodeColors) as GraphNodeType[]).map((type) => (
            <span key={type} className="inline-flex items-center gap-1.5">
              <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: nodeColors[type] }} />
              {type}
            </span>
          ))}
        </div>
      </div>
    </div>
  )
}

export default KnowledgeGraph
