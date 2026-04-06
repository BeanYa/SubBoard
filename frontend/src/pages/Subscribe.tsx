import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuth } from '@/hooks/useAuth'
import { userApi } from '@/api/user'
import { formatBytes } from '@/lib/utils'
import type { SubscriptionDetail } from '@/api/user'

export default function Subscribe() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [sub, setSub] = useState<SubscriptionDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [copied, setCopied] = useState('')

  useEffect(() => {
    if (!user) { navigate('/login'); return }
    userApi.subscription()
      .then(r => setSub(r.data))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [user, navigate])

  const copy = (text: string, key: string) => {
    navigator.clipboard.writeText(text)
    setCopied(key)
    setTimeout(() => setCopied(''), 2000)
  }

  if (loading || !sub) return <div className="p-8 text-center">加载中...</div>

  return (
    <div className="max-w-3xl mx-auto space-y-6 p-6">
      <h1 className="text-3xl font-bold">订阅链接</h1>

      <Card>
        <CardHeader>
          <CardTitle>{sub.plan_name || '未分配套餐'}</CardTitle>
          <CardDescription>
            已用 {formatBytes(sub.traffic_used)}
            {sub.traffic_limit > 0 && ` / ${formatBytes(sub.traffic_limit)}`}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2 items-center">
            <Badge variant={sub.plan_name ? 'default' : 'secondary'}>
              {sub.plan_name ? '已激活' : '未激活'}
            </Badge>
            {sub.expire_at && (
              <Badge variant="warning">到期: {new Date(sub.expire_at).toLocaleDateString('zh-CN')}</Badge>
            )}
          </div>

          <Tabs defaultValue="clash">
            <TabsList>
              <TabsTrigger value="clash">Clash</TabsTrigger>
              <TabsTrigger value="singbox">Sing-box</TabsTrigger>
              <TabsTrigger value="base64">Base64</TabsTrigger>
            </TabsList>

            <TabsContent className="space-y-2">
              <div className="flex gap-2">
                <Input value={sub.sub_url_clash} readOnly className="font-mono text-xs" />
                <Button size="sm" onClick={() => copy(sub.sub_url_clash, 'clash')}>{copied === 'clash' ? '已复制' : '复制'}</Button>
              </div>
              <p className="text-xs text-muted-foreground">适用于 Clash/OpenClash 等客户端</p>
            </TabsContent>

            <TabsContent className="space-y-2">
              <div className="flex gap-2">
                <Input value={sub.sub_url_singbox} readOnly className="font-mono text-xs" />
                <Button size="sm" onClick={() => copy(sub.sub_url_singbox, 'singbox')}>{copied === 'singbox' ? '已复制' : '复制'}</Button>
              </div>
              <p className="text-xs text-muted-foreground">适用于 sing-box 客户端</p>
            </TabsContent>

            <TabsContent className="space-y-2">
              <div className="flex gap-2">
                <Input value={sub.sub_url_base64} readOnly className="font-mono text-xs" />
                <Button size="sm" onClick={() => copy(sub.sub_url_base64, 'base64')}>{copied === 'base64' ? '已复制' : '复制'}</Button>
              </div>
              <p className="text-xs text-muted-foreground">适用于 Surge / 原始订阅导入</p>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
