import React, { createContext, useContext, useState, useEffect } from 'react'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [user, setUser] = useState(null)
  
  useEffect(() => {
    // TODO: Initialize Keycloak here
    // For now, simulate authenticated state
    setTimeout(() => {
      setIsAuthenticated(true)
      setUser({ name: 'Demo User', email: 'demo@company.com' })
      setIsLoading(false)
    }, 500)
  }, [])
  
  const login = () => {
    // TODO: Trigger Keycloak login
    console.log('Login triggered')
  }
  
  const logout = () => {
    // TODO: Trigger Keycloak logout
    setIsAuthenticated(false)
    setUser(null)
  }
  
  return (
    <AuthContext.Provider value={{
      isAuthenticated,
      isLoading,
      user,
      login,
      logout
    }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
