"use client";
import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { api } from "@/lib/api";
import Link from "next/link";
export default function SearchForm() {
  const params = useSearchParams();
  const [q, setQ] = useState(params.get("q") || "");
  const [items, setItems] = useState<{ slug: string; title: string }[]>([]);
  useEffect(() => {
    if (q.length < 2) return;
    const ctrl = new AbortController();
    const timer = setTimeout(() => {
      api<{ tours: { slug: string; title: string }[] }>(
        "/tours/suggestions?q=" + encodeURIComponent(q),
        { signal: ctrl.signal },
      )
        .then((x) => setItems(x.tours))
        .catch(() => setItems([]));
    }, 250);
    return () => {
      ctrl.abort();
      clearTimeout(timer);
    };
  }, [q]);
  return (
    <form className="filters" action="/tours">
      <div className="search-field">
        <label htmlFor="q">Bạn muốn đi đâu?</label>
        <input
          id="q"
          name="q"
          value={q}
          onChange={(e) => {
            setQ(e.target.value);
            setItems([]);
          }}
          maxLength={100}
          autoComplete="off"
          placeholder="Tìm Ninh Chữ, Vĩnh Hy…"
        />
        {q.length >= 2 && items.length > 0 && (
          <ul className="suggestions" aria-label="Tour gợi ý">
            {items.map((i) => (
              <li key={i.slug}>
                <Link href={"/tours/" + i.slug}>{i.title}</Link>
              </li>
            ))}
          </ul>
        )}
      </div>
      <label>
        Từ ngày
        <input
          name="depart_from"
          type="date"
          defaultValue={params.get("depart_from") || ""}
        />
      </label>
      <label>
        Đến ngày
        <input
          name="depart_to"
          type="date"
          defaultValue={params.get("depart_to") || ""}
        />
      </label>
      <label>
        Ngân sách tối đa (đ)
        <input
          name="budget_max"
          type="number"
          min="0"
          step="1"
          defaultValue={params.get("budget_max") || ""}
        />
      </label>
      <label>
        Số khách
        <input
          name="travellers"
          type="number"
          min="1"
          max="20"
          defaultValue={params.get("travellers") || ""}
        />
      </label>
      <label>
        Đối tượng
        <select name="audience" defaultValue={params.get("audience") || ""}>
          <option value="">Tất cả</option>
          <option value="general">Khách lẻ</option>
          <option value="school">Trường học</option>
          <option value="company">Doanh nghiệp</option>
        </select>
      </label>
      <label>
        Sắp xếp
        <select name="sort" defaultValue={params.get("sort") || ""}>
          <option value="">Đề xuất</option>
          <option value="price_asc">Giá tăng dần</option>
          <option value="price_desc">Giá giảm dần</option>
          <option value="departure_asc">Khởi hành gần nhất</option>
        </select>
      </label>
      <button className="btn">Tìm tour</button>
      <Link className="text-action" href="/tours">
        Xoá bộ lọc
      </Link>
    </form>
  );
}
