export type OrderStatus =
  | 'PENDING'
  | 'CONFIRMED'
  | 'SHIPPED'
  | 'DELIVERED'
  | 'CANCELLED'
  | 'PAID'
  | 'PLACED'

export interface OrderItemDTO {
  product_id: string
  quantity: number
  price: number
}

export interface OrderDTO {
  id: string
  customer_id: string
  status: OrderStatus
  total: number
  /** O backend hoje sempre retorna "USD", independente da moeda real do pedido. */
  currency: string
  /** O backend hoje sempre retorna [] na leitura — a query não faz join com itens. */
  items: OrderItemDTO[]
  created_at: string
  updated_at: string
}

export interface PlaceOrderRequest {
  customer_id: string
  items: OrderItemDTO[]
}

export interface PlaceOrderResponse {
  order_id: string
}

export interface OrderActionRequest {
  order_id: string
}

export interface ListOrdersParams {
  limit?: number
  offset?: number
  customer_id?: string
}
