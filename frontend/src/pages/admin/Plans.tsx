import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { adminApi, type Plan } from '@/api/admin'
import { formatBytes } from '@/lib/utils'
import { Plus, RefreshCw, Trash2 } from 'lucide-react'

export default function AdminPlans() {
  const [plans, setPlans] = useState<Plan[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', description: '', traffic_limit: 0, duration_days: 0, price: '', enabled: true })

  const fetchPlans = async () => {
    setLoading(true)
    try {
      const res = await adminApi.plans()
      setPlans(res.data)
    } finally { setLoading(false) }
  }

  useEffect(() => { fetchPlans() }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await adminApi.createPlan(form)
    setShowForm(false)
    setForm({ name: '', description: '', traffic_limit: 0, duration_days: 0, price: '', enabled: true })
    fetchPlans()
  }

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除此套餐？')) return
    await adminApi.deletePlan(id)
    fetchPlans()
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">套餐管理</h2>
        <div className="flex gap-2">
          <Button variant="outline" size="icon" onClick={fetchPlans}><RefreshCw className="w-4 h-4" /></Button>
          <Button size="sm" onClick={() => setShowForm(!showForm)}><Plus className="w-4 h-4 mr-1" />新建套餐</Button>
        </div>
      </div>

      {showForm && (
        <Card>
          <CardHeader><CardTitle>新建套餐</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="grid grid-cols-2 gap-4">
              <div><Label>名称</Label><Input value={form.name} onChange={e => setForm({...form, name: e.target.value})} required /></div>
              <div><Label>描述</Label><Input value={form.description} onChange={e => setForm({...form, description: e.target.value})} /></div>
              <div><Label>流量限额 (GB)</Label><Input type="number" value={form.traffic_limit / (1024**3)} onChange={e => setForm({...form, traffic_limit: Number(e.target.value) * (1024**3)})} /></div>
              <div><Label>有效期 (天, 0=永久)</Label><Input type="number" value={form.duration_days} onChange={e => setForm({...form, duration_days: Number(e.target.value)})} /></div>
              <div><Label>价格</Label><Input value={form.price} onChange={e => setForm({...form, price: e.target.value})} placeholder="¥10/月" /></div>
              <div className="flex items-end col-span-2">
                <Button type="submit">保存</Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="p-3 text-left font-medium">ID</th>
                <th className="p-3 text-left font-medium">名称</th>
                <th className="p-3 text-left font-medium">流量</th>
                <th className="p-3 text-left font-medium">天数</th>
                <th className="p-3 text-left font-medium">价格</th>
                <th className="p-3 text-left font-medium">状态</th>
                <th className="p-3 text-left font-medium">操作</th>
              </tr>
            </thead>
            <tbody>
              {plans.map(p => (
                <tr key={p.id} className="border-b">
                  <td className="p-3">{p.id}</td>
                  <td className="p-3 font-medium">{p.name}</td>
                  <td className="p-3">{formatBytes(p.traffic_limit)}</td>
                  <td className="p-3">{p.duration_days === 0 ? '永久' : `${p.duration_days}天`}</td>
                  <td className="p-3">{p.price}</td>
                  <td className="p-3"><Badge variant={p.enabled ? 'success' : 'secondary'}>{p.enabled ? '启用' : '禁用'}</Badge></td>
                  <td className="p-3">
                    <Button variant="ghost" size="icon" onClick={() => handleDelete(p.id)}><Trash2 className="w-4 h-4 text-destructive" /></Button>
                  </td>
                </tr>
              ))}
              {plans.length === 0 && !loading && (
                <tr><td colSpan={7} className="p-8 text-center text-muted-foreground">暂无套餐</td></tr>
              )}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
