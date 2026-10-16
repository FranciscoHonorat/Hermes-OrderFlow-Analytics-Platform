import { userApi } from '@/api/http'
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  UpdateRoleRequest,
  UserDTO,
} from '@/types/user'

export const register = (body: RegisterRequest) => userApi.post('/auth/register', body)

export const login = (body: LoginRequest) => userApi.post<LoginResponse>('/auth/login', body)

export const getMe = () => userApi.get<UserDTO>('/users/me')

export const listUsers = () => userApi.get<UserDTO[]>('/users')

export const getUser = (id: string) => userApi.get<UserDTO>(`/users/${id}`)

export const updateUserRole = (id: string, body: UpdateRoleRequest) =>
  userApi.patch(`/users/${id}/role`, body)

export const activateUser = (id: string) => userApi.post(`/users/${id}/activate`)

export const deactivateUser = (id: string) => userApi.post(`/users/${id}/deactivate`)
