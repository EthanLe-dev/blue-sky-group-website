import { useState, useCallback, useEffect } from "react";
import type { FormEvent } from "react";
import "./App.css";
import Captcha from "./Captcha";
import Settings from "./Settings";
import Catalog from "./Catalog";
import Media from "./Media";
import Bookings from "./Bookings";
import NotificationLog from "./NotificationLog";
import Sales from "./Sales";
import Customers from "./Customers";
import {
  login as loginSession,
  logout as logoutSession,
  restore,
} from "./lib/session";

type User = { full_name: string; permissions: string[] };
export default function App() {
  const [tab, setTab] = useState("catalog");
  const [user, setUser] = useState<User | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [captcha, setCaptcha] = useState(
    import.meta.env.DEV || import.meta.env.VITE_TURNSTILE_MODE === "mock"
      ? "local-test-token"
      : "",
  );
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    let active = true;
    restore().then((user) => {
      if (active && user) setUser(user);
    });
    return () => {
      active = false;
    };
  }, []);
  const onToken = useCallback((token: string) => setCaptcha(token), []);
  async function login(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      setUser(
        await loginSession(
          String(form.get("email")),
          String(form.get("password")),
          captcha,
        ),
      );
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "Không kết nối được máy chủ.",
      );
      setAttempt((n) => n + 1);
    } finally {
      setLoading(false);
    }
  }
  return (
    <main>
      <a className="admin-brand" href="/">
        <img
          src="/brand/logo.jpg"
          width="72"
          height="72"
          alt="Logo Blue Sky Group"
        />
        BLUE SKY GROUP · Quản trị
      </a>
      {user ? (
        <>
          <h1>Xin chào, {user.full_name}</h1>
          <button
            onClick={async () => {
              try {
                await logoutSession();
                setUser(null);
              } catch {
                setError("Không đăng xuất được. Thử lại.");
              }
            }}
          >
            Đăng xuất
          </button>
          {error && <p role="alert">{error}</p>}
          <nav className="admin-tabs" aria-label="Module quản trị">
            {[
              ["bookings", "Đơn tour"],
              ["catalog", "Tour & Nội dung"],
              ["media", "Media"],
              ...(user.permissions.includes("lead.write") ? [["sales", "Yêu cầu tư vấn"]] : []),
              ...(user.permissions.includes("customer.write") ? [["customers", "Khách hàng"]] : []),
              ...(user.permissions.includes("notification.read") ? [["notifications", "Nhật ký thông báo"]] : []),
              ["settings", "Cài đặt"],
            ].map(([id, label]) => (
              <button
                key={id}
                aria-pressed={tab === id}
                onClick={() => setTab(id)}
              >
                {label}
              </button>
            ))}
          </nav>
          {tab === "bookings" && user.permissions.includes("booking.read") && (
            <Bookings write={user.permissions.includes("booking.write")} canPay={user.permissions.includes("payment.write")} canReadNotifications={user.permissions.includes("notification.read")} />
          )}{" "}
          {tab === "notifications" && user.permissions.includes("notification.read") && <NotificationLog />}
          {tab === "sales" && user.permissions.includes("lead.write") && <Sales canAttach={user.permissions.includes("media.private.write")} />}
          {tab === "customers" && user.permissions.includes("customer.write") && (
            <Customers canReadBookings={user.permissions.includes("booking.read")} canLink={user.permissions.includes("booking.write")} />
          )}
          {tab === "catalog" && <Catalog permissions={user.permissions} />}{" "}
          {tab === "media" && user.permissions.includes("media.write") && (
            <Media />
          )}
          {tab === "settings" &&
            user.permissions.includes("settings.write") && <Settings />}
        </>
      ) : (
        <>
          <h1>Đăng nhập quản trị</h1>
          <p>Chỉ dành cho nhân viên được cấp quyền.</p>
          <form onSubmit={login}>
            <label htmlFor="email">Email</label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="username"
              required
              maxLength={254}
            />
            <label htmlFor="password">Mật khẩu</label>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              required
              maxLength={72}
            />
            {error && <p role="alert">{error}</p>}
            <Captcha key={attempt} onToken={onToken} />
            <button disabled={loading || !captcha}>
              {loading ? "Đang đăng nhập…" : "Đăng nhập"}
            </button>
          </form>
        </>
      )}
    </main>
  );
}
