import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:9001/api',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

let isRefreshing = false
let refreshSubscribers: ((err?: unknown) => void)[] = []

const subscribeTokenRefresh = (cb: (err?: unknown) => void) => {
  refreshSubscribers.push(cb)
}

const onRefreshed = (err?: unknown) => {
  refreshSubscribers.forEach((cb) => cb(err))
  refreshSubscribers = []
}

client.interceptors.response.use(
  (response) => response,
  async (error) => {
    const { config, response } = error
    const originalRequest = config

    // Skip redirect/rotation loops on actual auth endpoints
    if (response && response.status === 401 && !originalRequest._retry) {
      if (originalRequest.url?.includes('/auth/refresh') || originalRequest.url?.includes('/auth/login')) {
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          subscribeTokenRefresh((err?: unknown) => {
            if (err) {
              reject(err)
            } else {
              resolve(client(originalRequest))
            }
          })
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        await client.post('/auth/refresh')
        isRefreshing = false
        onRefreshed()
        return client(originalRequest)
      } catch (refreshError) {
        isRefreshing = false
        onRefreshed(refreshError)
        const authStore = useAuthStore()
        authStore.clearSession()
        if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
          window.location.href = '/login'
        }
        return Promise.reject(refreshError)
      }
    }

    if (response && response.status === 402) {
      window.location.href = '/billing'
      return Promise.reject(error)
    }

    return Promise.reject(error)
  }
)

export default client
