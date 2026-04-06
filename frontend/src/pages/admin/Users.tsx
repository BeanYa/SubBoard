import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { adminApi, type User } from '@/api/admin'
import { formatBytes } from '@/lib/utils'
import { Plus, Search, RefreshCw } from 'lucide-react'

export default function AdminUsers() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [editingPassword, setEditingPassword] = useState<number | null>(null)
  const [newPassword, setNewPassword] = useState('')

  const fetchUsers = async () => {
    setLoading(true)
    try {
      const res = await adminApi.users(page, 20, search)
      setUsers(res.data.items)
      setTotal(res.data.total)
    } finally { setLoading(false) }
  }

  useEffect(() => { fetchUsers() }, [page, search])

  const handleToggle = async (id: number) => {
    await adminApi.toggleUser(id)
    fetchUsers()
  }

  const handleReset = async (id: number) => {
    await adminApi.resetTraffic(id)
    fetchUsers()
  }

  const handleChangePassword = async (id: number) => {
    if (!newPassword || newPassword.length < 6) {
      alert('密码长度至少6位')
      return
    }
    await adminApi.changePassword(id, newPassword)
    setEditingPassword(null)
    setNewPassword('')
    alert('密码修改成功')
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">用户管理</h2>
        <div className="flex gap-2">
          <Input placeholder="搜索用户名..." value={search} onChange={e => { setSearch(e.target.value); setPage(1) }} className="w-64" />
          <Button variant="outline" size="icon" onClick={fetchUsers}><RefreshCw className="w-4 h-4" /></Button>
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="p-3 text-left font-medium">ID</th>
                <th className="p-3 text-left font-medium">用户名</th>
                <th className="p-3 text-left font-medium">套餐</th>
                <th className="p-3 text-left font-medium">流量</th>
                <th className="p-3 text-left font-medium">状态</th>
                <th className="p-3 text-left font-medium">操作</th>
              </tr>
            </thead>
            <tbody>
              {users.map(u => (
                <tr key={u.id} className="border-b">
                  <td className="p-3">{u.id}</td>
                  <td className="p-3 font-medium">{u.username}</td>
                  <td className="p-3">{u.plan_id ?? '-'}</td>
                  <td className="p-3">{formatBytes(u.traffic_used)}</td>
                  <td className="p-3">
                    <Badge variant={u.enabled ? 'success' : 'destructive'}>{u.enabled ? '正常' : '禁用'}</Badge>
                    {u.is_admin && <Badge variant="secondary" className="ml-1">管理员</Badge>}
                  </td>
                  <td className="p-3">
                    <div className="flex gap-1 flex-wrap">
                      <Button variant="ghost" size="sm" onClick={() => handleToggle(u.id)}>切换</Button>
                      <Button variant="ghost" size="sm" onClick={() => handleReset(u.id)}>重置流量</Button>
                      <Button variant="ghost" size="sm" onClick={() => setEditingPassword(u.id)}>改密码</Button>
                      {editingPassword === u.id && (
                        <div className="flex gap-1 items-center bg-muted p-2 rounded">
                          <Input
                            type="password"
                            placeholder="新密码"
                            value={newPassword}
                            onChange={e => setNewPassword(e.target.value)}
                            className="w-32 h-8"
                          />
                          <Button size="sm" onClick={() => handleChangePassword(u.id)}>确认</Button>
                          <Button size="sm" variant="ghost" onClick={() => { setEditingPassword(null); setNewPassword('') }}>取消</Button>
                        </div>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {users.length === 0 && !loading && (
                <tr><td colSpan={6} className="p-8 text-center text-muted-foreground">暂无数据</td></tr>
              )}
            </tbody>
          </table>
        </CardContent>
      </Card>

      <div className="flex justify-center gap-2">
        <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>上一页</Button>
        <span className="px-4 py-2 text-sm">第 {page} 页，共 {total} 条</span>
        <Button variant="outline" size="sm" onClick={() => setPage(p => p + 1)}>下一页</Button>
      </div>
    </div>
  )
}
