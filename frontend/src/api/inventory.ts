import { inventoryApi } from '@/api/http'
import type { StockDTO, StockMutationRequest, StockMutationResponse } from '@/types/inventory'

export const reserveStock = (body: StockMutationRequest) =>
  inventoryApi.post<StockMutationResponse>('/stock/reserve', body)

export const releaseStock = (body: StockMutationRequest) =>
  inventoryApi.post<StockMutationResponse>('/stock/release', body)

export const replenishStock = (body: StockMutationRequest) =>
  inventoryApi.post<StockMutationResponse>('/stock/replenish', body)

export const listLowStock = () => inventoryApi.get<StockDTO[]>('/stock/low')

export const getStock = (sku: string) => inventoryApi.get<StockDTO>(`/stock/${sku}`)
