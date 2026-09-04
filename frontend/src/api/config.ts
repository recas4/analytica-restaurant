export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

export const getAuthToken = () => localStorage.getItem('auth_token')

export const setAuthToken = (token: string) => localStorage.setItem('auth_token', token)

export const clearAuthToken = () => localStorage.removeItem('auth_token')

// Wrapper for authenticated fetch requests with 401 handling
export async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const token = getAuthToken()
  const headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`,
  }

  const response = await fetch(url, { ...options, headers })

  if (response.status === 401) {
    clearAuthToken()
    window.location.href = '/login'
    throw new Error('Session expired')
  }

  return response
}
