import { useState } from 'react'
import type { FormEvent } from 'react'
import './App.css'

type User = { full_name: string; permissions: string[] }
export default function App() {
  const [user, setUser] = useState<User | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  async function login(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setLoading(true); setError('')
    const form = new FormData(event.currentTarget)
    try {
      const response = await fetch('/api/v1/admin/auth/login', { method: 'POST', credentials: 'include', headers: {'Content-Type':'application/json'}, body: JSON.stringify({ email: form.get('email'), password: form.get('password'), turnstile_token: import.meta.env.DEV ? 'local-test-token' : form.get('cf-turnstile-response') }) })
      const result = await response.json()
      if (!response.ok || !result.success) throw new Error('Không đăng nhập được. Kiểm tra thông tin và thử lại.')
      setUser(result.data.user)
    } catch (reason) { setError(reason instanceof Error ? reason.message : 'Không kết nối được máy chủ.') }
    finally { setLoading(false) }
  }
  return <main><p>BLUE SKY GROUP · Nền tảng quản trị</p>{user ? <><h1>Xin chào, {user.full_name}</h1><p>Phiên đăng nhập đã được xác thực. Các module quản trị sẽ được bổ sung theo kế hoạch.</p></> : <><h1>Đăng nhập quản trị</h1><p>Chỉ dành cho nhân viên được cấp quyền.</p><form onSubmit={login}><label htmlFor="email">Email</label><input id="email" name="email" type="email" autoComplete="username" required maxLength={254}/><label htmlFor="password">Mật khẩu</label><input id="password" name="password" type="password" autoComplete="current-password" required maxLength={72}/>{error && <p role="alert">{error}</p>}<button disabled={loading}>{loading ? 'Đang đăng nhập…' : 'Đăng nhập'}</button></form>{!import.meta.env.DEV && <p>Form demo công khai cần widget Turnstile trước khi sử dụng (T107 UI integration).</p>}</>}</main>
}
