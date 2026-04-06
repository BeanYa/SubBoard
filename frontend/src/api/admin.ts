import client from './client'

export interface User {
  id: number
  username: string
  plan_id: number | null
  sub_token: string
  traffic_used: number
  expire_at: string | null
  traffic_reset_at: string | null
  is_admin: boolean
  enabled: boolean
  created_at: string
}

export interface Plan {
  id: number
  name: string
  description: string
  traffic_limit: number
  duration_days: number
  price: string
  enabled: boolean
  created_at: string
}

export interface SubscriptionSource {
  id: number
  name: string
  type: 'substore' | 'url' | 'raw'
  url: string
  headers: Record<string, string>
  refresh_interval: number
  node_count: number
  last_fetch_at: string | null
  fetch_error: string
  enabled: boolean
}

export interface ServiceGroup {
  id: number
  name: string
  description: string
  sort_order: number
  enabled: boolean
  subscription_ids: number[]
  agent_ids: number[]
}

export interface Agent {
  id: number
  name: string
  token: string
  server_addr: string
  port: number
  protocol: string
  protocol_config: Record<string, unknown>
  traffic_used: number
  traffic_total: number
  cpu_usage: number
  mem_usage: number
  status: string
  last_report_at: string | null
  enabled: boolean
}

export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

const admin = '/admin'

export const adminApi = {
  // Users
  users: (page = 1, pageSize = 20, search = '') =>
    client.get<Paginated<User>>(`${admin}/users`, { params: { page, page_size: pageSize, search } }),
  createUser: (data: { username: string; password: string }) =>
    client.post<User>(`${admin}/users`, data),
  updateUser: (id: number, data: Partial<User>) =>
    client.put<User>(`${admin}/users/${id}`, data),
  deleteUser: (id: number) => client.delete(`${admin}/users/${id}`),
  assignPlan: (id: number, plan_id: number | null) =>
    client.post(`${admin}/users/${id}/plan`, { plan_id }),
  resetTraffic: (id: number) => client.post(`${admin}/users/${id}/reset`),
  toggleUser: (id: number) => client.post<User>(`${admin}/users/${id}/toggle`),
  changePassword: (id: number, password: string) =>
    client.put<User>(`${admin}/users/${id}`, { password }),

  // Plans
  plans: () => client.get<Plan[]>(`${admin}/plans`),
  createPlan: (data: Omit<Plan, 'id' | 'created_at'>) =>
    client.post<Plan>(`${admin}/plans`, data),
  updatePlan: (id: number, data: Partial<Plan>) =>
    client.put<Plan>(`${admin}/plans/${id}`, data),
  deletePlan: (id: number) => client.delete(`${admin}/plans/${id}`),

  // Subscriptions
  subscriptions: () => client.get<SubscriptionSource[]>(`${admin}/subscriptions`),
  createSubscription: (data: Partial<SubscriptionSource>) =>
    client.post<SubscriptionSource>(`${admin}/subscriptions`, data),
  updateSubscription: (id: number, data: Partial<SubscriptionSource>) =>
    client.put<SubscriptionSource>(`${admin}/subscriptions/${id}`, data),
  deleteSubscription: (id: number) => client.delete(`${admin}/subscriptions/${id}`),
  refreshSubscription: (id: number) =>
    client.post(`${admin}/subscriptions/${id}/refresh`),

  // Groups
  groups: () => client.get<ServiceGroup[]>(`${admin}/groups`),
  createGroup: (data: Partial<ServiceGroup>) =>
    client.post<ServiceGroup>(`${admin}/groups`, data),
  updateGroup: (id: number, data: Partial<ServiceGroup>) =>
    client.put<ServiceGroup>(`${admin}/groups/${id}`, data),
  deleteGroup: (id: number) => client.delete(`${admin}/groups/${id}`),

  // Agents
  agents: () => client.get<Agent[]>(`${admin}/agents`),
  createAgent: (data: Partial<Agent>) =>
    client.post<Agent>(`${admin}/agents`, data),
  updateAgent: (id: number, data: Partial<Agent>) =>
    client.put<Agent>(`${admin}/agents/${id}`, data),
  deleteAgent: (id: number) => client.delete(`${admin}/agents/${id}`),
}
