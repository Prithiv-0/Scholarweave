import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import SearchPage from '@/pages/Search'
import PaperDetail from '@/pages/PaperDetail'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<SearchPage />} />
        <Route path="/papers/:id" element={<PaperDetail />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
