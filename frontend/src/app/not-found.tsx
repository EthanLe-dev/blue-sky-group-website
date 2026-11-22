import Link from "next/link";
export default function NotFound() {
  return (
    <main id="main" className="page wrap state">
      <h1>Không tìm thấy trang</h1>
      <p>Nội dung có thể đã được thay đổi hoặc ngừng xuất bản.</p>
      <Link className="btn" href="/tours">
        Khám phá tour
      </Link>
    </main>
  );
}
