import { useEffect, useState, useRef } from "react";
import BookingPayments from "./BookingPayments";
import { authorized } from "./lib/session";
type Row = Record<string, unknown>;
type Contact = { name: string; email: string; phone: string };
type Person = {
  passenger_type: string;
  full_name: string;
  date_of_birth: string;
};
type Booking = {
  id: string;
  code: string;
  status: string;
  version: number;
  contact: Contact;
  passengers: Person[];
  special_requests: string;
  source: string;
  total_amount: string;
  deposit_amount: string;
  paid_amount: string;
  refunded_amount: string;
  payment_option: string;
  balance_amount: string;
  expires_at: string;
  tour_snapshot: { title: string; starts_at: string };
  price_snapshot: { cancellation_policy: string };
  notes: { id: string; body: string; created_at: string }[];
  events: { id: string; event_type: string; created_at: string }[];
};
type Quote = {
  id: string;
  allowed_payment_options: string[];
  price_breakdown: {
    total_amount: string;
    deposit_amount: string;
    balance_after_deposit: string;
  };
};
const money = (v: string) => Number(v).toLocaleString("vi-VN") + " đ";
export default function Bookings({ write, canPay, canReadNotifications }: { write: boolean; canPay: boolean; canReadNotifications: boolean }) {
  const [rows, setRows] = useState<Booking[]>([]);
  const [record, setRecord] = useState<Booking | null>(null);
  const [message, setMessage] = useState("");
  const [status, setStatus] = useState("");
  const [q, setQ] = useState("");
  const [page, setPage] = useState(1);
  const [manual, setManual] = useState(false);
  async function load() {
    try {
      setRows(
        await authorized<Booking[]>(
          "/admin/bookings?page=" +
            page +
            "&status=" +
            status +
            "&q=" +
            encodeURIComponent(q),
        ),
      );
    } catch (e) {
      setMessage(String(e));
    }
  }
  useEffect(() => {
    let active = true;
    authorized<Booking[]>(
      "/admin/bookings?page=" +
        page +
        "&status=" +
        status +
        "&q=" +
        encodeURIComponent(q),
    )
      .then((rows) => {
        if (active) setRows(rows);
      })
      .catch((e) => {
        if (active) setMessage(String(e));
      });
    return () => {
      active = false;
    };
  }, [page, status, q]);
  async function detail(id: string) {
    try {
      setRecord(await authorized<Booking>("/admin/bookings/" + id));
      setManual(false);
    } catch (e) {
      setMessage(String(e));
    }
  }
  return (
    <section className="workspace">
      <h2>Quản lý đơn tour</h2>
      <div className="toolbar">
        <label>
          Trạng thái
          <select
            value={status}
            onChange={(e) => {
              setPage(1);
              setStatus(e.target.value);
            }}
          >
            <option value="">Tất cả</option>
            {["pending", "confirmed", "expired", "cancelled", "completed"].map(
              (s) => (
                <option key={s}>{s}</option>
              ),
            )}
          </select>
        </label>
        <label>
          Mã đơn / tên liên hệ
          <input value={q} onChange={(e) => setQ(e.target.value)} />
        </label>
        <button onClick={() => void load()}>Tìm đơn</button>
        {write && (
          <button
            onClick={() => {
              setManual(true);
              setRecord(null);
            }}
          >
            Tạo đơn thủ công
          </button>
        )}
      </div>
      <p role="status">{message}</p>
      {rows.length ? (
        rows.map((b) => (
          <button
            className="list-row"
            key={b.id}
            onClick={() => void detail(b.id)}
          >
            <strong>
              {b.code} · {b.tour_snapshot.title}
            </strong>
            <span>
              {b.status} · {b.contact.name} · {money(b.total_amount)} ·{" "}
              {b.source}
            </span>
          </button>
        ))
      ) : (
        <p>Chưa có đơn phù hợp.</p>
      )}
      <div className="toolbar">
        <button disabled={page === 1} onClick={() => setPage(page - 1)}>
          Trang trước
        </button>
        <span>Trang {page}</span>
        <button disabled={rows.length < 20} onClick={() => setPage(page + 1)}>
          Trang tiếp
        </button>
      </div>
      {manual && (
        <Manual
          onSaved={async (id) => {
            await detail(id);
            await load();
          }}
        />
      )}
      {record && (
        <div className="editor">
          <h3>Đơn {record.code}</h3>
          <p>
            {record.tour_snapshot.title} ·{" "}
            {new Date(record.tour_snapshot.starts_at).toLocaleString("vi-VN")}
          </p>
          <p>
            {record.status} · Phiên bản {record.version} · Hạn giữ chỗ{" "}
            {new Date(record.expires_at).toLocaleString("vi-VN")}
          </p>
          <p>
            Tổng {money(record.total_amount)} · Cọc{" "}
            {money(record.deposit_amount)} · Đã thu {money(record.paid_amount)}{" "}
            · Còn nợ {money(record.balance_amount)}
          </p>
          <p>Giả lập — không chuyển tiền thật. Thanh toán ở M4.</p>
          <form
            onSubmit={async (e) => {
              e.preventDefault();
              try {
                await authorized("/admin/bookings/" + record.id, {
                  method: "PATCH",
                  body: JSON.stringify({
                    expected_version: record.version,
                    contact: record.contact,
                    passengers: record.passengers,
                    special_requests: record.special_requests,
                  }),
                });
                await detail(record.id);
                await load();
                setMessage("Đã lưu đơn.");
              } catch (e) {
                setMessage(String(e));
              }
            }}
          >
            <fieldset
              disabled={
                !write || !["pending", "confirmed"].includes(record.status)
              }
            >
              <legend>Thông tin đơn</legend>
              <ContactFields
                value={record.contact}
                onChange={(contact) => setRecord({ ...record, contact })}
              />
              <People
                value={record.passengers}
                onChange={(passengers) => setRecord({ ...record, passengers })}
              />
              <label>
                Yêu cầu thêm
                <textarea
                  value={record.special_requests}
                  maxLength={5000}
                  onChange={(e) =>
                    setRecord({ ...record, special_requests: e.target.value })
                  }
                />
              </label>
              <button>Lưu chỉnh sửa đơn</button>
            </fieldset>
          </form>
          <details>
            <summary>Chính sách đã chốt trong đơn</summary>
            <div
              dangerouslySetInnerHTML={{
                __html: record.price_snapshot.cancellation_policy,
              }}
            />
          </details>
          {write && (
            <div className="toolbar">
              <button
                disabled={!["pending", "confirmed"].includes(record.status)}
                onClick={async () => {
                  const reason = prompt("Lý do huỷ đơn (không tự hoàn tiền)");
                  if (!reason) return;
                  try {
                    await authorized(
                      "/admin/bookings/" + record.id + "/cancel",
                      {
                        method: "POST",
                        headers: { "Idempotency-Key": crypto.randomUUID() },
                        body: JSON.stringify({
                          expected_version: record.version,
                          reason,
                        }),
                      },
                    );
                    await detail(record.id);
                    await load();
                    setMessage(
                      "Đã huỷ đơn. Tiền đã thu giữ nguyên; hoàn tiền xử lý riêng.",
                    );
                  } catch (e) {
                    setMessage(String(e));
                  }
                }}
              >
                Huỷ đơn và trả chỗ
              </button>
              <button
                disabled={record.status !== "confirmed"}
                onClick={async () => {
                  try {
                    await authorized(
                      "/admin/bookings/" + record.id + "/complete",
                      {
                        method: "POST",
                        body: JSON.stringify({
                          expected_version: record.version,
                        }),
                      },
                    );
                    await detail(record.id);
                    await load();
                  } catch (e) {
                    setMessage(String(e));
                  }
                }}
              >
                Đánh dấu hoàn thành chuyến
              </button>
            </div>
          )}
          <BookingPayments key={record.id} booking={record} canPay={canPay}
            canReadNotifications={canReadNotifications} onChanged={async () => { await detail(record.id); await load(); }} />
          <h3>Ghi chú nội bộ</h3>
          {record.notes.map((n) => (
            <p key={n.id}>
              {n.body} · {new Date(n.created_at).toLocaleString("vi-VN")}
            </p>
          ))}
          {write && (
            <form
              onSubmit={async (e) => {
                e.preventDefault();
                const form = e.currentTarget;
                try {
                  await authorized("/admin/bookings/" + record.id + "/notes", {
                    method: "POST",
                    body: JSON.stringify({
                      body: new FormData(form).get("note"),
                    }),
                  });
                  form.reset();
                  await detail(record.id);
                } catch (e) {
                  setMessage(String(e));
                }
              }}
            >
              <label>
                Nội dung ghi chú
                <textarea name="note" required maxLength={5000} />
              </label>
              <button>Thêm ghi chú</button>
            </form>
          )}
          <h3>Lịch sử trạng thái</h3>
          <ul>
            {record.events.map((e) => (
              <li key={e.id}>
                {e.event_type} ·{" "}
                {new Date(e.created_at).toLocaleString("vi-VN")}
              </li>
            ))}
          </ul>
          <button onClick={() => void detail(record.id)}>Tải lại đơn</button>
        </div>
      )}
    </section>
  );
}
function ContactFields({
  value,
  onChange,
}: {
  value: Contact;
  onChange: (c: Contact) => void;
}) {
  return (
    <div className="editor-grid">
      {(["name", "email", "phone"] as const).map((key) => (
        <label key={key}>
          {
            {
              name: "Họ tên liên hệ",
              email: "Email liên hệ",
              phone: "Số điện thoại liên hệ",
            }[key]
          }
          <input
            type={key === "email" ? "email" : key === "phone" ? "tel" : "text"}
            required
            maxLength={key === "email" ? 254 : 120}
            value={value[key]}
            onChange={(e) => onChange({ ...value, [key]: e.target.value })}
          />
        </label>
      ))}
    </div>
  );
}
function People({
  value,
  onChange,
}: {
  value: Person[];
  onChange: (p: Person[]) => void;
}) {
  return (
    <>
      {value.map((p, i) => (
        <fieldset key={i}>
          <legend>
            Hành khách {i + 1} · {p.passenger_type}
          </legend>
          <div className="editor-grid">
            <label>
              Họ tên hành khách {i + 1}
              <input
                required
                minLength={2}
                maxLength={120}
                value={p.full_name}
                onChange={(e) =>
                  onChange(
                    value.map((v, j) =>
                      j === i ? { ...v, full_name: e.target.value } : v,
                    ),
                  )
                }
              />
            </label>
            <label>
              Ngày sinh hành khách {i + 1}
              <input
                type="date"
                required
                value={p.date_of_birth.slice(0, 10)}
                onChange={(e) =>
                  onChange(
                    value.map((v, j) =>
                      j === i ? { ...v, date_of_birth: e.target.value } : v,
                    ),
                  )
                }
              />
            </label>
          </div>
        </fieldset>
      ))}
    </>
  );
}
function Manual({ onSaved }: { onSaved: (id: string) => Promise<void> }) {
  const [tours, setTours] = useState<Row[]>([]);
  const [tour, setTour] = useState("");
  const [departures, setDepartures] = useState<Row[]>([]);
  const [departure, setDeparture] = useState("");
  const [counts, setCounts] = useState({ adult: 1, child: 0, infant: 0 });
  const [codes, setCodes] = useState<string[]>([]);
  const [quote, setQuote] = useState<Quote | null>(null);
  const [contact, setContact] = useState<Contact>({
    name: "",
    email: "",
    phone: "",
  });
  const [passengers, setPassengers] = useState<Person[]>([]);
  const [source, setSource] = useState("phone");
  const [option, setOption] = useState("full");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const key = useRef("");
  const payload = useRef("");
  useEffect(() => {
    authorized<Row[]>("/tours?page_size=100")
      .then(setTours)
      .catch((e) => setMessage(String(e)));
  }, []);
  useEffect(() => {
    if (tour)
      authorized<Row[]>("/tours/" + tour + "/departures?page_size=100")
        .then((d) => {
          setDepartures(d);
          setDeparture(String(d[0]?.id || ""));
          setQuote(null);
        })
        .catch((e) => setMessage(String(e)));
  }, [tour]);
  return (
    <section className="editor">
      <h3>Tạo đơn thủ công</h3>
      <p>Dùng cùng kho chỗ và quy tắc giữ chỗ như khách đặt trên web.</p>
      <label>
        Tour
        <select
          value={tour}
          onChange={(e) => {
            setTour(e.target.value);
            setCodes([]);
          }}
        >
          <option value="">Chọn tour</option>
          {tours.map((t) => (
            <option key={String(t.id)} value={String(t.slug)}>
              {String(t.title)}
            </option>
          ))}
        </select>
      </label>
      <label>
        Ngày đi
        <select
          value={departure}
          onChange={(e) => {
            setDeparture(e.target.value);
            setCodes([]);
            setQuote(null);
          }}
        >
          {departures.map((d) => (
            <option key={String(d.id)} value={String(d.id)}>
              {new Date(String(d.starts_at)).toLocaleDateString("vi-VN")} ·{" "}
              {String(d.remaining_seats)} chỗ
            </option>
          ))}
        </select>
      </label>
      {Object.entries(counts).map(([type, n]) => (
        <label key={type}>
          {type}
          <input
            type="number"
            min={type === "adult" ? 1 : 0}
            max={20}
            value={n}
            onChange={(e) => {
              setCounts({ ...counts, [type]: Number(e.target.value) });
              setQuote(null);
            }}
          />
        </label>
      ))}
      {(
        (departures.find((d) => d.id === departure)?.surcharges || []) as {
          code: string;
          label: string;
          amount: string;
          unit: string;
          required: boolean;
        }[]
      ).map((extra) => (
        <label key={extra.code}>
          <input
            type="checkbox"
            disabled={extra.required}
            checked={extra.required || codes.includes(extra.code)}
            onChange={(e) => {
              setQuote(null);
              setCodes(
                e.target.checked
                  ? [...codes, extra.code]
                  : codes.filter((c) => c !== extra.code),
              );
            }}
          />
          {extra.label} · {money(extra.amount)} /{" "}
          {extra.unit === "per_person" ? "người" : "đơn"}{" "}
          {extra.required ? "(bắt buộc)" : ""}
        </label>
      ))}
      <button
        disabled={!departure || busy}
        onClick={async () => {
          setBusy(true);
          try {
            const q = await authorized<Quote>("/admin/booking-quotes", {
              method: "POST",
              body: JSON.stringify({
                departure_id: departure,
                counts,
                surcharge_codes: codes,
              }),
            });
            setQuote(q);
            setPassengers(
              Object.entries(counts).flatMap(([type, n]) =>
                Array.from({ length: n }, () => ({
                  passenger_type: type,
                  full_name: "",
                  date_of_birth: "",
                })),
              ),
            );
            setMessage("Đã lấy báo giá mới, có hiệu lực 5 phút.");
          } catch (e) {
            setMessage(String(e));
          } finally {
            setBusy(false);
          }
        }}
      >
        Lấy báo giá
      </button>
      {quote && (
        <form
          onSubmit={async (e) => {
            e.preventDefault();
            setBusy(true);
            const data = {
              quote_id: quote.id,
              contact,
              passengers,
              source,
              payment_option: option,
              consent: { accepted: true, policy_version: "booking-v1" },
              special_requests: "",
            };
            const text = JSON.stringify(data);
            if (payload.current !== text) {
              payload.current = text;
              key.current = crypto.randomUUID();
            }
            try {
              const b = await authorized<Booking>("/admin/bookings", {
                method: "POST",
                headers: { "Idempotency-Key": key.current },
                body: text,
              });
              await onSaved(b.id);
            } catch (e) {
              setMessage(String(e));
            } finally {
              setBusy(false);
            }
          }}
        >
          <p>
            Tổng {money(quote.price_breakdown.total_amount)} · Cọc{" "}
            {money(quote.price_breakdown.deposit_amount)}
          </p>
          <ContactFields value={contact} onChange={setContact} />
          <People value={passengers} onChange={setPassengers} />
          <label>
            Nguồn đơn
            <select value={source} onChange={(e) => setSource(e.target.value)}>
              <option value="phone">Điện thoại</option>
              <option value="counter">Tại quầy</option>
            </select>
          </label>
          <label>
            Thanh toán
            <select value={option} onChange={(e) => setOption(e.target.value)}>
              {quote.allowed_payment_options.map((o) => (
                <option key={o} value={o}>
                  {o === "full" ? "Toàn bộ" : "Đặt cọc"}
                </option>
              ))}
            </select>
          </label>
          <label>
            <input type="checkbox" required />
            Khách đã đồng ý điều khoản booking-v1 và xử lý dữ liệu cho đơn này.
          </label>
          <button disabled={busy}>
            {busy ? "Đang tạo đơn…" : "Tạo đơn thủ công và giữ chỗ"}
          </button>
        </form>
      )}
      <p role="status">{message}</p>
    </section>
  );
}
