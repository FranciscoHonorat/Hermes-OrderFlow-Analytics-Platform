export type UserRole = 'admin' | 'user'

export interface UserDTO {
  id: string
  email: string
  role: UserRole
  active: boolean
  created_at: string
  updated_at: string
}

export interface RegisterRequest {
  email: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  expires_at: string
}

export interface UpdateRoleRequest {
  role: UserRole
}
