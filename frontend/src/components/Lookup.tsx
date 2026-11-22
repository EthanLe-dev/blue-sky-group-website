"use client";
import { useState, useCallback } from "react";
import { api } from "@/lib/api";
import Captcha from "./Captcha";
type Summary = {
  code: string;
  status: string;
  payment_status: string;
  total_amount: string;
  deposit_amount: string;
  paid_amount: string;
  contact_masked: { email: string; phone: string };
  tour: { title: string; starts_at: string };
};
export default function Lookup({
  mock,
  siteKey,
}: {
  mock: boolean;
  siteKey: string;
}) {
  const [token, setToken] = useState("");
  const [booking, setBooking] = useState<Summary | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [captcha, setCaptcha] = useState("");
  const [attempt, setAttempt] = useState(0);
  const onToken = useCallback((t: string) => setCaptcha(t), []);
  return (
    <>
      <p>
        Nhập mã đơn và email hoặc số điện thoại đã dùng đặt tour. Kết quả tra
        cứu chỉ hiển thị thông tin rút gọn.
      </p>
      <form
        className="form-grid"
        onSubmit={async (e) => {
          e.preventDefault();
          const f = new FormData(e.currentTarget);
          setBusy(true);
          setError("");
          try {
            const out = await api<{ booking: Summary; lookup_token: string }>(
              "/booking-lookups",
              {
                method: "POST",
                body: JSON.stringify({
                  code: f.get("code"),
                  contact: f.get("contact"),
                  turnstile_token: captcha,
                }),
              },
            );
            setBooking(out.booking);
            setToken(out.lookup_token);
          } catch {
            setError(
              "Không tra cứu được. Kiểm tra mã đơn, thông tin liên hệ hoặc thử lại sau.",
            );
            setBooking(null);
            setAttempt((n) => n + 1);
          } finally {
            setBusy(false);
          }
        }}
      >
        <label>
          Mã đơn
          <input name="code" required maxLength={24} autoComplete="off" />
        </label>
        <label>
          Email hoặc số điện thoại
          <input name="contact" required maxLength={254} />
        </label>
        <Captcha
          key={attempt}
          action="lookup"
          mock={mock}
          siteKey={siteKey}
          onToken={onToken}
        />
        <button className="btn" disabled={busy || !captcha}>
          {busy ? "Đang tra cứu…" : "Tra cứu đơn"}
        </button>
      </form>
      {error && <p role="alert">{error}</p>}
      {booking && (
        <section className="summary">
          <h2>Đơn {booking.code}</h2>
          <p>{booking.tour.title}</p>
          <p>
            Trạng thái:{" "}
            {{
              pending: "Đang giữ chỗ",
              expired: "Đã hết hạn",
              cancelled: "Đã huỷ",
              confirmed: "Đã xác nhận",
              completed: "Đã hoàn thành",
            }[booking.status] || booking.status}
          </p>
          <p>
            Thanh toán:{" "}
            {booking.payment_status === "unpaid"
              ? "Chưa thanh toán"
              : booking.payment_status}
          </p>
          <p>
            Tổng {Number(booking.total_amount).toLocaleString("vi-VN")} đ · Đã
            thu {Number(booking.paid_amount).toLocaleString("vi-VN")} đ
          </p>
          <p>
            {booking.contact_masked.email} · {booking.contact_masked.phone}
          </p>
          <p>
            Kết quả chỉ có quyền đọc. Thanh toán và thông tin hành khách cần
            phiên đã tạo đơn hoặc nhân viên hỗ trợ.
          </p>
          <button
            onClick={async () => {
              try {
                setBooking(
                  await api<Summary>("/booking-lookups/current", {
                    headers: { Authorization: "Bearer " + token },
                  }),
                );
                setError("");
              } catch {
                setError("Phiên tra cứu đã hết hạn. Vui lòng tra cứu lại.");
              }
            }}
          >
            Cập nhật trạng thái
          </button>
        </section>
      )}
    </>
  );
}
