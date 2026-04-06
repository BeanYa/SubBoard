import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { adminApi, type Agent } from '@/api/admin'
import { Plus, RefreshCw, Trash2 } from 'lucide-react'

export default function AdminAgents() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', server_addr: '', port: 443, protocol: 'vless', protocol_config: {}, enabled: true })

  const fetch = async () => {
    setLoading(true)
    try { const r = await adminApi.agents(); setAgents(r.data) }
    finally { setLoading(false) }
  }

  useEffect(() => { fetch() }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await adminApi.createAgent(form)
    setShowForm(false)
    fetch()
  }

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除？')) return
    await adminApi.deleteAgent(id)
    fetch()
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">节点管理</h2>
        <div className="flex gap-2">
          <Button variant="outline" size="icon" onClick={fetch}><RefreshCw className="w-4 h-4" /></Button>
          <Button size="sm" onClick={() => setShowForm(!showForm)}><Plus className="w-4 h-4 mr-1" />添加</Button>
        </div>
      </div>

      {showForm && (
        <Card>
          <CardHeader><CardTitle>添加节点</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="grid grid-cols-2 gap-4">
              <div><Label>名称</Label><Input value={form.name} onChange={e => setForm({...form, name: e.target.value})} required /></div>
              <div><Label>地址</Label><Input value={form.server_addr} onChange={e => setForm({...form, server_addr: e.target.value})} /></div>
              <div><Label>端口</Label><Input type="number" value={form.port} onChange={e => setForm({...form, port: Number(e.target.value)})} /></div>
              <div><Label>协议</Label>
                <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={form.protocol} onChange={e => setForm({...form, protocol: e.target.value})}>
                  <option value="vless">VLESS</option><option value="vmess">VMess</option>
                  <option value="shadowsocks">Shadowsocks</option><option value="trojan">Trojan</option>
                </select>
              </div>
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
              <th className="p-3 text-left font-medium">地址</th><th className="p-3 text-left font-medium">协议</th>
              <th className="p-3 text-left font-medium">状态</th><th className="p-3 text-left font-medium">操作</th>
            </tr></thead>
            <tbody>
              {agents.map(a => (
                <tr key={a.id} className="border-b">
                  <td className="p-3">{a.id}</td>
                  <td className="p-3 font-medium">{a.name}</td>
                  <td className="p-3 font-mono text-xs">{a.server_addr}:{a.port}</td>
                  <td className="p-3"><Badge variant="secondary">{a.protocol}</Badge></td>
                  <td className="p-3">
                    <Badge variant={a.status === 'online' ? 'success' : a.status === 'offline' ? 'destructive' : 'secondary'}>
                      {a.status}
                    </Badge>
                  </td>
                  <td className="p-3">
                    <Button variant="ghost" size="icon" onClick={() => handleDelete(a.id)}><Trash2 className="w-4 h-4 text-destructive" /></Button>
                  </td>
                </tr>
              ))}
              {agents.length === 0 && !loading && <tr><td colSpan={6} className="p-8 text-center text-muted-foreground">暂无数据</td></tr>}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
