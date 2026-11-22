"use client";
import { useEffect, useRef } from "react";
type Turnstile = {
  render: (el: HTMLElement, options: Record<string, unknown>) => string;
  remove: (id: string) => void;
};
declare global {
  interface Window {
    turnstile?: Turnstile;
  }
}
let ready: Promise<void> | null = null;
export default function Captcha({
  action,
  onToken,
  mock,
  siteKey,
}: {
  action: string;
  onToken: (token: string) => void;
  mock: boolean;
  siteKey: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (mock) {
      onToken("local-test-token");
      return;
    }
    if (!siteKey) return;
    if (!ready)
      ready = new Promise((resolve, reject) => {
        const s = document.createElement("script");
        s.src =
          "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";
        s.async = true;
        s.onload = () => resolve();
        s.onerror = reject;
        document.head.append(s);
      });
    let active = true;
    let widget: string | undefined;
    ready
      .then(() => {
        if (active && ref.current && window.turnstile)
          widget = window.turnstile.render(ref.current, {
            sitekey: siteKey,
            action,
            callback: onToken,
            "expired-callback": () => onToken(""),
          });
      })
      .catch(() => onToken(""));
    return () => {
      active = false;
      if (widget) window.turnstile?.remove(widget);
    };
  }, [action, onToken, mock, siteKey]);
  return (
    <div ref={ref}>
      {!mock && !siteKey && (
        <p className="error">
          Chưa cấu hình xác minh bảo mật. Vui lòng liên hệ hỗ trợ.
        </p>
      )}
    </div>
  );
}
