import axios from 'axios'
import type { ApiErrorPayload } from '@/types/api'

export function extractErrorMessage(err: unknown): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data
    if (typeof data === 'string' && data.trim()) return data
    if (data && typeof data === 'object' && 'error' in data) {
      return String((data as ApiErrorPayload).error)
    }
    if (err.message) return err.message
  }
  return 'Unexpected error'
}
