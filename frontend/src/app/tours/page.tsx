import { Suspense } from "react";
import Link from "next/link";
import SearchForm from "@/components/SearchForm";
import TourCard from "@/components/TourCard";
import { read, type Tour } from "@/lib/catalog";
export const metadata = {
  title: "Khám phá tour",
  alternates: { canonical: "/tours" },
};
export default async function Tours({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = await searchParams;
  const query = new URLSearchParams();
  for (const [k, v] of Object.entries(params))
    if (typeof v === "string" && v) query.set(k, v);
  let tours: Tour[] = [];
  let error = false;
  try {
    tours = await read("/tours?" + query);
  } catch {
    error = true;
  }
  return (
    <main id="main" className="page wrap">
      <p className="eyebrow">CHỌN NGÀY ĐI, TÌM HÀNH TRÌNH</p>
      <h1>Khám phá tour</h1>
      <p>
        Giá và số chỗ được cập nhật theo từng lịch khởi hành. Dữ liệu hiện tại
        dùng cho demo.
      </p>
      <Suspense>
        <SearchForm />
      </Suspense>
      {error ? (
        <div className="state" role="alert">
          <h2>Chưa tải được tour</h2>
          <p>Kiểm tra bộ lọc hoặc kết nối rồi thử lại.</p>
          <Link className="btn" href="/tours">
            Tải lại danh mục
          </Link>
        </div>
      ) : tours.length ? (
        <>
          <p>{tours.length} tour trong trang này</p>
          <div className="tour-grid">
            {tours.map((t) => (
              <TourCard key={t.id} tour={t} />
            ))}
          </div>
          <nav className="pagination" aria-label="Phân trang">
            {Number(query.get("page") || 1) > 1 && (
              <Link
                href={
                  "/tours?" +
                  new URLSearchParams({
                    ...Object.fromEntries(query),
                    page: String(Number(query.get("page") || 1) - 1),
                  })
                }
              >
                Trang trước
              </Link>
            )}
            {tours.length === 20 && (
              <Link
                href={
                  "/tours?" +
                  new URLSearchParams({
                    ...Object.fromEntries(query),
                    page: String(Number(query.get("page") || 1) + 1),
                  })
                }
              >
                Trang tiếp
              </Link>
            )}
          </nav>
        </>
      ) : (
        <div className="state">
          <h2>Chưa có tour phù hợp</h2>
          <p>Thử ngày khác hoặc mở rộng ngân sách.</p>
          <Link href="/tours" className="btn">
            Xem tất cả tour
          </Link>
        </div>
      )}
    </main>
  );
}
