import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { setAuthToken, clearAuthToken, getAuthToken } from '@/api/config'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<{ id: string; email: string; role: string } | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(email: string, password: string) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.login({ email, password })
      user.value = response.user
      setAuthToken(response.token)
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Login failed'
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(email: string, password: string, role?: string) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.register({ email, password, role })
      user.value = response.user
      setAuthToken(response.token)
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Registration failed'
      return false
    } finally {
      loading.value = false
    }
  }

  async function checkAuth() {
    const token = getAuthToken()
    if (!token) return false

    try {
      const userData = await authApi.me()
      user.value = userData
      return true
    } catch (e) {
      clearAuthToken()
      user.value = null
      return false
    }
  }

  function logout() {
    user.value = null
    clearAuthToken()
    window.location.href = '/login'
  }

  return {
    user,
    loading,
    error,
    isAuthenticated,
    isAdmin,
    login,
    register,
    checkAuth,
    logout,
  }
})
