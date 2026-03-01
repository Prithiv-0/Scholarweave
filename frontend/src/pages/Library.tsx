import React, { useEffect, useState } from 'react'
import axios from 'axios'
import { Link, useNavigate } from 'react-router-dom'
import AppShell from '@/components/AppShell'
import apiService, { ReadingList, SavedPaper } from '@/api/client'

const getErrorMessage = (err: unknown, fallback: string): string => {
  if (axios.isAxiosError<{ error?: string }>(err)) {
    return err.response?.data?.error || err.message || fallback
  }
  return err instanceof Error ? err.message : fallback
}

const LibraryPage: React.FC = () => {
  const navigate = useNavigate()
  const [favorites, setFavorites] = useState<SavedPaper[]>([])
  const [lists, setLists] = useState<ReadingList[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [newListName, setNewListName] = useState('')
  const [newListDescription, setNewListDescription] = useState('')
  const [targetListByPaper, setTargetListByPaper] = useState<Record<string, string>>({})

  const loadLibrary = async () => {
    try {
      setLoading(true)
      setError(null)
      const [loadedFavorites, loadedLists] = await Promise.all([
        apiService.listFavorites(),
        apiService.listReadingLists(),
      ])
      setFavorites(loadedFavorites)
      setLists(loadedLists)
    } catch (err: unknown) {
      setError(getErrorMessage(err, 'Failed to load library'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadLibrary()
  }, [])

  const handleCreateList = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!newListName.trim()) return
    try {
      await apiService.createReadingList(newListName.trim(), newListDescription.trim())
      setNewListName('')
      setNewListDescription('')
      await loadLibrary()
    } catch (err: unknown) {
      setError(getErrorMessage(err, 'Failed to create reading list'))
    }
  }

  const handleRemoveFavorite = async (paperId: string) => {
    try {
      await apiService.removeFavorite(paperId)
      await loadLibrary()
    } catch (err: unknown) {
      setError(getErrorMessage(err, 'Failed to remove favorite'))
    }
  }

  const handleAddFavoriteToList = async (paper: SavedPaper) => {
    const listId = targetListByPaper[paper.paper_id]
    if (!listId) return
    try {
      await apiService.addPaperToReadingList(listId, paper)
      await loadLibrary()
    } catch (err: unknown) {
      setError(getErrorMessage(err, 'Failed to add paper to reading list'))
    }
  }

  const handleRemoveFromList = async (listId: string, paperId: string) => {
    try {
      await apiService.removePaperFromReadingList(listId, paperId)
      await loadLibrary()
    } catch (err: unknown) {
      setError(getErrorMessage(err, 'Failed to remove paper from list'))
    }
  }

  if (loading) return <AppShell title="Knowledge Library"><div>Loading your library...</div></AppShell>

  return (
    <AppShell title="Knowledge Library" subtitle="Curate favorites and structured reading lists">
      <div className="flex items-center justify-end mb-4">
        <Link to="/" className="px-4 py-2 rounded border border-slate-300 hover:bg-slate-50">Back to search</Link>
      </div>

      {error && (
        <div className="mb-6 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg">
          {error}
        </div>
      )}

      <div className="bg-white border border-slate-200 rounded-lg p-5 mb-8">
        <h2 className="text-xl font-semibold mb-3">Create Reading List</h2>
        <form className="grid grid-cols-1 md:grid-cols-3 gap-3" onSubmit={handleCreateList}>
          <input
            type="text"
            placeholder="List name"
            value={newListName}
            onChange={(event) => setNewListName(event.target.value)}
            className="px-3 py-2 rounded border border-slate-300"
            required
          />
          <input
            type="text"
            placeholder="Description"
            value={newListDescription}
            onChange={(event) => setNewListDescription(event.target.value)}
            className="px-3 py-2 rounded border border-slate-300"
          />
          <button type="submit" className="px-4 py-2 bg-slate-900 text-white rounded hover:bg-slate-800">Create</button>
        </form>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <section className="bg-white border border-slate-200 rounded-lg p-5">
          <h2 className="text-xl font-semibold mb-4">Favorites ({favorites.length})</h2>
          {favorites.length === 0 && <p className="text-slate-600 text-sm">No favorites yet.</p>}
          <div className="space-y-4">
            {favorites.map((paper) => (
              <div key={paper.paper_id} className="border border-slate-100 rounded p-3">
                <p className="font-medium text-slate-900">{paper.title}</p>
                <p className="text-xs text-slate-500 mt-1">{paper.paper_id}</p>
                <div className="mt-3 flex flex-wrap items-center gap-2">
                  <select
                    className="px-2 py-1 rounded border border-slate-300 text-sm"
                    value={targetListByPaper[paper.paper_id] || ''}
                    onChange={(event) => {
                      setTargetListByPaper((prev) => ({ ...prev, [paper.paper_id]: event.target.value }))
                    }}
                  >
                    <option value="">Select reading list</option>
                    {lists.map((list) => (
                      <option key={list.id} value={list.id}>{list.name}</option>
                    ))}
                  </select>
                  <button
                    type="button"
                    onClick={() => handleAddFavoriteToList(paper)}
                    className="px-3 py-1 text-sm rounded bg-slate-900 text-white hover:bg-slate-800"
                    disabled={!targetListByPaper[paper.paper_id]}
                  >
                    Add to list
                  </button>
                  <button
                    type="button"
                    onClick={() => handleRemoveFavorite(paper.paper_id)}
                    className="px-3 py-1 text-sm rounded bg-red-600 text-white hover:bg-red-700"
                  >
                    Remove
                  </button>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="bg-white border border-slate-200 rounded-lg p-5">
          <h2 className="text-xl font-semibold mb-4">Reading Lists ({lists.length})</h2>
          {lists.length === 0 && <p className="text-slate-600 text-sm">No reading lists yet.</p>}
          <div className="space-y-4">
            {lists.map((list) => (
              <div key={list.id} className="border border-slate-100 rounded p-3">
                <p className="font-medium text-slate-900">{list.name}</p>
                {list.description && <p className="text-sm text-slate-600 mt-1">{list.description}</p>}
                <div className="mt-2 space-y-2">
                  {list.papers.map((paper) => (
                    <div key={paper.paper_id} className="flex items-center justify-between gap-2 text-sm">
                      <span className="text-slate-700 truncate">{paper.title}</span>
                      <button
                        type="button"
                        onClick={() => handleRemoveFromList(list.id, paper.paper_id)}
                        className="px-2 py-1 rounded bg-red-50 text-red-700 hover:bg-red-100"
                      >
                        Remove
                      </button>
                    </div>
                  ))}
                  {list.papers.length === 0 && <p className="text-xs text-slate-500">No papers in this list.</p>}
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>
    </AppShell>
  )
}

export default LibraryPage
