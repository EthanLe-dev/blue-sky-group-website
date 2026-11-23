"use client";
import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { api } from "@/lib/api";
import Link from "next/link";
export default function SearchForm({
  destinations,
  categories,
}: {
  destinations: { slug: string; name: string }[];
  categories: { slug: string; name: string }[];
}) {
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
          placeholder="Nhập tên tour hoặc điểm đến"
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
        Số chỗ cần đặt
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
          <option value="duration_asc">Thời gian ngắn nhất</option>
          <option value="rating_desc">Đánh giá cao nhất</option>
        </select>
      </label>
      <details className="filter-advanced">
        <summary>Bộ lọc nâng cao</summary>
        <div className="advanced-fields">
          <label>
            Điểm đến
            <select
              name="destination"
              defaultValue={params.get("destination") || ""}
            >
              <option value="">Tất cả điểm đến</option>
              {destinations.map((d) => (
                <option key={d.slug} value={d.slug}>
                  {d.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Danh mục tour
            <select name="category" defaultValue={params.get("category") || ""}>
              <option value="">Tất cả danh mục</option>
              {categories.map((d) => (
                <option key={d.slug} value={d.slug}>
                  {d.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Khu vực
            <select name="region" defaultValue={params.get("region") || ""}>
              <option value="">Tất cả</option>
              <option value="domestic">Trong nước</option>
              <option value="international">Quốc tế</option>
            </select>
          </label>
          <label>
            Số ngày tối thiểu
            <input
              name="duration_min"
              type="number"
              min="1"
              max="365"
              defaultValue={params.get("duration_min") || ""}
            />
          </label>
          <label>
            Số ngày tối đa
            <input
              name="duration_max"
              type="number"
              min="1"
              max="365"
              defaultValue={params.get("duration_max") || ""}
            />
          </label>
          <label>
            Ngân sách tối thiểu (đ)
            <input
              name="budget_min"
              type="number"
              min="0"
              step="1"
              defaultValue={params.get("budget_min") || ""}
            />
          </label>
          <label>
            Phương tiện
            <input
              name="transport"
              defaultValue={params.get("transport") || ""}
              placeholder="Ví dụ: Xe du lịch"
            />
          </label>
          <label>
            Mức vận động
            <select
              name="difficulty"
              defaultValue={params.get("difficulty") || ""}
            >
              <option value="">Tất cả</option>
              <option value="easy">Nhẹ nhàng</option>
              <option value="moderate">Vừa phải</option>
              <option value="hard">Thử thách</option>
            </select>
          </label>
          <label>
            Đánh giá tối thiểu
            <select
              name="min_rating"
              defaultValue={params.get("min_rating") || ""}
            >
              <option value="">Tất cả</option>
              {[3, 4, 4.5, 5].map((n) => (
                <option key={n} value={n}>
                  Từ {n}/5
                </option>
              ))}
            </select>
          </label>
        </div>
      </details>
      <button className="btn">Tìm tour</button>
      {[...params].some(([key, value]) => value && key !== "page") && (
        <p className="applied-filters">Đang lọc theo {[
          q && "từ khóa", params.get("destination") && "điểm đến", params.get("category") && "danh mục",
          (params.get("depart_from") || params.get("depart_to")) && "ngày đi",
          (params.get("budget_min") || params.get("budget_max")) && "ngân sách",
          params.get("audience") && "đối tượng", params.get("travellers") && "số chỗ",
          params.get("region") && "khu vực", (params.get("duration_min") || params.get("duration_max")) && "thời lượng",
          params.get("transport") && "phương tiện", params.get("difficulty") && "mức vận động",
          params.get("min_rating") && "đánh giá", params.get("sort") && "sắp xếp",
        ].filter(Boolean).join(", ") || "lựa chọn của bạn"}.</p>
      )}
      <Link className="text-action" href="/tours">
        Xoá bộ lọc
      </Link>
    </form>
  );
}
