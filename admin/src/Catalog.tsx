import { useEffect, useState } from "react";
import { authorized } from "./lib/session";
type Row = Record<string, unknown>;
type Field = {
  key: string;
  label: string;
  type?: "number" | "check" | "text" | "long" | "json";
  options?: string[];
};
const f = (
  key: string,
  label: string,
  type: Field["type"] = "text",
  options?: string[],
): Field => ({ key, label, type, options });
const tourFields = [
  f("title", "Tên tour"),
  f("slug", "Đường dẫn (slug)"),
  f("code", "Mã tour"),
  f("summary", "Tóm tắt", "long"),
  f("description", "Giới thiệu (HTML an toàn)", "long"),
  f("region", "Khu vực", "text", ["domestic", "international"]),
  f("audience", "Đối tượng", "text", ["general", "school", "company"]),
  f("duration_days", "Số ngày", "number"),
  f("duration_nights", "Số đêm", "number"),
  f("origin_name", "Điểm khởi hành"),
  f("transport", "Phương tiện"),
  f("difficulty", "Mức độ", "text", ["easy", "moderate", "hard"]),
  f("hotel_summary", "Lưu trú"),
  f("included", "Giá bao gồm", "long"),
  f("excluded", "Giá chưa bao gồm", "long"),
  f("cancellation_policy", "Chính sách huỷ", "long"),
  f("preparation_guide", "Chuẩn bị trước chuyến đi", "long"),
  f("cover_media_id", "ID ảnh bìa"),
  f("poster_media_id", "ID poster"),
  f("media_ids", "Danh sách ID media (JSON)", "json"),
  f("destination_ids", "ID điểm đến (JSON)", "json"),
  f("category_ids", "ID danh mục (JSON)", "json"),
  f(
    "days",
    "Lịch trình (JSON: day_number, title, body, meals, accommodation)",
    "json",
  ),
  f(
    "stops",
    "Điểm dừng (JSON: day_number, sort_order, name, latitude, longitude, description)",
    "json",
  ),
  f("featured", "Tour nổi bật", "check"),
  f("is_demo", "Nội dung demo", "check"),
  f("seo_title", "Tiêu đề SEO"),
  f("seo_description", "Mô tả SEO"),
];
const resources: Record<
  string,
  { label: string; fields: Field[]; defaults: Row }
> = {
  tours: {
    label: "Tour",
    fields: tourFields,
    defaults: {
      region: "domestic",
      audience: "general",
      difficulty: "easy",
      duration_days: 2,
      duration_nights: 1,
      origin_name: "TP. Hồ Chí Minh",
      transport: "Ô tô",
      is_demo: true,
      days: [],
      stops: [],
      destination_ids: [],
      category_ids: [],
      media_ids: [],
    },
  },
  destinations: {
    label: "Điểm đến",
    fields: [
      f("name", "Tên"),
      f("slug", "Slug"),
      f("description", "Giới thiệu", "long"),
      f("country_code", "Mã quốc gia"),
      f("latitude", "Vĩ độ", "number"),
      f("longitude", "Kinh độ", "number"),
      f("cover_media_id", "ID ảnh"),
      f("published", "Xuất bản", "check"),
    ],
    defaults: { country_code: "VN", published: false },
  },
  "tour-categories": {
    label: "Danh mục tour",
    fields: [
      f("name", "Tên"),
      f("slug", "Slug"),
      f("description", "Mô tả", "long"),
      f("sort_order", "Thứ tự", "number"),
    ],
    defaults: { sort_order: 0 },
  },
  posts: {
    label: "Bài viết",
    fields: [
      f("title", "Tiêu đề"),
      f("slug", "Slug"),
      f("excerpt", "Tóm tắt", "long"),
      f("body", "Nội dung HTML", "long"),
      f("category_id", "ID danh mục bài viết"),
      f("tag_ids", "ID thẻ (JSON)", "json"),
      f("cover_media_id", "ID ảnh bìa"),
      f("status", "Trạng thái", "text", ["draft", "published", "archived"]),
      f("seo_title", "Tiêu đề SEO"),
      f("seo_description", "Mô tả SEO"),
    ],
    defaults: { status: "draft", tag_ids: [] },
  },
  pages: {
    label: "Trang thông tin",
    fields: [
      f("title", "Tiêu đề"),
      f("slug", "Slug"),
      f("body", "Nội dung HTML", "long"),
      f("status", "Trạng thái", "text", ["draft", "published", "archived"]),
      f("seo_title", "Tiêu đề SEO"),
      f("seo_description", "Mô tả SEO"),
    ],
    defaults: { status: "draft" },
  },
  "post-categories": {
    label: "Danh mục bài viết",
    fields: [f("name", "Tên"), f("slug", "Slug")],
    defaults: {},
  },
  tags: {
    label: "Thẻ bài viết",
    fields: [f("name", "Tên"), f("slug", "Slug")],
    defaults: {},
  },
  banners: {
    label: "Banner",
    fields: [
      f("title", "Tiêu đề"),
      f("media_id", "ID ảnh"),
      f("target_url", "Liên kết (/ hoặc https)"),
      f("placement", "Vị trí", "text", [
        "home",
        "tours",
        "schools",
        "companies",
      ]),
      f("sort_order", "Thứ tự", "number"),
      f("active", "Hiển thị", "check"),
    ],
    defaults: { placement: "home", sort_order: 0, active: false },
  },
  faqs: {
    label: "Câu hỏi thường gặp",
    fields: [
      f("topic", "Chủ đề"),
      f("tour_id", "ID tour (tuỳ chọn)"),
      f("question", "Câu hỏi"),
      f("answer", "Trả lời HTML", "long"),
      f("sort_order", "Thứ tự", "number"),
      f("published", "Xuất bản", "check"),
    ],
    defaults: { topic: "general", sort_order: 0, published: false },
  },
  reviews: {
    label: "Đánh giá",
    fields: [
      f("display_name", "Tên hiển thị"),
      f("organization_name", "Tổ chức"),
      f("tour_id", "ID tour"),
      f("group_type", "Loại nhóm", "text", ["school", "company"]),
      f("rating", "Điểm 1–5", "number"),
      f("body", "Nội dung", "long"),
      f("source", "Nguồn/quyền sử dụng"),
      f("is_demo", "Đánh giá demo", "check"),
    ],
    defaults: { rating: 5, is_demo: true, group_type: "school" },
  },
};
export default function Catalog({ permissions }: { permissions: string[] }) {
  const [resource, setResource] = useState("tours");
  const [rows, setRows] = useState<Row[]>([]);
  const [record, setRecord] = useState<Row | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [page, setPage] = useState(1);
  const [q, setQ] = useState("");
  const allowed = Object.keys(resources).filter((x) =>
    permissions.includes(
      ["tours", "destinations", "tour-categories"].includes(x)
        ? "catalog.write"
        : "content.write",
    ),
  );
  const current = allowed.includes(resource) ? resource : allowed[0];
  const cfg = resources[current];
  async function load() {
    setError("");
    try {
      setRows(
        await authorized<Row[]>(
          "/admin/" +
            current +
            "?page_size=20&page=" +
            page +
            "&q=" +
            encodeURIComponent(q),
        ),
      );
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không tải được dữ liệu");
    }
  }
  useEffect(() => {
    if (!current) return;
    let active = true;
    authorized<Row[]>(
      "/admin/" +
        current +
        "?page_size=20&page=" +
        page +
        "&q=" +
        encodeURIComponent(q),
    )
      .then((rows) => {
        if (active) setRows(rows);
      })
      .catch((e) => {
        if (active) setError(String(e));
      });
    return () => {
      active = false;
    };
  }, [current, page, q]);
  if (!cfg) return null;
  async function action(path: string, body: Row = {}) {
    setBusy(true);
    setError("");
    try {
      await authorized("/admin/" + path, {
        method: "POST",
        body: JSON.stringify(body),
      });
      setMessage("Đã cập nhật.");
      if (record?.id)
        setRecord(await authorized<Row>("/admin/" + current + "/" + record.id));
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không lưu được dữ liệu");
    } finally {
      setBusy(false);
    }
  }
  async function save(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!record) return;
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const data: Row = {};
      for (const field of cfg.fields) {
        const value = record[field.key];
        if (field.type === "json") {
          data[field.key] =
            typeof value === "string" ? JSON.parse(value || "[]") : value || [];
        } else if (field.type === "number") {
          if (value !== "" && value != null) data[field.key] = Number(value);
        } else if (field.type === "check") data[field.key] = Boolean(value);
        else if (field.key.endsWith("_id")) data[field.key] = value || null;
        else if (value != null) data[field.key] = value;
      }
      const out = await authorized<Row>(
        "/admin/" + current + (record.id ? "/" + record.id : ""),
        { method: record.id ? "PATCH" : "POST", body: JSON.stringify(data) },
      );
      setMessage("Đã lưu nội dung.");
      setRecord(
        await authorized<Row>(
          "/admin/" + current + "/" + (record.id || out.id),
        ),
      );
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Không lưu được dữ liệu");
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="workspace">
      <h2>Danh mục và nội dung</h2>
      <div className="toolbar">
        <label>
          Loại nội dung
          <select
            value={current}
            onChange={(e) => {
              setResource(e.target.value);
              setPage(1);
              setRecord(null);
              setMessage("");
            }}
          >
            {allowed.map((k) => (
              <option key={k} value={k}>
                {resources[k].label}
              </option>
            ))}
          </select>
        </label>
        <label>
          Tìm kiếm
          <input value={q} onChange={(e) => setQ(e.target.value)} />
        </label>
        <button
          onClick={() => {
            setPage(1);
            void load();
          }}
        >
          Tìm nội dung
        </button>
        <button
          onClick={() => {
            setRecord({ ...cfg.defaults });
            setMessage("");
          }}
        >
          Thêm {cfg.label.toLowerCase()}
        </button>
      </div>
      {error && <p role="alert">{error}</p>}
      <p role="status">{message}</p>
      <div className="data-list">
        {rows.length ? (
          rows.map((r) => (
            <button
              key={String(r.id)}
              className="list-row"
              onClick={async () => {
                setError("");
                try {
                  setRecord(
                    await authorized<Row>("/admin/" + current + "/" + r.id),
                  );
                } catch (e) {
                  setError(String(e));
                }
              }}
            >
              <strong>
                {String(
                  r.title || r.name || r.question || r.display_name || r.id,
                )}
              </strong>
              <span>
                {String(r.status ?? r.slug ?? "")} · {String(r.id)}
              </span>
            </button>
          ))
        ) : (
          <p>Chưa có nội dung.</p>
        )}
      </div>
      <div className="toolbar">
        <button disabled={page === 1} onClick={() => setPage(page - 1)}>
          Trang trước
        </button>
        <span>Trang {page}</span>
        <button disabled={rows.length < 20} onClick={() => setPage(page + 1)}>
          Trang tiếp
        </button>
      </div>
      {record && (
        <div className="editor">
          <div className="toolbar">
            <h3>
              {record.id ? "Chỉnh sửa" : "Thêm"} {cfg.label.toLowerCase()}
            </h3>
            <button onClick={() => setRecord(null)}>Đóng trình sửa</button>
          </div>
          {Boolean(record.id) && (
            <p>
              ID: <code>{String(record.id)}</code> ·{" "}
              {String(record.status || "")}
            </p>
          )}
          <form onSubmit={save}>
            <div className="editor-grid">
              {cfg.fields.map((field) => (
                <label
                  key={field.key}
                  className={
                    ["long", "json"].includes(field.type || "") ? "wide" : ""
                  }
                >
                  {field.label}
                  {field.type === "check" ? (
                    <input
                      type="checkbox"
                      checked={Boolean(record[field.key])}
                      onChange={(e) =>
                        setRecord({ ...record, [field.key]: e.target.checked })
                      }
                    />
                  ) : field.options ? (
                    <select
                      value={String(record[field.key] || field.options[0])}
                      onChange={(e) =>
                        setRecord({ ...record, [field.key]: e.target.value })
                      }
                    >
                      {field.options.map((v) => (
                        <option key={v}>{v}</option>
                      ))}
                    </select>
                  ) : ["long", "json"].includes(field.type || "") ? (
                    <textarea
                      rows={field.type === "json" ? 8 : 5}
                      value={
                        typeof record[field.key] === "object"
                          ? JSON.stringify(record[field.key] ?? [], null, 2)
                          : String(record[field.key] ?? "")
                      }
                      onChange={(e) =>
                        setRecord({ ...record, [field.key]: e.target.value })
                      }
                    />
                  ) : (
                    <input
                      type={field.type === "number" ? "number" : "text"}
                      step={
                        field.key === "latitude" || field.key === "longitude"
                          ? "any"
                          : "1"
                      }
                      value={String(record[field.key] ?? "")}
                      onChange={(e) =>
                        setRecord({ ...record, [field.key]: e.target.value })
                      }
                    />
                  )}
                </label>
              ))}
            </div>
            <button disabled={busy}>
              {busy ? "Đang lưu…" : "Lưu nội dung"}
            </button>
          </form>
          {Boolean(record.id) && (
            <>
              <div className="toolbar">
                {current === "tours" ? (
                  <>
                    <button
                      disabled={busy}
                      onClick={() =>
                        void action("tours/" + record.id + "/publish")
                      }
                    >
                      Xuất bản tour
                    </button>
                    <button
                      disabled={busy}
                      onClick={() =>
                        void action("tours/" + record.id + "/archive")
                      }
                    >
                      Lưu trữ tour
                    </button>
                    <button
                      onClick={() => {
                        const slug = prompt(
                          "Slug cho bản sao (không dấu, dùng dấu -)",
                        );
                        if (slug)
                          void action("tours/" + record.id + "/clone", {
                            slug,
                            code:
                              String(record.code) +
                              "-COPY-" +
                              Date.now().toString().slice(-4),
                            title: String(record.title) + " — Bản sao",
                          });
                      }}
                    >
                      Nhân bản tour
                    </button>
                  </>
                ) : (
                  <button
                    onClick={async () => {
                      if (!confirm("Lưu trữ/xoá nội dung này?")) return;
                      try {
                        await authorized(
                          "/admin/" + current + "/" + record.id,
                          { method: "DELETE" },
                        );
                        setRecord(null);
                        await load();
                      } catch (e) {
                        setError(String(e));
                      }
                    }}
                  >
                    Lưu trữ/xoá nội dung
                  </button>
                )}
                {current === "reviews" &&
                  ["approved", "rejected"].map((status) => (
                    <button
                      key={status}
                      onClick={() =>
                        void action("reviews/" + record.id + "/moderation", {
                          status,
                        })
                      }
                    >
                      {status === "approved"
                        ? "Duyệt đánh giá"
                        : "Từ chối đánh giá"}
                    </button>
                  ))}
              </div>
              {current === "tours" && (
                <>
                  <Poster
                    busy={busy}
                    record={record}
                    refresh={async () =>
                      setRecord(
                        await authorized<Row>("/admin/tours/" + record.id),
                      )
                    }
                  />
                  <Departures
                    key={String(record.id)}
                    tour={String(record.id)}
                  />
                </>
              )}
            </>
          )}
        </div>
      )}
    </section>
  );
}
function Poster({
  record,
  refresh,
  busy,
}: {
  busy: boolean;
  record: Row;
  refresh: () => Promise<void>;
}) {
  const [error, setError] = useState("");
  return (
    <section>
      <h3>Đối chiếu poster</h3>
      <p>
        So sánh giá, lịch trình và điều kiện trong ảnh với dữ liệu tour trước
        khi xác minh. Mọi thay đổi tour/giá/lịch làm mất xác minh cũ.
      </p>
      {record.poster_media_id ? (
        <a
          href={"/api/v1/media/" + record.poster_media_id}
          target="_blank"
          rel="noreferrer"
        >
          <img
            className="poster-preview"
            src={"/api/v1/media/" + record.poster_media_id}
            alt="Poster tour cần đối chiếu"
          />
        </a>
      ) : (
        <p>Chưa chọn poster. Tải ảnh ở Media rồi điền ID poster và lưu tour.</p>
      )}
      <button
        disabled={busy || !record.poster_media_id}
        onClick={async () => {
          try {
            await authorized(
              "/admin/tours/" + record.id + "/poster-verification",
              {
                method: "POST",
                body: JSON.stringify({
                  expected_poster_revision: record.poster_revision,
                  confirmed: true,
                }),
              },
            );
            setError("Đã xác minh poster.");
            await refresh();
          } catch (e) {
            setError(String(e));
          }
        }}
      >
        Xác minh poster khớp dữ liệu
      </button>
      <p role="status">{error}</p>
    </section>
  );
}
function Departures({ tour }: { tour: string }) {
  const [rows, setRows] = useState<Row[]>([]);
  const [record, setRecord] = useState<Row | null>(null);
  const [bulk, setBulk] = useState("");
  const [message, setMessage] = useState("");
  async function load() {
    try {
      setRows(
        await authorized<Row[]>(
          "/admin/tours/" + tour + "/departures?page_size=100",
        ),
      );
    } catch (e) {
      setMessage(String(e));
    }
  }
  useEffect(() => {
    let active = true;
    authorized<Row[]>("/admin/tours/" + tour + "/departures?page_size=100")
      .then((rows) => {
        if (active) setRows(rows);
      })
      .catch((e) => {
        if (active) setMessage(String(e));
      });
    return () => {
      active = false;
    };
  }, [tour]);
  const defaults = {
    starts_at: "",
    ends_at: "",
    sales_close_at: "",
    capacity: 30,
    deposit_bps: 3000,
    balance_due_at: "",
    prices: [
      { passenger_type: "adult", unit_price: "1650000" },
      { passenger_type: "child", unit_price: "1155000" },
      { passenger_type: "infant", unit_price: "0" },
    ],
    surcharges: [],
  };
  return (
    <section>
      <h3>Lịch khởi hành và giá</h3>
      <p>
        Tiền VND là số nguyên dạng chuỗi trong JSON. Tuổi/chỗ được lấy từ cài
        đặt khi tạo lịch; mọi lần sửa giữ phiên bản hiện tại.
      </p>
      <button onClick={() => setRecord({ ...defaults })}>
        Thêm lịch khởi hành
      </button>
      {rows.map((r) => (
        <div className="list-row" key={String(r.id)}>
          <strong>
            {new Date(String(r.starts_at)).toLocaleString("vi-VN")}
          </strong>
          <p>
            {String(r.status)} · Còn {String(r.remaining_seats)}/
            {String(r.capacity)} chỗ · Phiên bản giá {String(r.price_version)}
          </p>
          <div className="toolbar">
            <button onClick={() => setRecord(r)}>Sửa lịch và giá</button>
            {["open", "closed"].map((status) => (
              <button
                key={status}
                onClick={async () => {
                  try {
                    await authorized(
                      "/admin/departures/" + r.id + "/sales-status",
                      { method: "POST", body: JSON.stringify({ status }) },
                    );
                    await load();
                    setMessage("Đã cập nhật trạng thái bán.");
                  } catch (e) {
                    setMessage(String(e));
                  }
                }}
              >
                {status === "open" ? "Mở bán" : "Đóng bán"}
              </button>
            ))}
          </div>
        </div>
      ))}
      {record && (
        <form
          onSubmit={async (e) => {
            e.preventDefault();
            try {
              const body: Row = {};
              for (const key of Object.keys(defaults)) {
                const v = record[key];
                if (key === "prices" || key === "surcharges")
                  body[key] = typeof v === "string" ? JSON.parse(v) : v;
                else if (key === "capacity" || key === "deposit_bps")
                  body[key] = Number(v);
                else if (v) body[key] = new Date(String(v)).toISOString();
              }
              if (record.id) body.expected_price_version = record.price_version;
              await authorized(
                record.id
                  ? "/admin/departures/" + record.id
                  : "/admin/tours/" + tour + "/departures",
                {
                  method: record.id ? "PATCH" : "POST",
                  body: JSON.stringify(body),
                },
              );
              setRecord(null);
              await load();
              setMessage("Đã lưu lịch khởi hành.");
            } catch (e) {
              setMessage(String(e));
            }
          }}
        >
          <div className="editor-grid">
            {(
              [
                "starts_at",
                "ends_at",
                "sales_close_at",
                "balance_due_at",
              ] as const
            ).map((key) => (
              <label key={key}>
                {
                  {
                    starts_at: "Khởi hành",
                    ends_at: "Kết thúc",
                    sales_close_at: "Đóng bán",
                    balance_due_at: "Hạn trả dư",
                  }[key]
                }{" "}
                (giờ máy tính)
                <input
                  type="datetime-local"
                  required={key === "starts_at" || key === "ends_at"}
                  value={
                    record[key]
                      ? new Date(
                          new Date(String(record[key])).getTime() -
                            new Date().getTimezoneOffset() * 60000,
                        )
                          .toISOString()
                          .slice(0, 16)
                      : ""
                  }
                  onChange={(e) =>
                    setRecord({
                      ...record,
                      [key]: e.target.value
                        ? new Date(e.target.value).toISOString()
                        : "",
                    })
                  }
                />
              </label>
            ))}
            <label>
              Sức chứa
              <input
                type="number"
                min="1"
                max="10000"
                value={String(record.capacity)}
                onChange={(e) =>
                  setRecord({ ...record, capacity: e.target.value })
                }
              />
            </label>
            <label>
              Tỷ lệ cọc (10000 = 100%)
              <input
                type="number"
                min="1"
                max="10000"
                value={String(record.deposit_bps)}
                onChange={(e) =>
                  setRecord({ ...record, deposit_bps: e.target.value })
                }
              />
            </label>
            {["prices", "surcharges"].map((key) => (
              <label key={key} className="wide">
                {key === "prices"
                  ? "Bảng giá (3 nhóm tuổi)"
                  : "Phụ thu (code, label, amount, unit=per_booking|per_person, required)"}
                <textarea
                  rows={10}
                  value={
                    typeof record[key] === "string"
                      ? (record[key] as string)
                      : JSON.stringify(record[key], null, 2)
                  }
                  onChange={(e) =>
                    setRecord({ ...record, [key]: e.target.value })
                  }
                />
              </label>
            ))}
          </div>
          <button>Lưu lịch và bảng giá</button>
          <button type="button" onClick={() => setRecord(null)}>
            Huỷ chỉnh sửa
          </button>
        </form>
      )}
      <details>
        <summary>Tạo nhiều lịch trong một lần</summary>
        <p>
          Nhập mảng JSON các lịch có cùng cấu trúc. Toàn bộ được tạo hoặc toàn
          bộ bị huỷ nếu có lỗi.
        </p>
        <textarea
          aria-label="Danh sách lịch khởi hành JSON"
          rows={10}
          value={bulk}
          onChange={(e) => setBulk(e.target.value)}
        />
        <button
          onClick={async () => {
            try {
              await authorized("/admin/tours/" + tour + "/departures/bulk", {
                method: "POST",
                body: JSON.stringify({ departures: JSON.parse(bulk) }),
              });
              setBulk("");
              await load();
              setMessage("Đã tạo toàn bộ lịch.");
            } catch (e) {
              setMessage(String(e));
            }
          }}
        >
          Tạo hàng loạt
        </button>
      </details>
      <p role="status">{message}</p>
    </section>
  );
}
