import client from './client'

export interface LoginReq { username: string; password: string }
export interface RegisterReq { username: string; password: string }
export interface LoginRes { token: string; user: { id: number; username: string; is_admin: boolean } }

export const authApi = {
  login: (data: LoginReq) => client.post<LoginRes>('/auth/login', data),
  register: (data: RegisterReq) => client.post<{ id: number; username: string }>('/auth/register', data),
}
