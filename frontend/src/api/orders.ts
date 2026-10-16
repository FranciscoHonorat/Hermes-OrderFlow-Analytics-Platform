import { orderApi } from '@/api/http'
import type {
  ListOrdersParams,
  OrderDTO,
  PlaceOrderRequest,
  PlaceOrderResponse,
} from '@/types/order'

export const placeOrder = (body: PlaceOrderRequest) =>
  orderApi.post<PlaceOrderResponse>('/orders', body)

export const listOrders = (params: ListOrdersParams) =>
  orderApi.get<OrderDTO[]>('/orders', { params })

export const getOrder = (id: string) => orderApi.get<OrderDTO>(`/orders/${id}`)

export const confirmOrder = (orderId: string) =>
  orderApi.post('/orders/confirm', { order_id: orderId })

export const cancelOrder = (orderId: string) =>
  orderApi.post('/orders/cancel', { order_id: orderId })

export const shipOrder = (orderId: string) =>
  orderApi.post('/orders/ship', { order_id: orderId })
