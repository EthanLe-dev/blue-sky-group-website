import { useEffect, useRef, useState, type FormEvent } from "react";
import { authorized } from "./lib/session";

const groupPath = "/admin/group-requests";
const callbackPath = "/admin/callback-requests";
const groupStates = ["new", "preparing", "quoted", "won", "lost"];
const callbackStates = ["new", "contacted", "closed"];
const labels: Record<string, string> = {
  new: "Mới", preparing: "Đang chuẩn bị", quoted: "Đã phát hành báo giá",
  won: "Đã chốt", lost: "Không chốt", contacted: "Đã liên hệ", closed: "Đã đóng",
  draft: "Bản nháp", issued: "Đã phát hành", accepted: "Đã xác nhận", rejected: "Đã từ chối",
};
const dateTime = new Intl.DateTimeFormat("vi-VN", {
  timeZone: "Asia/Ho_Chi_Minh", dateStyle: "short", timeStyle: "short",
});
const dateInput = new Intl.DateTimeFormat("sv-SE", {
  timeZone: "Asia/Ho_Chi_Minh", year: "numeric", month: "2-digit", day: "2-digit",
  hour: "2-digit", minute: "2-digit", hourCycle: "h23",
});
const money = new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" });
const errorMessage = (error: unknown) => error instanceof Error ? error.message : "Không thể kết nối. Vui lòng thử lại.";

type Lead = {
  id: string; code: string; status: string; assigned_to: string | null;
  follow_up_due_at: string | null; created_at: string;
  organization_name?: string; group_type?: "school" | "company"; contact_name?: string;
  contact_email?: string; contact_phone?: string; participant_count?: number;
  destination_text?: string; preferred_start_date?: string; preferred_end_date?: string;
  budget_per_person?: string | null; requirements?: string;
  name?: string; phone?: string; preferred_time?: string; message?: string; tour_id?: string;
};
type Quote = {
  id: string; version: number; status: string; title: string; itinerary: string;
  inclusions: string; exclusions: string; total_amount: string; valid_until: string;
  decision_reason?: string;
};
type Staff = { id: string; name: string };
type Detail = Lead & {
  notes: { id: string; body: string; created_at: string }[];
  quotes?: Quote[];
  attachments?: { media_id: string; name: string }[];
};
type Perform = (operation: () => Promise<unknown>, after?: () => void) => Promise<void>;

function write<T>(path: string, method: string, payload: object) {
  return authorized<T>(path, { method, body: JSON.stringify(payload) });
}

function isOverdue(lead: Lead) {
  return lead.follow_up_due_at && !["closed", "won", "lost"].includes(lead.status) &&
    Date.parse(lead.follow_up_due_at) < Date.now();
}

export default function Sales({ canAttach }: { canAttach: boolean }) {
  const [group, setGroup] = useState(true);
  return (
    <section className="workspace" aria-labelledby="sales-heading">
      <h2 id="sales-heading">Yêu cầu tư vấn</h2>
      <div className="admin-tabs">
        <button aria-pressed={group} onClick={() => setGroup(true)}>Báo giá đoàn</button>
        <button aria-pressed={!group} onClick={() => setGroup(false)}>Gọi lại</button>
      </div>
      <LeadQueue key={String(group)} group={group} canAttach={canAttach} />
    </section>
  );
}

function LeadQueue({ group, canAttach }: { group: boolean; canAttach: boolean }) {
  const path = group ? groupPath : callbackPath;
  const states = group ? groupStates : callbackStates;
  const [staff, setStaff] = useState<Staff[]>([]);
  const [leads, setLeads] = useState<Lead[] | null>(null);
  const [error, setError] = useState("");
  const [filter, setFilter] = useState("");
  const [page, setPage] = useState(1);
  const [revision, setRevision] = useState(0);
  const [selected, setSelected] = useState<string | null>(null);
  useEffect(() => {
    let active = true;
    authorized<Staff[]>("/admin/lead-assignees").then(
      (out) => { if (active) setStaff(out); },
      (err) => { if (active) setError(errorMessage(err)); },
    );
    return () => { active = false; };
  }, [revision]);

  useEffect(() => {
    let active = true;
    authorized<Lead[]>(path + "?" + filter + "&page=" + page).then(
      (out) => { if (active) setLeads(out); },
      (err) => { if (active) { setLeads([]); setError(errorMessage(err)); } },
    );
    return () => { active = false; };
  }, [path, filter, page, revision]);

  function reload() {
    setError("");
    setLeads(null);
    setRevision((n) => n + 1);
  }

  return (
    <>
      <form onSubmit={(event) => {
        event.preventDefault();
        const fields = new FormData(event.currentTarget);
        const query = new URLSearchParams();
        for (const [name, value] of fields) if (typeof value === "string" && value) query.set(name, value);
        setFilter(query.toString());
        setPage(1);
        reload();
      }}>
        <label>Trạng thái
          <select name="status">
            <option value="">Tất cả trạng thái</option>
            {states.map((status) => <option key={status} value={status}>{labels[status]}</option>)}
          </select>
        </label>
        <label>Người phụ trách
          <select name="assigned_to">
            <option value="">Tất cả nhân viên</option>
            {staff.map((person) => <option key={person.id} value={person.id}>{person.name}</option>)}
          </select>
        </label>
        {group && <>
          <label>Nhận từ ngày<input name="from" type="date" /></label>
          <label>Nhận đến ngày<input name="to" type="date" /></label>
        </>}
        <button>Lọc yêu cầu</button>
      </form>
      {error && <p role="alert">{error}</p>}
      {leads === null ? <p role="status">Đang tải yêu cầu…</p> : (
        <ul className="data-list">
          {leads.map((lead) => (
            <li key={lead.id}>
              <button className="list-row" aria-pressed={selected === lead.id} onClick={() => setSelected(lead.id)}>
                <strong>{lead.code} · {lead.organization_name || lead.name || lead.phone}</strong>
                <span>{labels[lead.status]}{isOverdue(lead) && " · Quá hạn theo dõi"}</span>
              </button>
            </li>
          ))}
          {leads.length === 0 && !error && <li>Chưa có yêu cầu phù hợp.</li>}
        </ul>
      )}
      <div className="admin-tabs">
        <button disabled={page === 1} onClick={() => { setPage(page - 1); setLeads(null); }}>Trang trước</button>
        <span>Trang {page}</span>
        <button disabled={leads?.length !== 20} onClick={() => { setPage(page + 1); setLeads(null); }}>Trang sau</button>
        {error && <button onClick={reload}>Thử lại</button>}
      </div>
      {selected && <LeadEditor key={selected} path={path + "/" + selected} staff={staff} canAttach={canAttach} reload={reload} />}
    </>
  );
}

const transitions: Record<string, string[]> = {
  new: ["preparing", "contacted"], preparing: ["lost"], quoted: ["lost"],
  lost: ["preparing"], contacted: ["closed"],
};

function LeadEditor({ path, staff, canAttach, reload }: {
  path: string; staff: Staff[]; canAttach: boolean; reload: () => void;
}) {
  const [lead, setLead] = useState<Detail | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState("");
  const [retry, setRetry] = useState(0);
  const [next, setNext] = useState("");
  const [quote, setQuote] = useState<{ value?: Quote; copy: boolean } | null>(null);
  const running = useRef(false);
  useEffect(() => {
    let active = true;
    authorized<Detail>(path).then(
      (out) => { if (active) setLead(out); },
      (err) => { if (active) setError(errorMessage(err)); },
    );
    return () => { active = false; };
  }, [path, retry]);

  const perform: Perform = async (operation, after) => {
    if (running.current) return;
    running.current = true;
    setBusy(true);
    setError("");
    setSaved("");
    try {
      await operation();
      after?.();
      setRetry((n) => n + 1);
      reload();
      setSaved("Đã lưu thay đổi.");
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      running.current = false;
      setBusy(false);
    }
  };

  if (!lead) {
    return error
      ? <><p role="alert">{error}</p><button onClick={() => { setError(""); setRetry((n) => n + 1); }}>Thử tải lại</button></>
      : <p role="status">Đang tải chi tiết…</p>;
  }

  const group = path.startsWith(groupPath + "/");
  const canQuote = lead.status === "preparing" || lead.status === "quoted";
  const moves = (transitions[lead.status] || []).filter((status) =>
    (group ? groupStates : callbackStates).includes(status));

  function saveLead(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const fields = new FormData(event.currentTarget);
    const due = String(fields.get("follow_up_due_at") || "");
    void perform(() => write(path, "PATCH", {
      status: next || undefined,
      assigned_to: fields.has("assigned_to") ? fields.get("assigned_to") || null : undefined,
      follow_up_due_at: due ? new Date(due + ":00+07:00").toISOString() : null,
      reason: fields.get("reason") || undefined,
    }), () => setNext(""));
  }

  return (
    <article className="editor" aria-label={"Yêu cầu " + lead.code}>
      <h3>{lead.code} · {labels[lead.status]}</h3>
      <p>Tiếp nhận {dateTime.format(new Date(lead.created_at))}{isOverdue(lead) && " · Quá hạn theo dõi"}</p>
      {group ? (
        <>
          <p>{lead.organization_name} · {lead.group_type === "school" ? "Trường học" : "Doanh nghiệp"}</p>
          <p>{lead.contact_name} · {lead.contact_phone} · {lead.contact_email}</p>
          <p>{lead.participant_count} người · {lead.destination_text}</p>
          {lead.preferred_start_date && <p>Ngày dự kiến {lead.preferred_start_date}{lead.preferred_end_date && " đến " + lead.preferred_end_date}</p>}
          {lead.budget_per_person != null && <p>Ngân sách {money.format(Number(lead.budget_per_person))} mỗi người</p>}
          {lead.requirements && <p>{lead.requirements}</p>}
        </>
      ) : (
        <>
          <p>{lead.name && lead.name + " · "}{lead.phone}</p>
          {lead.preferred_time && <p>Khung giờ {lead.preferred_time}</p>}
          {lead.tour_id && <p>Mã tour {lead.tour_id}</p>}
          {lead.message && <p>{lead.message}</p>}
        </>
      )}
      {error && <p role="alert">{error}</p>}
      {saved && <p role="status">{saved}</p>}
      <form key={lead.status + lead.assigned_to + lead.follow_up_due_at} onSubmit={saveLead}>
        <fieldset disabled={busy}>
          <legend>Phân công và theo dõi</legend>
          <label>Giao cho nhân viên
            <select name="assigned_to" defaultValue={lead.assigned_to || ""}>
              <option value="">Chưa phân công</option>
              {lead.assigned_to && !staff.some((person) => person.id === lead.assigned_to) &&
                <option value={lead.assigned_to} disabled>Người phụ trách hiện không nhận việc</option>}
              {staff.map((person) => <option key={person.id} value={person.id}>{person.name}</option>)}
            </select>
          </label>
          <label>Hạn theo dõi (giờ Việt Nam)
            <input name="follow_up_due_at" type="datetime-local" defaultValue={
              lead.follow_up_due_at ? dateInput.format(new Date(lead.follow_up_due_at)).replace(" ", "T") : ""
            } />
          </label>
          {moves.length > 0 && <label>Chuyển trạng thái
            <select value={next} onChange={(event) => setNext(event.target.value)}>
              <option value="">Giữ {labels[lead.status].toLowerCase()}</option>
              {moves.map((status) => <option key={status} value={status}>{labels[status]}</option>)}
            </select>
          </label>}
          {lead.status === "lost" && next === "preparing" && <label>Lý do mở lại
            <textarea name="reason" required maxLength={500} />
          </label>}
          <button>Lưu phân công</button>
        </fieldset>
      </form>
      <h4>Ghi chú nội bộ</h4>
      {lead.notes.length ? <ul>{lead.notes.map((note) => (
        <li key={note.id}><time dateTime={note.created_at}>{dateTime.format(new Date(note.created_at))}</time> · {note.body}</li>
      ))}</ul> : <p>Chưa có ghi chú.</p>}
      <form onSubmit={(event) => {
        event.preventDefault();
        const form = event.currentTarget;
        const body = new FormData(form).get("body");
        void perform(() => write(path + "/notes", "POST", { body }), () => form.reset());
      }}>
        <label>Ghi chú mới<textarea name="body" required maxLength={5000} disabled={busy} /></label>
        <button disabled={busy}>Thêm ghi chú</button>
      </form>
      {group && (
        <>
          <h4>Phiên bản báo giá</h4>
          {lead.quotes?.length === 0 && <p>Chưa có báo giá.</p>}
          {canQuote && <button disabled={busy} onClick={() => setQuote({ copy: false })}>Tạo báo giá</button>}
          {quote && canQuote && (
            <QuoteEditor key={(quote.value?.id || "new") + quote.copy} path={path} value={quote.value} copy={quote.copy}
              busy={busy} perform={perform} close={() => setQuote(null)} />
          )}
          {lead.quotes?.map((item) => (
            <div key={item.id}>
              <h5>Phiên bản {item.version} · {item.title} · {labels[item.status]}</h5>
              <p>{money.format(Number(item.total_amount))} · Hiệu lực đến {item.valid_until}</p>
              <details>
                <summary>Xem nội dung báo giá</summary>
                <p>Lịch trình</p><p style={{ whiteSpace: "pre-wrap" }}>{item.itinerary}</p>
                <p>Bao gồm</p><p style={{ whiteSpace: "pre-wrap" }}>{item.inclusions || "—"}</p>
                <p>Không bao gồm</p><p style={{ whiteSpace: "pre-wrap" }}>{item.exclusions || "—"}</p>
                {item.decision_reason && <p>Kết quả liên hệ: {item.decision_reason}</p>}
              </details>
              {canQuote && (item.status === "draft" ? (
                <div className="admin-tabs">
                  <button disabled={busy} onClick={() => setQuote({ value: item, copy: false })}>Sửa bản nháp</button>
                  <button disabled={busy} onClick={() => void perform(() => write("/admin/group-quotes/" + item.id + "/issue", "POST", {}), () => setQuote(null))}>Phát hành</button>
                </div>
              ) : <button disabled={busy} onClick={() => setQuote({ value: item, copy: true })}>Tạo phiên bản từ báo giá này</button>)}
              {item.status === "issued" && lead.status === "quoted" && (
                <form onSubmit={(event) => {
                  event.preventDefault();
                  const fields = new FormData(event.currentTarget);
                  void perform(() => write("/admin/group-quotes/" + item.id + "/decision", "POST",
                    { status: fields.get("status"), reason: fields.get("reason") }));
                }}>
                  <label>Kết quả trao đổi ngoại tuyến
                    <select name="status">
                      <option value="accepted">Khách đã xác nhận</option>
                      <option value="rejected">Khách từ chối</option>
                    </select>
                  </label>
                  <label>Lý do hoặc cách xác minh<textarea name="reason" required maxLength={500} /></label>
                  <button disabled={busy}>Ghi nhận kết quả</button>
                </form>
              )}
            </div>
          ))}
          <LeadAttachments lead={lead} path={path} canAttach={canAttach} busy={busy} perform={perform} />
        </>
      )}
    </article>
  );
}

function QuoteEditor({ path, value, copy, busy, perform, close }: {
  path: string; value?: Quote; copy: boolean; busy: boolean; perform: Perform; close: () => void;
}) {
  const editing = value && !copy;
  return (
    <form onSubmit={(event) => {
      event.preventDefault();
      const fields = new FormData(event.currentTarget);
      void perform(() => write(editing ? "/admin/group-quotes/" + value.id : path + "/quotes",
        editing ? "PATCH" : "POST", {
          title: fields.get("title"), itinerary: fields.get("itinerary"), inclusions: fields.get("inclusions"),
          exclusions: fields.get("exclusions"), total_amount: fields.get("total_amount"),
          valid_until: fields.get("valid_until"), currency: "VND",
        }), close);
    }}>
      <fieldset disabled={busy}>
        <legend>{editing ? "Sửa bản nháp" : "Bản báo giá mới"}</legend>
        <label>Tiêu đề<input name="title" defaultValue={value?.title} required maxLength={200} /></label>
        <label>Lịch trình<textarea name="itinerary" defaultValue={value?.itinerary} required maxLength={20000} /></label>
        <label>Bao gồm<textarea name="inclusions" defaultValue={value?.inclusions} maxLength={5000} /></label>
        <label>Không bao gồm<textarea name="exclusions" defaultValue={value?.exclusions} maxLength={5000} /></label>
        <label>Tổng tiền (VND)<input name="total_amount" type="number" min={1} max={1000000000000} step={1} required defaultValue={value?.total_amount} /></label>
        <label>Hiệu lực đến ngày<input name="valid_until" type="date" defaultValue={value?.valid_until} required /></label>
        <button>Lưu bản nháp</button>
        <button type="button" onClick={close}>Hủy sửa</button>
      </fieldset>
    </form>
  );
}

function LeadAttachments({ lead, path, canAttach, busy, perform }: {
  lead: Detail; path: string; canAttach: boolean; busy: boolean; perform: Perform;
}) {
  const [download, setDownload] = useState<{ media: string; url: string } | null>(null);
  const [error, setError] = useState("");
  const pending = useRef<{ file: File; id: string } | null>(null);
  return (
    <>
      <h4>Tệp đính kèm riêng tư</h4>
      {lead.attachments?.length ? (
        <ul>{lead.attachments.map((attachment) => (
          <li key={attachment.media_id}>
            {attachment.name}{" "}
            {canAttach && (download?.media === attachment.media_id
              ? <a href={download.url} target="_blank" rel="noreferrer" onClick={() => setDownload(null)}>Tải xuống</a>
              : <button onClick={async () => {
                setError("");
                try {
                  const { url } = await authorized<{ url: string }>(path + "/attachments/" + attachment.media_id + "/download-url", { method: "POST" });
                  setDownload({ media: attachment.media_id, url });
                } catch (err) { setError(errorMessage(err)); }
              }}>Tạo liên kết tải</button>)}
          </li>
        ))}</ul>
      ) : <p>Chưa có tệp đính kèm.</p>}
      {error && <p role="alert">{error}</p>}
      {canAttach && (
        <form onSubmit={(event) => {
          event.preventDefault();
          const form = event.currentTarget;
          const input = form.elements.namedItem("file");
          const file = input instanceof HTMLInputElement ? input.files?.[0] : undefined;
          if (!(file instanceof File) || !file.size) return;
          void perform(async () => {
            let item = pending.current;
            if (!item || item.file !== file) {
              const body = new FormData();
              body.set("file", file);
              const { id } = await authorized<{ id: string }>("/admin/media?visibility=private", { method: "POST", body });
              item = { file, id };
              pending.current = item;
            }
            await write(path + "/attachments", "POST", { media_id: item.id });
          }, () => { pending.current = null; form.reset(); });
        }}>
          <label>Tệp báo giá
            <input name="file" type="file" accept=".pdf,.jpg,.jpeg,.png,.webp" required disabled={busy} />
          </label>
          <button disabled={busy}>Đính kèm</button>
        </form>
      )}
    </>
  );
}
