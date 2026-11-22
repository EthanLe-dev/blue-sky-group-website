"use client";
import { useState, useCallback, useEffect, useRef } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import Captcha from "./Captcha";
import type { Surcharge } from "@/lib/catalog";
type Counts = { adult: number; child: number; infant: number };
type Person = {
  passenger_type: string;
  full_name: string;
  date_of_birth: string;
};
type Contact = { name: string; email: string; phone: string };
type Quote = {
  id: string;
  counts: Counts;
  seat_count: number;
  price_version: number;
  expires_at: string;
  server_time: string;
  allowed_payment_options: string[];
  balance_due_at: string;
  price_breakdown: {
    total_amount: string;
    deposit_amount: string;
    balance_after_deposit: string;
    items: {
      code: string;
      label: string;
      quantity: number;
      unit_amount: string;
      total_amount: string;
    }[];
    cancellation_policy: string;
  };
  hold_minutes: number;
};
type Booking = {
  id: string;
  code: string;
  status: string;
  total_amount: string;
  deposit_amount: string;
  balance_amount: string;
  expires_at: string;
  server_time: string;
  payment_option: string;
  version: number;
  tour_snapshot: { title: string; starts_at: string };
};
const money = (v: string) => Number(v).toLocaleString("vi-VN") + " ₫";
const labels: Record<string, string> = {
  adult: "Người lớn",
  child: "Trẻ em",
  infant: "Em bé",
};
function Timer({ expires, server }: { expires: string; server: string }) {
  const [remaining, setRemaining] = useState(
    Math.max(0, new Date(expires).getTime() - new Date(server).getTime()),
  );
  useEffect(() => {
    const duration = new Date(expires).getTime() - new Date(server).getTime();
    const started = performance.now();
    const interval = setInterval(
      () => setRemaining(Math.max(0, duration - (performance.now() - started))),
      1000,
    );
    return () => clearInterval(interval);
  }, [expires, server]);
  return (
    <p role="timer">
      {remaining > 0
        ? "Còn " +
          Math.floor(remaining / 60000) +
          ":" +
          String(Math.floor(remaining / 1000) % 60).padStart(2, "0") +
          " để hoàn tất giữ chỗ"
        : "Hạn giữ chỗ đã kết thúc. Hãy tải lại trạng thái."}
    </p>
  );
}
export function BookingResult({ initial }: { initial: Booking }) {
  const [booking, setBooking] = useState(initial);
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    api<Booking>("/bookings/" + initial.id)
      .then((b) => {
        if (active) setBooking(b);
      })
      .catch(() => {
        if (active)
          setError(
            "Chưa cập nhật được giờ máy chủ. Bấm cập nhật trạng thái để thử lại.",
          );
      });
    return () => {
      active = false;
    };
  }, [initial.id]);
  return (
    <section className="summary">
      <h2>
        {booking.status === "pending"
          ? "Đã tạo đơn giữ chỗ"
          : "Trạng thái đơn: " +
            ({
              expired: "Đã hết hạn giữ chỗ",
              cancelled: "Đã huỷ",
              confirmed: "Đã xác nhận",
              completed: "Đã hoàn thành",
            }[booking.status] || booking.status)}
      </h2>
      <p>
        Mã đơn <strong>{booking.code}</strong>. Lưu mã này để tra cứu bằng email
        hoặc số điện thoại.
      </p>
      <p>{booking.tour_snapshot.title}</p>
      <Timer
        key={booking.server_time}
        expires={booking.expires_at}
        server={booking.server_time}
      />
      <dl>
        <dt>Tổng tiền</dt>
        <dd>{money(booking.total_amount)}</dd>
        <dt>Tiền cọc theo chính sách</dt>
        <dd>{money(booking.deposit_amount)}</dd>
        <dt>Còn phải thanh toán</dt>
        <dd>{money(booking.balance_amount)}</dd>
      </dl>
      <p className="notice">
        Giả lập — không chuyển tiền thật. Thanh toán và email/SMS giả lập được
        triển khai ở M4. Đơn chưa được xác nhận thanh toán và sẽ tự hết hạn giữ
        chỗ.
      </p>
      <button disabled>Thanh toán giả lập — sắp có ở M4</button>
      <button
        onClick={async () => {
          try {
            setBooking(await api<Booking>("/bookings/" + booking.id));
            setError("");
          } catch (e) {
            setError(String(e));
          }
        }}
      >
        Cập nhật trạng thái đơn
      </button>
      {error && <p role="alert">{error}</p>}
      <Link className="text-action" href="/lookup">
        Tra cứu đơn
      </Link>
    </section>
  );
}
export default function Checkout({
  departure,
  title,
  prices,
  surcharges,
  mock,
  siteKey,
}: {
  departure: string;
  title: string;
  surcharges: Surcharge[];
  prices: {
    passenger_type: string;
    min_age: number;
    max_age_exclusive: number | null;
    unit_price: string;
  }[];
  mock: boolean;
  siteKey: string;
}) {
  const [codes, setCodes] = useState<string[]>([]);
  const [counts, setCounts] = useState<Counts>({
    adult: 1,
    child: 0,
    infant: 0,
  });
  const [quote, setQuote] = useState<Quote | null>(null);
  const [contact, setContact] = useState<Contact>({
    name: "",
    email: "",
    phone: "",
  });
  const [people, setPeople] = useState<Person[]>([]);
  const [option, setOption] = useState("full");
  const [requests, setRequests] = useState("");
  const [consent, setConsent] = useState(false);
  const [step, setStep] = useState(1);
  const [csrf, setCsrf] = useState("");
  const [captcha, setCaptcha] = useState("");
  const [attempt, setAttempt] = useState(0);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [booking, setBooking] = useState<Booking | null>(null);
  const key = useRef("");
  const payload = useRef<string>("");
  const onToken = useCallback((t: string) => setCaptcha(t), []);
  async function getQuote() {
    setBusy(true);
    setError("");
    try {
      const session = await api<{ csrf_token: string }>("/guest-sessions", {
        method: "POST",
        body: "{}",
      });
      setCsrf(session.csrf_token);
      const q = await api<Quote>("/booking-quotes", {
        method: "POST",
        headers: { "X-CSRF-Token": session.csrf_token },
        body: JSON.stringify({
          departure_id: departure,
          counts,
          surcharge_codes: codes,
        }),
      });
      setQuote(q);
      setPeople((old) =>
        Object.entries(counts).flatMap(([type, n]) =>
          Array.from(
            { length: n },
            (_, i) =>
              old.filter((p) => p.passenger_type === type)[i] || {
                passenger_type: type,
                full_name: "",
                date_of_birth: "",
              },
          ),
        ),
      );
      setOption(q.allowed_payment_options.includes(option) ? option : "full");
      setConsent(false);
      setStep(2);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không tính được giá.");
    } finally {
      setBusy(false);
    }
  }
  async function create() {
    if (!quote) return;
    setBusy(true);
    setError("");
    const business = {
      quote_id: quote.id,
      contact,
      passengers: people,
      payment_option: option,
      special_requests: requests,
      consent: { accepted: consent, policy_version: "booking-v1" },
    };
    const serialized = JSON.stringify(business);
    if (payload.current !== serialized) {
      payload.current = serialized;
      key.current = crypto.randomUUID();
    }
    try {
      setBooking(
        await api<Booking>("/bookings", {
          method: "POST",
          headers: { "X-CSRF-Token": csrf, "Idempotency-Key": key.current },
          body: JSON.stringify({ ...business, turnstile_token: captcha }),
        }),
      );
    } catch (e) {
      const code = e instanceof ApiError ? e.code : "";
      const messages: Record<string, string> = {
        QUOTE_EXPIRED:
          "Báo giá hết hạn. Quay lại bước 1 để lấy giá mới; thông tin hành khách vẫn được giữ.",
        PRICE_CHANGED:
          "Giá đã được cập nhật. Quay lại bước 1 để xem và đồng ý báo giá mới.",
        SOLD_OUT: "Lịch này không còn đủ chỗ. Vui lòng chọn lịch khác.",
        PASSENGER_MISMATCH:
          "Ngày sinh hoặc nhóm tuổi chưa khớp giá theo ngày khởi hành. Kiểm tra từng hành khách.",
      };
      setError(
        messages[code] ||
          (e instanceof Error ? e.message : "Chưa tạo được đơn."),
      );
      setAttempt((n) => n + 1);
    } finally {
      setBusy(false);
    }
  }
  if (booking) return <BookingResult initial={booking} />;
  return (
    <>
      <ol className="checkout-steps" aria-label="Các bước đặt tour">
        {["Chọn số khách", "Thông tin hành khách", "Kiểm tra & giữ chỗ"].map(
          (s, i) => (
            <li key={s} aria-current={step === i + 1 ? "step" : undefined}>
              {i + 1}. {s}
            </li>
          ),
        )}
      </ol>
      <p className="notice">
        Bản demo — không chuyển tiền thật. Báo giá có hiệu lực 5 phút và chưa
        giữ chỗ.
      </p>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {step === 1 && (
        <form
          className="form-grid"
          onSubmit={(e) => {
            e.preventDefault();
            void getQuote();
          }}
        >
          <h2>{title}</h2>
          {prices.map((p) => (
            <label key={p.passenger_type}>
              {labels[p.passenger_type]} · {p.min_age} đến{" "}
              {p.max_age_exclusive == null
                ? "không giới hạn"
                : p.max_age_exclusive - 1}{" "}
              tuổi · {money(p.unit_price)}
              <input
                type="number"
                min={p.passenger_type === "adult" ? 1 : 0}
                max="20"
                step="1"
                required
                value={counts[p.passenger_type as keyof Counts]}
                onChange={(e) =>
                  setCounts({
                    ...counts,
                    [p.passenger_type]: Number(e.target.value),
                  })
                }
              />
            </label>
          ))}
          {surcharges.length > 0 && (
            <fieldset>
              <legend>Phụ thu và dịch vụ thêm</legend>
              {surcharges.map((extra) => (
                <label className="check" key={extra.code}>
                  <input
                    type="checkbox"
                    disabled={extra.required||busy}
                    checked={extra.required || codes.includes(extra.code)}
                    onChange={(e) =>
                      setCodes(
                        e.target.checked
                          ? [...codes, extra.code]
                          : codes.filter((c) => c !== extra.code),
                      )
                    }
                  />
                  {extra.label} · {money(extra.amount)} /{" "}
                  {extra.unit === "per_person" ? "người" : "đơn"}{" "}
                  {extra.required ? "(bắt buộc)" : ""}
                </label>
              ))}
            </fieldset>
          )}
          <button className="btn" disabled={busy}>
            {busy ? "Đang tính giá…" : "Xem báo giá"}
          </button>
        </form>
      )}
      {step === 2 && quote && (
        <form
          className="form-grid"
          onSubmit={(e) => {
            e.preventDefault();
            setStep(3);
          }}
        >
          <h2>Thông tin liên hệ</h2>
          <label>
            Họ tên người liên hệ
            <input
              required
              minLength={2}
              maxLength={120}
              autoComplete="name"
              value={contact.name}
              onChange={(e) => setContact({ ...contact, name: e.target.value })}
            />
          </label>
          <div className="form-row">
            <label>
              Email
              <input
                required
                type="email"
                maxLength={254}
                autoComplete="email"
                value={contact.email}
                onChange={(e) =>
                  setContact({ ...contact, email: e.target.value })
                }
              />
            </label>
            <label>
              Số điện thoại
              <input
                required
                type="tel"
                autoComplete="tel"
                maxLength={20}
                value={contact.phone}
                onChange={(e) =>
                  setContact({ ...contact, phone: e.target.value })
                }
              />
            </label>
          </div>
          <h2>Danh sách hành khách</h2>
          <p>
            Tuổi được tính tại ngày khởi hành; không thu CCCD hoặc hộ chiếu
            trong demo.
          </p>
          {people.map((p, i) => (
            <fieldset key={i}>
              <legend>
                {labels[p.passenger_type]} {i + 1}
              </legend>
              <label>
                Họ tên hành khách {i + 1}
                <input
                  required
                  minLength={2}
                  maxLength={120}
                  value={p.full_name}
                  onChange={(e) =>
                    setPeople(
                      people.map((v, j) =>
                        j === i ? { ...v, full_name: e.target.value } : v,
                      ),
                    )
                  }
                />
              </label>
              <label>
                Ngày sinh hành khách {i + 1}
                <input
                  required
                  type="date"
                  value={p.date_of_birth}
                  onChange={(e) =>
                    setPeople(
                      people.map((v, j) =>
                        j === i ? { ...v, date_of_birth: e.target.value } : v,
                      ),
                    )
                  }
                />
              </label>
            </fieldset>
          ))}
          <label>
            Yêu cầu thêm
            <textarea
              maxLength={5000}
              value={requests}
              onChange={(e) => setRequests(e.target.value)}
            />
          </label>
          <div className="tools">
            <button type="button" onClick={() => setStep(1)}>
              Quay lại số khách
            </button>
            <button className="btn">Kiểm tra đơn</button>
          </div>
        </form>
      )}
      {step === 3 && quote && (
        <div className="form-grid">
          <h2>Kiểm tra trước khi giữ chỗ</h2>
          <p>
            {contact.name} · {contact.email} · {contact.phone}
          </p>
          <p>
            {people.length} hành khách · {quote.seat_count} chỗ
          </p>
          <div className="summary">
            <dl>
              {quote.price_breakdown.items.map((item) => (
                <div className="wide" key={item.code}>
                  <dt>
                    {labels[item.code] || item.label} × {item.quantity}
                  </dt>
                  <dd>{money(item.total_amount)}</dd>
                </div>
              ))}
              <dt>Tổng tiền</dt>
              <dd>{money(quote.price_breakdown.total_amount)}</dd>
              <dt>Đặt cọc</dt>
              <dd>{money(quote.price_breakdown.deposit_amount)}</dd>
              <dt>Dư sau khi cọc</dt>
              <dd>{money(quote.price_breakdown.balance_after_deposit)}</dd>
            </dl>
            <p>
              Hạn thanh toán dư:{" "}
              {new Date(quote.balance_due_at).toLocaleString("vi-VN", {
                timeZone: "Asia/Ho_Chi_Minh",
              })}
            </p>
          </div>
          <label>
            Phương án thanh toán
            <select value={option} onChange={(e) => setOption(e.target.value)}>
              {quote.allowed_payment_options.map((o) => (
                <option key={o} value={o}>
                  {o === "full" ? "Thanh toán toàn bộ" : "Đặt cọc trước"}
                </option>
              ))}
            </select>
          </label>
          <label>
            Mã giảm giá
            <input disabled placeholder="Chưa hỗ trợ mã giảm giá trong demo" />
          </label>
          <h3>Chính sách huỷ</h3>
          <div
            className="prose"
            dangerouslySetInnerHTML={{
              __html: quote.price_breakdown.cancellation_policy,
            }}
          />
          <label className="check">
            <input
              type="checkbox"
              checked={consent}
              onChange={(e) => setConsent(e.target.checked)}
            />
            Tôi đồng ý điều khoản và chính sách đặt tour (booking-v1), cho phép
            xử lý thông tin để phục vụ đơn này.
          </label>
          <Captcha
            key={attempt}
            action="checkout"
            onToken={onToken}
            mock={mock}
            siteKey={siteKey}
          />
          <p>
            Sau khi tạo đơn thành công, chỗ được giữ {quote.hold_minutes} phút
            theo đồng hồ máy chủ.
          </p>
          <div className="tools">
            <button disabled={busy} onClick={() => setStep(2)}>
              Sửa thông tin
            </button>
            <button disabled={busy} onClick={() => setStep(1)}>
              Lấy báo giá mới
            </button>
            <button
              className="btn"
              disabled={busy || !consent || !captcha}
              onClick={() => void create()}
            >
              {busy ? "Đang giữ chỗ…" : "Tạo đơn giữ chỗ"}
            </button>
          </div>
        </div>
      )}
      <p>
        <Link className="text-action" href="/contact">
          Cần hỗ trợ? Xem thông tin liên hệ
        </Link>
      </p>
    </>
  );
}
