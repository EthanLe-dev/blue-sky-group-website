import { useEffect, useState } from "react";
import { authorized } from "./lib/session";

export type NotificationJournal = {
  logs: { id: string; booking_code: string; channel: string; subject: string | null; body: string; created_at: string }[];
  jobs: { id: string; booking_code: string; template_code: string; status: string; attempts: number; error_code: string | null }[];
};

export function NotificationEntries({ journal }: { journal: NotificationJournal }) {
  return (
    <>
      {!journal.logs.length && !journal.jobs.length && <p>Chưa có bản xem trước trên trang này.</p>}
      {journal.jobs.map(e => (
        <p key={e.id}>{e.booking_code} · {e.status === "failed" ? "Tạo bản xem trước thất bại" : "Đang chờ tạo bản xem trước"} · {e.attempts}/5 lần</p>
      ))}
      {journal.logs.map(n => (
        <details key={n.id}>
          <summary>{n.channel === "email" ? "Email" : "SMS"} giả lập · {n.booking_code} · {new Date(n.created_at).toLocaleString("vi-VN")}</summary>
          {n.subject && <p><strong>{n.subject}</strong></p>}
          <p>{n.body}</p>
        </details>
      ))}
    </>
  );
}

export default function NotificationLog() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("");
  const [revision, setRevision] = useState(0);
  const [journal, setJournal] = useState<NotificationJournal | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    authorized<NotificationJournal>("/admin/notifications?" + new URLSearchParams({ page: String(page), status }), { signal: controller.signal })
      .then(setJournal)
      .catch(e => {
        if (!controller.signal.aborted) setError(e?.status === 404 ? "Môi trường này chưa bật thông báo giả lập." : String(e));
      });
    return () => controller.abort();
  }, [page, status, revision]);

  function changePage(next: number) {
    setError("");
    setJournal(null);
    setPage(next);
  }

  return (
    <section className="workspace">
      <h2>Nhật ký thông báo giả lập</h2>
      <div className="toolbar">
        <label>Trạng thái
          <select value={status} onChange={e => {
            setStatus(e.target.value);
            setRevision(n => n + 1);
            changePage(1);
          }}>
            <option value="">Tất cả</option>
            <option value="simulated">Đã tạo bản xem trước</option>
            <option value="pending">Đang chờ</option>
            <option value="processing">Đang xử lý</option>
            <option value="failed">Lỗi</option>
          </select>
        </label>
        <button onClick={() => {
          changePage(page);
          setRevision(n => n + 1);
        }}>Tải lại</button>
      </div>
      {error && <p role="alert">{error}</p>}
      {journal ? <NotificationEntries journal={journal} /> : !error && <p role="status">Đang tải nhật ký…</p>}
      <div className="toolbar">
        <button disabled={page === 1} onClick={() => changePage(page - 1)}>Trang trước</button>
        <span>Trang {page}</span>
        <button disabled={!journal || Math.max(journal.logs.length, journal.jobs.length) < 20} onClick={() => changePage(page + 1)}>Trang sau</button>
      </div>
    </section>
  );
}
