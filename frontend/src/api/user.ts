import client from './client'

export interface UserProfile {
  id: number
  username: string
  is_admin: boolean
  enabled: boolean
  sub_token: string
  traffic_used: number
  expire_at: string | null
  traffic_reset_at: string | null
  plan: { id: number; name: string; traffic_limit: number; duration_days: number } | null
  created_at: string
}

export interface SubscriptionDetail {
  plan_name: string
  traffic_used: number
  traffic_limit: number
  expire_at: string | null
  sub_url: string
  sub_url_clash: string
  sub_url_singbox: string
  sub_url_base64: string
}

export const userApi = {
  profile: () => client.get<UserProfile>('/user/profile'),
  subscription: () => client.get<SubscriptionDetail>('/user/subscription'),
  changePassword: (data: { old_password: string; new_password: string }) =>
    client.put('/user/password', data),
}
