import { useState, useCallback, useEffect, useRef } from "react";
import type { FormEvent } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { Bell, ChartNoAxesCombined, ClipboardList, Image, Map, Menu, Settings as SettingsIcon, Users, MessagesSquare } from "lucide-react";
import "./App.css";
import { SheetContent } from "./ui";
import Captcha from "./Captcha";
import Settings from "./Settings";
import Catalog from "./Catalog";
import Media from "./Media";
import Bookings from "./Bookings";
import Reports from "./Reports";
import NotificationLog from "./NotificationLog";
import Sales from "./Sales";
import Customers from "./Customers";
import {
  login as loginSession,
  logout as logoutSession,
  restore,
} from "./lib/session";

type User = { full_name: string; permissions: string[] };
const desktopAdmin = "(min-width: 801px)";

export default function App() {
  const [tab, setTab] = useState("catalog");
  const [menuOpen, setMenuOpen] = useState(false);
  const focusContentOnClose = useRef(false);
  const content = useRef<HTMLElement>(null);
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
  useEffect(() => {
    if (!menuOpen) return;
    const screen = window.matchMedia(desktopAdmin);
    const closeOnDesktop = (event: MediaQueryListEvent) => {
      if (event.matches) {
        focusContentOnClose.current = true;
        setMenuOpen(false);
      }
    };
    screen.addEventListener("change", closeOnDesktop);
    return () => screen.removeEventListener("change", closeOnDesktop);
  }, [menuOpen]);
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
  const permissions = user?.permissions ?? [];
  const tabs = [
    { id: "bookings", label: "Đơn tour", icon: ClipboardList, allowed: permissions.includes("booking.read") },
    { id: "reports", label: "Báo cáo", icon: ChartNoAxesCombined, allowed: permissions.includes("report.read") },
    { id: "catalog", label: "Tour & Nội dung", icon: Map, allowed: permissions.includes("catalog.write") || permissions.includes("content.write") },
    { id: "media", label: "Media", icon: Image, allowed: permissions.includes("media.write") },
    { id: "sales", label: "Yêu cầu tư vấn", icon: MessagesSquare, allowed: permissions.includes("lead.write") },
    { id: "customers", label: "Khách hàng", icon: Users, allowed: permissions.includes("customer.write") },
    { id: "notifications", label: "Nhật ký thông báo", icon: Bell, allowed: permissions.includes("notification.read") },
    { id: "settings", label: "Cài đặt", icon: SettingsIcon, allowed: permissions.includes("settings.write") },
  ].filter(entry => entry.allowed);
  const activeTab = tabs.some(entry => entry.id === tab) ? tab : tabs[0]?.id;
  const brand = (
    <a className="admin-brand" href="/" aria-label="Blue Sky Group — Quản trị">
      <img src="/brand/logo.jpg" width="48" height="48" alt="" />
      <span>Blue Sky Group <small>Quản trị</small></span>
    </a>
  );
  function navigation(mobile: boolean) {
    return (
      <nav className="admin-nav" aria-label="Module quản trị">
        {tabs.map(({ id, label, icon: Icon }) => (
          <button key={id} type="button" aria-pressed={activeTab === id} onClick={() => {
            setTab(id);
            if (mobile) {
              focusContentOnClose.current = true;
              setMenuOpen(false);
            }
          }}>
            <Icon size={19} aria-hidden="true" />
            {label}
          </button>
        ))}
      </nav>
    );
  }

  return user ? (
    <div className="admin-shell">
      <aside className="admin-sidebar">
        <div className="admin-sidebar-header">
          {brand}
          <Dialog.Root open={menuOpen} onOpenChange={setMenuOpen}>
            <Dialog.Trigger className="admin-menu-toggle" type="button">
              <Menu size={20} aria-hidden="true" />
              Menu
            </Dialog.Trigger>
            <SheetContent onCloseAutoFocus={(event) => {
              if (!focusContentOnClose.current) return;
              focusContentOnClose.current = false;
              event.preventDefault();
              content.current?.focus({ preventScroll: true });
            }}>
              {navigation(true)}
            </SheetContent>
          </Dialog.Root>
        </div>
        {navigation(false)}
      </aside>
      <main className="admin-content" ref={content} tabIndex={-1}>
        <header className="admin-topbar">
<div><p className="admin-page-label">{tabs.find(entry => entry.id === activeTab)?.label || "Quản trị"}</p><h1>Xin chào, {user.full_name}</h1></div>
          <button
            type="button"
            onClick={async () => {
              try {
                await logoutSession();
                setUser(null);
                setMenuOpen(false);
              } catch {
                setError("Không đăng xuất được. Thử lại.");
              }
            }}
          >
            Đăng xuất
          </button>
        </header>
        {error && <p role="alert">{error}</p>}
        {activeTab === "bookings" && user.permissions.includes("booking.read") && (
          <Bookings canExport={user.permissions.includes("report.export")} write={user.permissions.includes("booking.write")} canPay={user.permissions.includes("payment.write")} canReadNotifications={user.permissions.includes("notification.read")} />
        )}{" "}
        {activeTab === "reports" && user.permissions.includes("report.read") && <Reports canExport={user.permissions.includes("report.export")} />}
        {activeTab === "notifications" && user.permissions.includes("notification.read") && <NotificationLog />}
        {activeTab === "sales" && user.permissions.includes("lead.write") && <Sales canAttach={user.permissions.includes("media.private.write")} />}
        {activeTab === "customers" && user.permissions.includes("customer.write") && (
          <Customers canReadBookings={user.permissions.includes("booking.read")} canLink={user.permissions.includes("booking.write")} />
        )}
        {activeTab === "catalog" && <Catalog permissions={user.permissions} />}{" "}
        {activeTab === "media" && user.permissions.includes("media.write") && (
          <Media />
        )}
        {activeTab === "settings" &&
          user.permissions.includes("settings.write") && <Settings />}
      </main>
    </div>
  ) : (
    <main className="admin-login">
      {brand}
      <h1>Đăng nhập quản trị</h1>
<p>Quản lý tour, yêu cầu tư vấn và hành trình của khách.</p>
      <p className="admin-access-note">Chỉ dành cho nhân viên được cấp quyền.</p>
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
    </main>
  );
}
