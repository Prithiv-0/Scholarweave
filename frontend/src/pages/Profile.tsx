import React from 'react'
import { Link, useNavigate } from 'react-router-dom'
import AppShell from '@/components/AppShell'
import { useAuth } from '@/context/AuthContext'

const ProfilePage: React.FC = () => {
  const navigate = useNavigate()
  const { user, logout } = useAuth()

  const handleLogout = () => {
    logout()
    navigate('/auth')
  }

  if (!user) {
    return (
      <AppShell title="Profile">
        <div className="flex items-center justify-center h-64">
          <p className="text-gray-500">Loading profile data...</p>
        </div>
      </AppShell>
    )
  }

  return (
    <AppShell title="User Profile" subtitle="Manage your account and activity">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 bg-white border border-slate-200 rounded-lg p-6 space-y-4">
          <h2 className="text-xl font-semibold text-slate-900">Account Information</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p className="text-xs uppercase tracking-wide text-slate-500 font-semibold">Full Name</p>
              <p className="text-slate-900 mt-1">{user.name}</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-slate-500 font-semibold">Email</p>
              <p className="text-slate-900 mt-1">{user.email}</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-slate-500 font-semibold">Member Since</p>
              <p className="text-slate-900 mt-1">{new Date(user.created_at).toLocaleString()}</p>
            </div>
          </div>
        </div>

        <div className="bg-white border border-slate-200 rounded-lg p-6 space-y-3">
          <h3 className="text-lg font-semibold text-slate-900">Quick Actions</h3>
          <Link to="/" className="block w-full px-4 py-2 rounded border border-slate-300 hover:bg-slate-50 text-center">Back to search</Link>
          <Link to="/library" className="block w-full px-4 py-2 rounded bg-slate-900 text-white hover:bg-slate-800 text-center">Open library</Link>
          <button onClick={handleLogout} className="w-full px-4 py-2 rounded bg-red-600 text-white hover:bg-red-700">Sign out</button>
        </div>
      </div>
    </AppShell>
  )
}

export default ProfilePage
