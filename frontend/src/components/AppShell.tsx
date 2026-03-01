import React from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'

interface AppShellProps {
  title: string
  subtitle?: string
  children: React.ReactNode
  rightSlot?: React.ReactNode
}

const AppShell: React.FC<AppShellProps> = ({ title, subtitle, children, rightSlot }) => {
  const location = useLocation()
  const navigate = useNavigate()
  const { token, logout } = useAuth()
  const isAuthenticated = Boolean(token)

  const handleSignOut = () => {
    logout()
    navigate('/auth')
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="bg-white border-b border-slate-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between gap-4">
          <div>
            <p className="text-xs font-semibold tracking-wide text-slate-500 uppercase">ScholarWeave</p>
            <h1 className="text-2xl font-bold">{title}</h1>
            {subtitle && <p className="text-sm text-slate-600">{subtitle}</p>}
          </div>

          <div className="flex items-center gap-2 text-sm">
            <nav aria-label="Main navigation" className="flex items-center gap-2">
            <Link to="/" className={`px-3 py-2 rounded-md ${location.pathname === '/' ? 'bg-slate-900 text-white' : 'bg-slate-100 hover:bg-slate-200'}`}>Search</Link>
            <Link to="/graph" className={`px-3 py-2 rounded-md ${location.pathname.startsWith('/graph') ? 'bg-slate-900 text-white' : 'bg-slate-100 hover:bg-slate-200'}`}>Knowledge Graph</Link>
            {isAuthenticated ? (
              <>
                <Link to="/library" className={`px-3 py-2 rounded-md ${location.pathname.startsWith('/library') ? 'bg-slate-900 text-white' : 'bg-slate-100 hover:bg-slate-200'}`}>Library</Link>
                <Link to="/profile" className={`px-3 py-2 rounded-md ${location.pathname.startsWith('/profile') ? 'bg-slate-900 text-white' : 'bg-slate-100 hover:bg-slate-200'}`}>Profile</Link>
                <button onClick={handleSignOut} className="px-3 py-2 rounded-md bg-red-600 text-white hover:bg-red-700">Sign out</button>
              </>
            ) : (
              <Link to="/auth" className="px-3 py-2 rounded-md bg-slate-900 text-white hover:bg-slate-800">Sign in</Link>
            )}
            </nav>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        {rightSlot && <div className="mb-6">{rightSlot}</div>}
        {children}
      </main>
    </div>
  )
}

export default AppShell
