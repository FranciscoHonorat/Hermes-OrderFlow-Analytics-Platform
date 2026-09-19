import axios, { type AxiosInstance } from 'axios'
import router from '@/router'
import { useAuthStore } from '@/stores/auth'

function createClient(basePath: string): AxiosInstance {
  const client = axios.create({ baseURL: basePath })

  client.interceptors.request.use((config) => {
    const auth = useAuthStore()
    if (auth.token) {
      config.headers.Authorization = `Bearer ${auth.token}`
    }
    return config
  })

  client.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.response?.status === 401) {
        useAuthStore().logout()
        router.push('/login')
      }
      return Promise.reject(error)
    },
  )

  return client
}

export const orderApi = createClient('/api/order')
export const inventoryApi = createClient('/api/inventory')
export const userApi = createClient('/api/user')
