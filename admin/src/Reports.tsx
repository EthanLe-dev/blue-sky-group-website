import { useEffect, useState, type FormEvent } from "react";
import { authorized } from "./lib/session";

type Movement = {
  period: string;
  cash_received: string;
  gross_booking_amount: string;
  refund_amount: string;
  net_received: string;
  booking_count: number;
};
type Revenue = Movement & { series: Movement[]; is_demo: boolean };
type Occupancy = {
  departure_id: string; tour_title: string; starts_at: string;
  capacity: number; confirmed_seats: number; held_seats: number; occupancy_percent: number;
};
type Quotes = { request_count: number; won_count: number; won_rate_percent: number; by_status: Record<string, number> };
const zone = "Asia/Ho_Chi_Minh";
const date = (value: Date) => value.toLocaleDateString("sv-SE", { timeZone: zone });
const today = () => date(new Date());
const money = (amount: string) => BigInt(amount).toLocaleString("vi-VN") + " ₫";
const counts: Record<string, string> = { new: "Mới", preparing: "Đang chuẩn bị", quoted: "Đã báo giá", won: "Đã chốt", lost: "Không chốt" };

export default function Reports() {
  const [result, setResult] = useState<{ revenue: Revenue; occupancy: Occupancy[]; quotes: Quotes } | null>(null);
  const [tours, setTours] = useState<{ id: string; title: string }[]>([]);
  const [filters, setFilters] = useState(() => new URLSearchParams({ from: today().slice(0, 7) + "-01", to: today() }));
  const [tourError, setTourError] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    authorized<typeof tours>("/admin/reports/tours", { signal: controller.signal })
      .then(setTours)
      .catch(() => { if (!controller.signal.aborted) setTourError(true); });
    return () => controller.abort();
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    const get = <T,>(name: string, query: URLSearchParams) =>
      authorized<T>("/admin/reports/" + name + "?" + query, { signal: controller.signal });
    const subset = (...keys: string[]) => new URLSearchParams([...filters].filter(([key]) => keys.includes(key)));
    Promise.all([
      get<Revenue>("revenue", subset("from", "to", "tour_id", "source", "group_by")),
      get<Occupancy[]>("occupancy", subset("from", "to", "tour_id")),
      get<Quotes>("group-quotes", subset("from", "to", "group_type")),
    ]).then(([revenue, occupancy, quotes]) => {
      if (!controller.signal.aborted) setResult({ revenue, occupancy, quotes });
    }).catch((e: Error) => {
      if (!controller.signal.aborted) setError(e.message || "Chưa tải được báo cáo.");
    });
    return () => controller.abort();
  }, [filters]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const next = new URLSearchParams();
    for (const [key, value] of new FormData(event.currentTarget)) {
      if (typeof value === "string" && value) next.set(key, value);
    }
    setResult(null);
    setError("");
    setFilters(next);
  }

  return (
    <section className="workspace">
      <h2>Báo cáo kinh doanh</h2>
      <p>Múi giờ Việt Nam. Số thu và hoàn là dữ liệu giả lập.</p>
      <form className="editor-grid" onSubmit={submit}>
        <label>Từ ngày<input type="date" name="from" required defaultValue={filters.get("from") || ""} /></label>
        <label>Đến ngày<input type="date" name="to" required defaultValue={filters.get("to") || ""} /></label>
        <label>Tour<select name="tour_id"><option value="">Tất cả tour</option>{tours.map(t => <option key={t.id} value={t.id}>{t.title}</option>)}</select></label>
        <label>Nguồn đơn<select name="source"><option value="">Tất cả</option><option value="web">Website</option><option value="phone">Điện thoại</option><option value="counter">Tại quầy</option></select></label>
        <label>Kỳ doanh thu<select name="group_by"><option value="day">Ngày</option><option value="month">Tháng</option></select></label>
        <label>Loại đoàn<select name="group_type"><option value="">Tất cả</option><option value="school">Trường học</option><option value="company">Doanh nghiệp</option></select></label>
        <button type="submit">Xem báo cáo</button>
      </form>
      {tourError && <p role="alert">Chưa tải được danh sách tour. Vẫn có thể xem số liệu tổng hợp.</p>}
      {error && <p role="alert">{error}</p>}
      {!error && !result && <p role="status">Đang tải báo cáo…</p>}
      {result && (
        <>
          <h3>Doanh thu</h3>
          <p>Giá trị đơn {money(result.revenue.gross_booking_amount)} · {result.revenue.booking_count} đơn</p>
          <p>Đã thu {money(result.revenue.cash_received)} · Đã hoàn {money(result.revenue.refund_amount)} · Thu ròng {money(result.revenue.net_received)}</p>
          {result.revenue.series.length === 0 ? <p>Chưa có giao dịch trong kỳ.</p> : (
            <div className="report-table" tabIndex={0} role="region" aria-label="Bảng doanh thu">
              <table><thead><tr><th>Kỳ</th><th>Đơn</th><th>Giá trị đơn</th><th>Đã thu</th><th>Đã hoàn</th><th>Thu ròng</th></tr></thead>
                <tbody>{result.revenue.series.map(r => <tr key={r.period}><th scope="row">{r.period}</th><td>{r.booking_count}</td><td>{money(r.gross_booking_amount)}</td><td>{money(r.cash_received)}</td><td>{money(r.refund_amount)}</td><td>{money(r.net_received)}</td></tr>)}</tbody>
              </table>
            </div>
          )}
          <h3>Lấp đầy theo lịch khởi hành</h3>
          {result.occupancy.length === 0 ? <p>Chưa có lịch khởi hành trong kỳ.</p> : (
            <div className="report-table" tabIndex={0} role="region" aria-label="Bảng lấp đầy">
              <table><thead><tr><th>Tour</th><th>Ngày đi</th><th>Sức chứa</th><th>Đang giữ</th><th>Đã xác nhận</th><th>Lấp đầy</th></tr></thead>
                <tbody>{result.occupancy.map(r => <tr key={r.departure_id}><th scope="row">{r.tour_title}</th><td>{date(new Date(r.starts_at))}</td><td>{r.capacity}</td><td>{r.held_seats}</td><td>{r.confirmed_seats}</td><td>{r.occupancy_percent.toLocaleString("vi-VN")}%</td></tr>)}</tbody>
              </table>
            </div>
          )}
          <h3>Tỷ lệ chốt yêu cầu đoàn</h3>
          <p>{result.quotes.won_count}/{result.quotes.request_count} yêu cầu đã chốt ({result.quotes.won_rate_percent.toLocaleString("vi-VN")}%)</p>
          {result.quotes.request_count === 0 ? <p>Chưa có yêu cầu đoàn trong kỳ.</p> :
            <ul>{Object.entries(counts).map(([key, label]) => <li key={key}>{label}: {result.quotes.by_status[key]}</li>)}</ul>}
        </>
      )}
    </section>
  );
}
