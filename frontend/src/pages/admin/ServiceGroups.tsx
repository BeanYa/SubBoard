import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { adminApi, type ServiceGroup } from '@/api/admin'
import { Plus, RefreshCw, Trash2 } from 'lucide-react'

export default function AdminServiceGroups() {
  const [groups, setGroups] = useState<ServiceGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', description: '', sort_order: 0, subscription_ids: [] as number[], agent_ids: [] as number[], enabled: true })

  const fetch = async () => {
    setLoading(true)
    try { const r = await adminApi.groups(); setGroups(r.data) }
    finally { setLoading(false) }
  }

  useEffect(() => { fetch() }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await adminApi.createGroup(form)
    setShowForm(false)
    fetch()
  }

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除？')) return
    await adminApi.deleteGroup(id)
    fetch()
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">服务群管理</h2>
        <div className="flex gap-2">
          <Button variant="outline" size="icon" onClick={fetch}><RefreshCw className="w-4 h-4" /></Button>
          <Button size="sm" onClick={() => setShowForm(!showForm)}><Plus className="w-4 h-4 mr-1" />新建</Button>
        </div>
      </div>

      {showForm && (
        <Card>
          <CardHeader><CardTitle>新建服务群</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="grid grid-cols-2 gap-4">
              <div><Label>名称</Label><Input value={form.name} onChange={e => setForm({...form, name: e.target.value})} required /></div>
              <div><Label>描述</Label><Input value={form.description} onChange={e => setForm({...form, description: e.target.value})} /></div>
              <div><Label>排序</Label><Input type="number" value={form.sort_order} onChange={e => setForm({...form, sort_order: Number(e.target.value)})} /></div>
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
              <th className="p-3 text-left font-medium">描述</th><th className="p-3 text-left font-medium">状态</th>
              <th className="p-3 text-left font-medium">操作</th>
            </tr></thead>
            <tbody>
              {groups.map(g => (
                <tr key={g.id} className="border-b">
                  <td className="p-3">{g.id}</td>
                  <td className="p-3 font-medium">{g.name}</td>
                  <td className="p-3">{g.description}</td>
                  <td className="p-3"><Badge variant={g.enabled ? 'success' : 'secondary'}>{g.enabled ? '启用' : '禁用'}</Badge></td>
                  <td className="p-3">
                    <Button variant="ghost" size="icon" onClick={() => handleDelete(g.id)}><Trash2 className="w-4 h-4 text-destructive" /></Button>
                  </td>
                </tr>
              ))}
              {groups.length === 0 && !loading && <tr><td colSpan={5} className="p-8 text-center text-muted-foreground">暂无数据</td></tr>}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
