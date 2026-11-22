import { useEffect, useRef, useState, type FormEvent } from "react";
import { authorized } from "./lib/session";

type Customer = {
  id: string; full_name: string; email: string | null; phone: string | null;
  company_name: string | null; notes: string | null;
};
type History = { id: string; code: string; status: string; tour_title: string; created_at: string; total_amount: string };
type BookingMatch = {
  id: string; code: string; version: number; tour_snapshot: { title: string };
  contact: { name: string; email: string; phone: string };
};
const path = "/admin/customers";
const message = (error: unknown) => error instanceof Error ? error.message : "Không thể kết nối. Vui lòng thử lại.";
const date = new Intl.DateTimeFormat("vi-VN", { timeZone: "Asia/Ho_Chi_Minh", dateStyle: "short" });
const vnd = new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" });
const bookingStatus: Record<string, string> = {
  pending: "Chờ thanh toán", confirmed: "Đã xác nhận", expired: "Đã hết hạn",
  cancelled: "Đã hủy", completed: "Đã hoàn thành",
};

export default function Customers({ canReadBookings, canLink }: { canReadBookings: boolean; canLink: boolean }) {
  const [rows, setRows] = useState<Customer[] | null>(null);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    let active = true;
    authorized<Customer[]>(path + "?q=" + encodeURIComponent(search) + "&page=" + page).then(
      (data) => { if (active) setRows(data); },
      (reason) => { if (active) { setRows([]); setError(message(reason)); } },
    );
    return () => { active = false; };
  }, [search, page, revision]);

  function reload() {
    setRows(null);
    setError("");
    setRevision((n) => n + 1);
  }

  return (
    <section className="workspace" aria-labelledby="customers-heading">
      <h2 id="customers-heading">Khách hàng</h2>
      <form className="toolbar" onSubmit={(event) => {
        event.preventDefault();
        setSearch(String(new FormData(event.currentTarget).get("q") || ""));
        setPage(1);
        reload();
      }}>
        <label>Tên, email, số điện thoại hoặc công ty<input name="q" maxLength={120} /></label>
        <button>Tìm khách hàng</button>
        <button type="button" onClick={() => setSelected("new")}>Thêm khách hàng</button>
      </form>
      {error && <p role="alert">{error} <button onClick={reload}>Thử lại</button></p>}
      {rows === null ? <p role="status">Đang tải khách hàng…</p> : (
        <div className="data-list">
          {rows.map((customer) => (
            <button key={customer.id} className="list-row" aria-pressed={selected === customer.id} onClick={() => setSelected(customer.id)}>
              <strong>{customer.full_name}</strong>
              <span>{[customer.email, customer.phone, customer.company_name].filter(Boolean).join(" · ")}</span>
            </button>
          ))}
          {rows.length === 0 && !error && <p>Chưa có khách hàng phù hợp.</p>}
        </div>
      )}
      <div className="toolbar">
        <button disabled={page === 1} onClick={() => { setPage(page - 1); reload(); }}>Trang trước</button>
        <span>Trang {page}</span>
        <button disabled={rows?.length !== 20} onClick={() => { setPage(page + 1); reload(); }}>Trang sau</button>
      </div>
      {selected && <CustomerEditor key={selected} id={selected} canReadBookings={canReadBookings} canLink={canLink}
        onSaved={(customer) => { setSelected(customer.id); reload(); }} />}
    </section>
  );
}

function CustomerEditor({ id, canReadBookings, canLink, onSaved }: {
  id: string; canReadBookings: boolean; canLink: boolean; onSaved: (customer: Customer) => void;
}) {
  const creating = id === "new";
  const [customer, setCustomer] = useState<Customer | null>(null);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const [retry, setRetry] = useState(0);
  const saving = useRef(false);
  useEffect(() => {
    if (creating) return;
    let active = true;
    authorized<Customer>(path + "/" + id).then(
      (data) => { if (active) setCustomer(data); },
      (reason) => { if (active) setError(message(reason)); },
    );
    return () => { active = false; };
  }, [id, creating, retry]);

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (saving.current) return;
    const fields = new FormData(event.currentTarget);
    const payload = Object.fromEntries([...fields].map(([k, v]) => [k, String(v).trim() || null]));
    setError("");
    setNotice("");
    if (!payload.email && !payload.phone) {
      setError("Nhập ít nhất email hoặc số điện thoại.");
      return;
    }
    saving.current = true;
    setBusy(true);
    try {
      const saved = await authorized<Customer>(creating ? path : path + "/" + id, {
        method: creating ? "POST" : "PATCH", body: JSON.stringify(payload),
      });
      setCustomer(saved);
      setNotice("Đã lưu thông tin khách hàng.");
      onSaved(saved);
    } catch (reason) {
      setError(message(reason));
    } finally {
      saving.current = false;
      setBusy(false);
    }
  }

  if (!creating && !customer) {
    return error
      ? <><p role="alert">{error}</p><button onClick={() => { setError(""); setRetry((n) => n + 1); }}>Thử tải lại</button></>
      : <p role="status">Đang tải hồ sơ…</p>;
  }
  return (
    <article className="editor">
      <h3>{creating ? "Hồ sơ khách mới" : customer?.full_name}</h3>
      <form onSubmit={save}>
        <fieldset disabled={busy}>
          <legend>Thông tin liên hệ</legend>
          <label>Họ và tên<input name="full_name" required minLength={2} maxLength={120} defaultValue={customer?.full_name} /></label>
          <label>Email<input name="email" type="email" maxLength={254} defaultValue={customer?.email || ""} /></label>
          <label>Số điện thoại<input name="phone" type="tel" maxLength={24} defaultValue={customer?.phone || ""} /></label>
          <label>Công ty<input name="company_name" maxLength={200} defaultValue={customer?.company_name || ""} /></label>
          <label>Ghi chú nội bộ<textarea name="notes" maxLength={5000} defaultValue={customer?.notes || ""} /></label>
          <button>Lưu khách hàng</button>
        </fieldset>
      </form>
      {error && <p role="alert">{error}</p>}
      {notice && <p role="status">{notice}</p>}
      {!creating && canReadBookings && <CustomerHistory id={id} canLink={canLink} />}
    </article>
  );
}

function CustomerHistory({ id, canLink }: { id: string; canLink: boolean }) {
  const [rows, setRows] = useState<History[] | null>(null);
  const [page, setPage] = useState(1);
  const [revision, setRevision] = useState(0);
  const [error, setError] = useState("");
  const [linkError, setLinkError] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const [booking, setBooking] = useState<BookingMatch | null>(null);
  const running = useRef(false);
  useEffect(() => {
    let active = true;
    authorized<History[]>(path + "/" + id + "/bookings?page=" + page).then(
      (data) => { if (active) setRows(data); },
      (reason) => { if (active) { setRows([]); setError(message(reason)); } },
    );
    return () => { active = false; };
  }, [id, page, revision]);

  function reload() {
    setRows(null);
    setError("");
    setRevision((n) => n + 1);
  }

  async function linkOrFind(event: FormEvent<HTMLFormElement>, link: boolean) {
    event.preventDefault();
    if (running.current) return;
    const fields = new FormData(event.currentTarget);
    running.current = true;
    setBusy(true);
    setLinkError("");
    setNotice("");
    try {
      if (link && booking && fields.has("verified")) {
        await authorized("/admin/bookings/" + booking.id, {
          method: "PATCH",
          body: JSON.stringify({ expected_version: booking.version, customer_id: id, customer_link_verified: true }),
        });
        setBooking(null);
        setPage(1);
        reload();
        setNotice("Đã xác minh và gắn đơn vào hồ sơ.");
      } else if (!link) {
        setBooking(null);
        const code = String(fields.get("code") || "").trim().toUpperCase();
        const matches = await authorized<BookingMatch[]>("/admin/bookings?page_size=100&q=" + encodeURIComponent(code));
        const match = matches.find((item) => item.code === code);
        if (!match) throw new Error("Không tìm thấy đơn có mã chính xác này.");
        setBooking(match);
      }
    } catch (reason) {
      setLinkError(message(reason));
    } finally {
      running.current = false;
      setBusy(false);
    }
  }

  return (
    <>
      <h4>Lịch sử đơn đã xác minh</h4>
      {error && <p role="alert">{error} <button onClick={reload}>Thử lại</button></p>}
      {rows === null ? <p role="status">Đang tải lịch sử…</p> : rows.length > 0 ? (
        <ul>{rows.map((row) => <li key={row.id}><strong>{row.code}</strong> · {row.tour_title} · {bookingStatus[row.status]} · {date.format(new Date(row.created_at))} · {vnd.format(Number(row.total_amount))}</li>)}</ul>
      ) : !error && <p>Chưa có đơn đã xác minh.</p>}
      <div className="toolbar">
        <button disabled={page === 1} onClick={() => { setPage(page - 1); reload(); }}>Trang lịch sử trước</button>
        <span>Trang {page}</span>
        <button disabled={rows?.length !== 20} onClick={() => { setPage(page + 1); reload(); }}>Trang lịch sử sau</button>
      </div>
      {canLink && (
        <>
          <form onSubmit={(event) => void linkOrFind(event, false)}>
            <label>Mã đơn cần xác minh<input name="code" required maxLength={24} onChange={() => setBooking(null)} disabled={busy} /></label>
            <button disabled={busy}>Kiểm tra mã đơn</button>
          </form>
          {booking && (
            <form key={booking.id + booking.version} onSubmit={(event) => void linkOrFind(event, true)}>
              <fieldset disabled={busy}>
                <legend>{booking.code} · {booking.tour_snapshot.title}</legend>
                <p>{booking.contact.name} · {booking.contact.email} · {booking.contact.phone}</p>
                <label>
                  <input name="verified" type="checkbox" required />
                  Tôi đã xác minh chủ đơn và được phép gắn đơn vào hồ sơ này.
                </label>
                <button>Gắn đơn đã xác minh</button>
              </fieldset>
            </form>
          )}
          {linkError && <p role="alert">{linkError}</p>}
          {notice && <p role="status">{notice}</p>}
        </>
      )}
    </>
  );
}
