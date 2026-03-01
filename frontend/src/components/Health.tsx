import React, { useEffect, useState } from 'react'
import apiService, { HealthResponse } from '@/api/client'

export const Health: React.FC = () => {
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        setLoading(true)
        const data = await apiService.health()
        setHealth(data)
        setError(null)
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : 'Failed to fetch health status'
        setError(message)
        setHealth(null)
      } finally {
        setLoading(false)
      }
    }

    fetchHealth()
    // Refresh every 30 seconds
    const interval = setInterval(fetchHealth, 30000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return (
      <div className="p-4 bg-gray-100 rounded-lg animate-pulse">
        <p className="text-gray-600">Loading health status...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
        <p className="text-red-700 font-semibold">Error</p>
        <p className="text-red-600 text-sm">{error}</p>
      </div>
    )
  }

  if (!health) {
    return (
      <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
        <p className="text-yellow-700">No health data available</p>
      </div>
    )
  }

  const isHealthy = health.status === 'ok'
  const statusColors = isHealthy
    ? { bg: 'bg-green-50 border-green-200', dot: 'bg-green-600', heading: 'text-green-900', text: 'text-green-800', sub: 'text-green-700' }
    : { bg: 'bg-yellow-50 border-yellow-200', dot: 'bg-yellow-600', heading: 'text-yellow-900', text: 'text-yellow-800', sub: 'text-yellow-700' }

  return (
    <div className={`p-4 ${statusColors.bg} border rounded-lg`}>
      <div className="flex items-center gap-2 mb-3">
        <div className={`w-3 h-3 ${statusColors.dot} rounded-full`}></div>
        <p className={`font-semibold ${statusColors.heading}`}>System Status: {health.status}</p>
      </div>
      <div className={`text-sm ${statusColors.text} space-y-1`}>
        <p>Version: {health.version}</p>
        <p>API: {health.services.api}</p>
        <p>OpenAlex: {health.services.openalex}</p>
        <p className={`text-xs ${statusColors.sub} mt-2`}>
          Last checked: {new Date(health.timestamp).toLocaleTimeString()}
        </p>
      </div>
    </div>
  )
}

export default Health
