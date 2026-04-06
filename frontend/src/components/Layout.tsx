import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { LayoutDashboard, Link2, LogOut, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/admin', icon: LayoutDashboard, label: '概览' },
  { to: '/admin/users', icon: Settings, label: '用户' },
  { to: '/admin/plans', icon: LayoutDashboard, label: '套餐' },
  { to: '/admin/subscriptions', icon: Link2, label: '订阅源' },
  { to: '/admin/groups', icon: LayoutDashboard, label: '服务群' },
  { to: '/admin/agents', icon: Link2, label: '节点' },
]

export function Layout() {
  const { user, logout } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  const handleLogout = () => { logout(); navigate('/login') }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b bg-white">
        <div className="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <Link to="/" className="font-bold text-lg">SubManager</Link>
            <nav className="hidden md:flex gap-4">
              <Link to="/dashboard" className="text-sm text-muted-foreground hover:text-foreground">面板</Link>
              <Link to="/subscribe" className="text-sm text-muted-foreground hover:text-foreground">订阅</Link>
            </nav>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-muted-foreground">{user?.username}</span>
            <Button variant="ghost" size="sm" onClick={handleLogout} className="gap-2">
              <LogOut className="w-4 h-4" /> 退出
            </Button>
          </div>
        </div>
      </header>
      <main><Outlet /></main>
    </div>
  )
}

export function AdminLayout() {
  const { user, logout } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="bg-white border-b shadow-sm">
        <div className="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <Link to="/" className="font-bold text-lg">SubManager 管理后台</Link>
          </div>
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="sm" onClick={() => navigate('/dashboard')}>返回用户面板</Button>
            <Button variant="ghost" size="sm" onClick={() => { logout(); navigate('/login') }} className="gap-2">
              <LogOut className="w-4 h-4" /> 退出
            </Button>
          </div>
        </div>
      </header>
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-[200px_1fr] gap-6">
          <aside className="space-y-1">
            {navItems.map(item => (
              <Link
                key={item.to}
                to={item.to}
                className={cn(
                  'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors',
                  location.pathname === item.to || (item.to !== '/admin' && location.pathname.startsWith(item.to))
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                )}
              >
                <item.icon className="w-4 h-4" />
                {item.label}
              </Link>
            ))}
          </aside>
          <div><Outlet /></div>
        </div>
      </div>
    </div>
  )
}
