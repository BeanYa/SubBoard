import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useAuth } from '@/hooks/useAuth'
import { userApi } from '@/api/user'
import { formatBytes } from '@/lib/utils'
import type { UserProfile } from '@/api/user'

export default function Dashboard() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!user) { navigate('/login'); return }
    userApi.profile()
      .then(r => setProfile(r.data))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [user, navigate])

  if (loading || !profile) return <div className="p-8 text-center">加载中...</div>

  const trafficPercent = profile.plan
    ? Math.min(100, (profile.traffic_used / profile.plan.traffic_limit) * 100)
    : 0

  return (
    <div className="max-w-4xl mx-auto space-y-6 p-6">
      <h1 className="text-3xl font-bold">欢迎, {profile.username}</h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm font-medium">套餐</CardTitle></CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{profile.plan?.name || '未分配套餐'}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm font-medium">已用流量</CardTitle></CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatBytes(profile.traffic_used)}</p>
            {profile.plan && <p className="text-xs text-muted-foreground">/ {formatBytes(profile.plan.traffic_limit)}</p>}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm font-medium">状态</CardTitle></CardHeader>
          <CardContent>
            <Badge variant={profile.enabled ? 'success' : 'destructive'}>
              {profile.enabled ? '正常' : '已禁用'}
            </Badge>
            {profile.expire_at && (
              <p className="text-xs text-muted-foreground mt-1">
                到期: {new Date(profile.expire_at).toLocaleDateString('zh-CN')}
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {profile.plan && (
        <Card>
          <CardHeader><CardTitle>流量使用</CardTitle></CardHeader>
          <CardContent>
            <div className="w-full bg-secondary rounded-full h-3">
              <div className="bg-primary h-3 rounded-full transition-all" style={{ width: `${trafficPercent}%` }} />
            </div>
            <p className="text-sm text-muted-foreground mt-2">{trafficPercent.toFixed(1)}% 已使用</p>
          </CardContent>
        </Card>
      )}

      <div className="flex gap-4">
        <Button onClick={() => navigate('/subscribe')} size="lg">获取订阅链接</Button>
        {profile.is_admin && <Button variant="outline" onClick={() => navigate('/admin')}>管理后台</Button>}
      </div>
    </div>
  )
}
