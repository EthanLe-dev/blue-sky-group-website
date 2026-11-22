"use client";

import Link from "next/link";
import { useRef, useState, type FormEvent } from "react";
import { leadEvent } from "@/lib/analytics";
import Captcha from "@/components/Captcha";
import { api, ApiError } from "@/lib/api";

type Receipt = {
  code: string;
  status: "new";
  received_at: string;
  response_sla_message: string;
  is_demo: boolean;
};

export default function LeadForm({
  groupType,
  tourId,
  sla,
  mock,
  siteKey,
}: {
  groupType?: "school" | "company";
  tourId?: string;
  sla?: string;
  mock: boolean;
  siteKey: string;
}) {
  const [token, setToken] = useState("");
  const [challenge, setChallenge] = useState(0);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [receipt, setReceipt] = useState<Receipt | null>(null);
  const sending = useRef(false);
  const attempt = useRef<{ key: string; payload: string } | null>(null);
  const action = groupType ? "group-request" : "callback-request";

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (sending.current || !token) return;
    setError("");
    const fields = new FormData(event.currentTarget);
    const field = (name: string) => String(fields.get(name) || "").trim();
    const start = field("preferred_start_date");
    const end = field("preferred_end_date");
    if (end && (!start || end < start)) {
      setError("Chọn ngày bắt đầu và ngày kết thúc bằng hoặc sau ngày bắt đầu.");
      return;
    }
    const consent = { accepted: fields.has("consent"), policy_version: "sales-v1" };
    const budget = field("budget_per_person");
    const input = groupType
      ? {
          group_type: groupType,
          organization_name: field("organization_name"),
          contact_name: field("contact_name"),
          contact_email: field("contact_email"),
          contact_phone: field("phone"),
          participant_count: Number(field("participant_count")),
          destination_text: field("destination_text"),
          preferred_start_date: start || undefined,
          preferred_end_date: end || undefined,
          budget_per_person: budget || undefined,
          requirements: field("message"),
          consent,
        }
      : { phone: field("phone"), name: field("contact_name"), preferred_time: field("preferred_time"),
          tour_id: tourId, message: field("message"), consent };
    const payload = JSON.stringify(input);
    let current = attempt.current;
    if (!current || current.payload !== payload) {
      current = { key: crypto.randomUUID(), payload };
      attempt.current = current;
    }
    sending.current = true;
    setBusy(true);
    try {
      const { csrf_token } = await api<{ csrf_token: string }>("/guest-sessions", {
        method: "POST",
        body: "{}",
      });
      const result = await api<Receipt>("/" + action + "s", {
        method: "POST",
        headers: { "X-CSRF-Token": csrf_token, "Idempotency-Key": current.key },
        body: JSON.stringify({ ...input, turnstile_token: token }),
      });
      setReceipt(result);
      leadEvent(groupType);
      attempt.current = null;
    } catch (err) {
      setError(err instanceof ApiError && err.code === "VALIDATION_ERROR"
        ? "Thông tin chưa hợp lệ. Hãy kiểm tra số điện thoại, email và ngày dự kiến."
        : err instanceof ApiError ? err.message : "Kết nối bị gián đoạn. Vui lòng thử gửi lại.");
      setToken("");
      setChallenge((value) => value + 1);
    } finally {
      sending.current = false;
      setBusy(false);
    }
  }

  return (
    <section id={groupType ? "bao-gia" : "goi-lai"} className="detail-section">
      <h2>{groupType ? "Yêu cầu báo giá đoàn" : "Yêu cầu gọi lại"}</h2>
      {receipt ? (
        <div className="notice" role="status">
          <p>Đã nhận yêu cầu của bạn. Mã yêu cầu <strong>{receipt.code}</strong>.</p>
          <p>{receipt.response_sla_message}</p>
          {receipt.is_demo && <p>Yêu cầu demo.</p>}
        </div>
      ) : (
        <>
          {sla && <p>{sla}</p>}
          <form onSubmit={submit} className="form-grid" aria-busy={busy}>
            <fieldset disabled={busy}>
              <legend>{groupType ? "Thông tin của đoàn" : "Thông tin liên hệ"}</legend>
              {groupType && (
                <label>{groupType === "school" ? "Tên trường" : "Tên doanh nghiệp"}
                  <input name="organization_name" required maxLength={200} />
                </label>
              )}
              <label>Họ và tên
                <input name="contact_name" autoComplete="name" required={!!groupType} minLength={groupType ? 2 : undefined} maxLength={120} />
              </label>
              {groupType && (
                <label>Email
                  <input name="contact_email" type="email" autoComplete="email" required maxLength={254} />
                </label>
              )}
              <label>Số điện thoại
                <input name="phone" type="tel" autoComplete="tel" required maxLength={24} />
              </label>
              {groupType ? (
                <>
                  <label>Số người
                    <input name="participant_count" type="number" min={1} max={10000} step={1} required />
                  </label>
                  <label>Điểm đến
                    <input name="destination_text" required maxLength={200} />
                  </label>
                  <label>Ngày đi dự kiến
                    <input name="preferred_start_date" type="date" />
                  </label>
                  <label>Ngày về dự kiến
                    <input name="preferred_end_date" type="date" />
                  </label>
                  <label>Ngân sách mỗi người (VND)
                    <input name="budget_per_person" type="number" min={0} max={1000000000000} step={1} />
                  </label>
                </>
              ) : (
                <label>Khung giờ thuận tiện
                  <select name="preferred_time">
                    <option value="">Linh hoạt</option>
                    <option>Buổi sáng</option>
                    <option>Buổi chiều</option>
                    <option>Buổi tối</option>
                  </select>
                </label>
              )}
              <label>{groupType ? "Yêu cầu đặc biệt" : "Nội dung cần tư vấn"}
                <textarea name="message" maxLength={5000} rows={3} />
              </label>
              <label className="check">
                <input type="checkbox" name="consent" required />
                Tôi đồng ý cho Blue Sky Group dùng thông tin trên để phản hồi yêu cầu.
              </label>
              <Link href="/policies/privacy">Đọc chính sách bảo mật</Link>
            </fieldset>
            <Captcha key={challenge} action={action} onToken={setToken} mock={mock} siteKey={siteKey} />
            {error && <p className="error" role="alert">{error}</p>}
            <button className="btn" disabled={busy || !token}>
              {busy ? "Đang gửi…" : "Gửi yêu cầu"}
            </button>
          </form>
        </>
      )}
    </section>
  );
}
