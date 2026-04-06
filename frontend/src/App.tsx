import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import Dashboard from '@/pages/Dashboard'
import Subscribe from '@/pages/Subscribe'
import { Layout, AdminLayout } from '@/components/Layout'
import AdminUsers from '@/pages/admin/Users'
import AdminPlans from '@/pages/admin/Plans'
import AdminSubscriptions from '@/pages/admin/Subscriptions'
import AdminServiceGroups from '@/pages/admin/ServiceGroups'
import AdminAgents from '@/pages/admin/Agents'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

function AdminRoute({ children }: { children: React.ReactNode }) {
  const user = localStorage.getItem('user')
  if (!user) return <Navigate to="/login" replace />
  const isAdmin = JSON.parse(user).is_admin
  if (!isAdmin) return <Navigate to="/dashboard" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route element={<PrivateRoute><Layout /></PrivateRoute>}>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/subscribe" element={<Subscribe />} />
        </Route>
        <Route element={<AdminRoute><AdminLayout /></AdminRoute>}>
          <Route path="/admin" element={<AdminUsers />} />
          <Route path="/admin/users" element={<AdminUsers />} />
          <Route path="/admin/plans" element={<AdminPlans />} />
          <Route path="/admin/subscriptions" element={<AdminSubscriptions />} />
          <Route path="/admin/groups" element={<AdminServiceGroups />} />
          <Route path="/admin/agents" element={<AdminAgents />} />
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
