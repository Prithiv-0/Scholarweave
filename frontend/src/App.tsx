import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import SearchPage from '@/pages/Search'
import PaperDetail from '@/pages/PaperDetail'
import AuthPage from '@/pages/Auth'
import ProfilePage from '@/pages/Profile'
import LibraryPage from '@/pages/Library'
import KnowledgeGraphPage from '@/pages/KnowledgeGraph'
import { AuthProvider } from '@/context/AuthContext'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { ErrorBoundary } from '@/components/ErrorBoundary'

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<SearchPage />} />
            <Route path="/auth" element={<AuthPage />} />
            <Route path="/graph" element={<KnowledgeGraphPage />} />
            <Route path="/papers/:id" element={<PaperDetail />} />

            {/* Protected Routes */}
            <Route path="/profile" element={
              <ProtectedRoute>
                <ProfilePage />
              </ProtectedRoute>
            } />
            <Route path="/library" element={
              <ProtectedRoute>
                <LibraryPage />
              </ProtectedRoute>
            } />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default App
