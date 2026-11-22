import { useEffect, useRef, useState } from "react";
import { authorized } from "./lib/session";
import { ApiError } from "./lib/api";

type Payment = { id: string; amount: string; purpose: string; status: string; expires_at: string };
type Refund = { id: string; payment_id: string; amount: string; reason: string };
type Ledger = { server_time: string; payments: Payment[]; refunds: Refund[]; reconciled: boolean };
type Notifications = {
  logs: { id: string; channel: string; subject: string | null; body: string; created_at: string }[];
  jobs: { id: string; template_code: string; status: string; attempts: number; error_code: string | null }[];
};
const money = (amount: string) => BigInt(amount).toLocaleString("vi-VN") + " ₫";
const purposes: Record<string, string> = { full: "Toàn bộ", deposit: "Đặt cọc", balance: "Phần còn lại" };

export default function BookingPayments({ booking, canPay, canReadNotifications, onChanged }: {
  booking: { id: string; status: string; payment_option: string; balance_amount: string; refunded_amount: string; version: number };
  canPay: boolean;
  canReadNotifications: boolean;
  onChanged: () => Promise<void>;
}) {
  const [ledger, setLedger] = useState<Ledger | null>(null);
  const [notifications, setNotifications] = useState<Notifications | null>(null);
  const [disabled, setDisabled] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const inFlight = useRef(false);
  const keys = useRef<Record<string, string>>({});
  const path = "/admin/bookings/" + booking.id + "/payments";
  const notificationPath = "/admin/notifications?booking_id=" + booking.id;

  useEffect(() => {
    const controller = new AbortController();
    authorized<Ledger>(path, { signal: controller.signal }).then(setLedger).catch((e) => {
      if (controller.signal.aborted) return;
      if (e instanceof ApiError && e.status === 404) setDisabled(true);
      else setError(String(e));
    });
    if (canReadNotifications) {
      authorized<Notifications>(notificationPath, { signal: controller.signal }).then(setNotifications).catch((e) => {
        if (!controller.signal.aborted && !(e instanceof ApiError && e.status === 404)) setError(String(e));
      });
    }
    return () => controller.abort();
  }, [path, notificationPath, canReadNotifications, booking.version]);

  async function run(action: () => Promise<void>) {
    if (inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    setMessage("");
    setError("");
    try {
      await action();
    } catch (e) {
      const code = e instanceof ApiError ? e.code : "";
      const messages: Record<string, string> = {
        PAYMENT_EXPIRED: "Mã đã hết hạn. Hãy tạo mã mới.",
        HOLD_EXPIRED: "Giữ chỗ đã hết hạn.",
        REFUND_UNAVAILABLE: "Số tiền hoàn vượt quá khoản còn lại của giao dịch.",
      };
      setError(messages[code] || String(e));
      if (code === "PAYMENT_EXPIRED" || code === "HOLD_EXPIRED" || code === "PAYMENT_IN_PROGRESS" || code === "INVALID_STATE") {
        authorized<Ledger>(path).then(setLedger).catch(() => {});
      }
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  async function post(url: string, body: object, label: string) {
    const serialized = JSON.stringify(body);
    const scope = url + serialized;
    const key = keys.current[scope] ||= crypto.randomUUID();
    await authorized(url, {
      method: "POST",
      headers: { "Idempotency-Key": key },
      body: serialized,
    });
    delete keys.current[scope];
    setMessage(label);
    setLedger(null);
    try {
      setLedger(await authorized<Ledger>(path));
      await onChanged();
    } catch {
      setError("Đã ghi nhận giao dịch. Tải lại đơn để cập nhật sổ giao dịch.");
    }
  }

  const purpose = booking.status === "pending" ? booking.payment_option :
    booking.status === "confirmed" && booking.balance_amount !== "0" && booking.refunded_amount === "0" ? "balance" : "";
  const pending = ledger?.payments.find(p => p.status === "pending" && new Date(p.expires_at) > new Date(ledger.server_time));

  if (disabled && !notifications) return null;
  return (
    <section className="editor">
      <h3>Thanh toán và hoàn tiền giả lập</h3>
      {ledger && (
        <>
          <p>{ledger.reconciled ? "Sổ thu và hoàn khớp với đơn." : "Số tiền trên đơn không khớp sổ thu và hoàn."}</p>
          {ledger.payments.length === 0 && <p>Chưa có giao dịch.</p>}
          {canPay && purpose && !pending && (
            <button disabled={busy} onClick={() => void run(() => post(path, { purpose }, "Đã tạo mã thanh toán giả lập."))}>
              Tạo mã thanh toán · {purposes[purpose]}
            </button>
          )}
          {ledger.payments.map(p => {
            const refunds = ledger.refunds.filter(r => r.payment_id === p.id);
            const remaining = refunds.reduce((sum, r) => sum - BigInt(r.amount), BigInt(p.amount));
            return (
              <fieldset key={p.id}>
                <legend>{purposes[p.purpose]} · {money(p.amount)}</legend>
                <p>{p.status === "succeeded" ? "Đã xác nhận giả lập" :
                  p.id === pending?.id && purpose ? "Chờ xác nhận" : "Mã không còn hiệu lực"}</p>
                {refunds.map(r => <p key={r.id}>Đã hoàn giả lập {money(r.amount)} · {r.reason}</p>)}
                {canPay && purpose && p.id === pending?.id && (
                  <form onSubmit={(e) => {
                    e.preventDefault();
                    const form = e.currentTarget;
                    const reason = String(new FormData(form).get("reason"));
                    void run(() => post("/admin/payments/" + p.id + "/mock-confirm", { reason }, "Đã xác nhận thanh toán giả lập."));
                  }}>
                    <label>Lý do xác nhận<input name="reason" required minLength={2} maxLength={1000} /></label>
                    <button disabled={busy}>Xác nhận thanh toán giả lập</button>
                  </form>
                )}
                {canPay && p.status === "succeeded" && remaining > 0n && (
                  <form onSubmit={(e) => {
                    e.preventDefault();
                    const form = e.currentTarget;
                    const fields = new FormData(form);
                    const amount = String(fields.get("amount"));
                    const reason = String(fields.get("reason"));
                    void run(async () => {
                      await post("/admin/payments/" + p.id + "/refunds", { amount, reason }, "Đã ghi nhận hoàn tiền giả lập.");
                      form.reset();
                    });
                  }}>
                    <label>Số tiền hoàn (₫)<input name="amount" inputMode="numeric" required pattern="[1-9][0-9]*" maxLength={19} /></label>
                    <label>Lý do hoàn tiền<input name="reason" required minLength={2} maxLength={1000} /></label>
                    <p>Có thể hoàn tối đa {money(String(remaining))}.</p>
                    <button disabled={busy}>Ghi nhận hoàn tiền giả lập</button>
                  </form>
                )}
              </fieldset>
            );
          })}
        </>
      )}
      {canReadNotifications && notifications && (
        <div>
          <h4>Email/SMS giả lập</h4>
          <button disabled={busy} onClick={() => void run(async () => {
            setNotifications(await authorized<Notifications>(notificationPath));
          })}>Tải lại nhật ký thông báo</button>
          {notifications.logs.length === 0 && <p>Chưa có bản xem trước.</p>}
          {notifications.jobs.map(e => <p key={e.id}>{e.template_code} · {e.status === "failed" ? "Tạo bản xem trước thất bại" : "Đang chờ tạo bản xem trước"} · {e.attempts}/5 lần</p>)}
          {notifications.logs.map(n => (
            <details key={n.id}>
              <summary>{n.channel === "email" ? "Email" : "SMS"} giả lập · {new Date(n.created_at).toLocaleString("vi-VN")}</summary>
              {n.subject && <p><strong>{n.subject}</strong></p>}
              <p>{n.body}</p>
            </details>
          ))}
        </div>
      )}
      {message && <p role="status">{message}</p>}
      {error && <p role="alert">{error}</p>}
    </section>
  );
}
