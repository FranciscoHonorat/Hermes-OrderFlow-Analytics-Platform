export interface StockDTO {
  sku: string
  available: number
  reserved: number
  minimum_quantity: number
  updated_at: string
}

export interface StockMutationRequest {
  sku: string
  quantity: number
}

export interface StockMutationResponse {
  message: string
}
