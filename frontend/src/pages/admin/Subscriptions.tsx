import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { adminApi, type SubscriptionSource } from '@/api/admin'
import { Plus, RefreshCw, Trash2, RefreshCw as RefreshIcon } from 'lucide-react'

export default function AdminSubscriptions() {
  const [subs, setSubs] = useState<SubscriptionSource[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', type: 'url' as const, url: '', headers: {}, refresh_interval: 0, enabled: true })

  const fetch = async () => {
    setLoading(true)
    try { const r = await adminApi.subscriptions(); setSubs(r.data) }
    finally { setLoading(false) }
  }

  useEffect(() => { fetch() }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await adminApi.createSubscription(form)
    setShowForm(false)
    setForm({ name: '', type: 'url', url: '', headers: {}, refresh_interval: 0, enabled: true })
    fetch()
  }

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除？')) return
    await adminApi.deleteSubscription(id)
    fetch()
  }

  const handleRefresh = async (id: number) => {
    await adminApi.refreshSubscription(id)
    fetch()
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">订阅源管理</h2>
        <div className="flex gap-2">
          <Button variant="outline" size="icon" onClick={fetch}><RefreshCw className="w-4 h-4" /></Button>
          <Button size="sm" onClick={() => setShowForm(!showForm)}><Plus className="w-4 h-4 mr-1" />添加</Button>
        </div>
      </div>

      {showForm && (
        <Card>
          <CardHeader><CardTitle>添加订阅源</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="grid grid-cols-2 gap-4">
              <div><Label>名称</Label><Input value={form.name} onChange={e => setForm({...form, name: e.target.value})} required /></div>
              <div><Label>类型</Label>
                <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={form.type} onChange={e => setForm({...form, type: e.target.value as any})}>
                  <option value="url">URL</option><option value="substore">SubStore</option><option value="raw">Raw</option>
                </select>
              </div>
              <div className="col-span-2"><Label>URL</Label><Input value={form.url} onChange={e => setForm({...form, url: e.target.value})} placeholder="https://..." /></div>
              <div className="col-span-2"><Button type="submit">保存</Button></div>
            </form>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead><tr className="border-b bg-muted/50">
              <th className="p-3 text-left font-medium">ID</th><th className="p-3 text-left font-medium">名称</th>
              <th className="p-3 text-left font-medium">类型</th><th className="p-3 text-left font-medium">节点数</th>
              <th className="p-3 text-left font-medium">最后拉取</th><th className="p-3 text-left font-medium">状态</th>
              <th className="p-3 text-left font-medium">操作</th>
            </tr></thead>
            <tbody>
              {subs.map(s => (
                <tr key={s.id} className="border-b">
                  <td className="p-3">{s.id}</td>
                  <td className="p-3 font-medium">{s.name}</td>
                  <td className="p-3"><Badge variant="secondary">{s.type}</Badge></td>
                  <td className="p-3">{s.node_count}</td>
                  <td className="p-3">{s.last_fetch_at ? new Date(s.last_fetch_at).toLocaleString('zh-CN') : '-'}</td>
                  <td className="p-3"><Badge variant={s.enabled ? 'success' : 'secondary'}>{s.enabled ? '启用' : '禁用'}</Badge></td>
                  <td className="p-3">
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" onClick={() => handleRefresh(s.id)}><RefreshIcon className="w-4 h-4" /></Button>
                      <Button variant="ghost" size="icon" onClick={() => handleDelete(s.id)}><Trash2 className="w-4 h-4 text-destructive" /></Button>
                    </div>
                  </td>
                </tr>
              ))}
              {subs.length === 0 && !loading && <tr><td colSpan={7} className="p-8 text-center text-muted-foreground">暂无数据</td></tr>}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
