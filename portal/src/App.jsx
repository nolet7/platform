import React from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from './auth/AuthProvider'

// Placeholder pages
const LoginPage = () => <div style={{ padding: '2rem' }}><h1>Login Page</h1><p>Keycloak integration will be configured here</p></div>
const CatalogPage = () => <div style={{ padding: '2rem' }}><h1>Catalog</h1><p>Service and model catalog will appear here</p></div>
const ModelsPage = () => <div style={{ padding: '2rem' }}><h1>ML Models</h1><p>Model registry will appear here</p></div>

const queryClient = new QueryClient()

function ProtectedRoute({ children }) {
  const { isAuthenticated, isLoading } = useAuth()
  
  if (isLoading) return <div>Loading...</div>
  if (!isAuthenticated) return <Navigate to="/login" />
  
  return children
}

function App() {
  return (
    <AuthProvider>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <div style={{ minHeight: '100vh', backgroundColor: '#f5f5f5' }}>
            <nav style={{ 
              background: '#1a1a1a', 
              color: 'white', 
              padding: '1rem 2rem',
              display: 'flex',
              gap: '2rem'
            }}>
              <h1 style={{ margin: 0, fontSize: '1.5rem' }}>Platform Portal</h1>
              <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                <a href="/catalog" style={{ color: 'white', textDecoration: 'none' }}>Catalog</a>
                <a href="/models" style={{ color: 'white', textDecoration: 'none' }}>Models</a>
              </div>
            </nav>
            
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/catalog" element={<ProtectedRoute><CatalogPage /></ProtectedRoute>} />
              <Route path="/models" element={<ProtectedRoute><ModelsPage /></ProtectedRoute>} />
              <Route path="/" element={<Navigate to="/catalog" />} />
            </Routes>
          </div>
        </BrowserRouter>
      </QueryClientProvider>
    </AuthProvider>
  )
}

export default App
