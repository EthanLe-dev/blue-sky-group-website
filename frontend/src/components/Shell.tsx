"use client";
import Image from "next/image";
import Link from "next/link";
import { useRef, useState } from "react";
const links = [
  ["/tours", "Khám phá tour"],
  ["/schools", "Trường học"],
  ["/companies", "Doanh nghiệp"],
  ["/blog", "Cẩm nang"],
  ["/lookup", "Tra cứu đơn"],
];
export function Header() {
  const dialog = useRef<HTMLDialogElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const [open, setOpen] = useState(false);
  function close() {
    dialog.current?.close();
    setOpen(false);
    trigger.current?.focus();
  }
  return (
    <header className="site-header nav">
      <a className="skip-link" href="#main">
        Bỏ qua đến nội dung
      </a>
      <div className="wrap nav-centre">
        <Link
          className="brand"
          href="/"
          aria-label="Blue Sky Group — Trang chủ"
        >
          <Image
            src="/brand/logo.jpg"
            loading="eager"
            sizes="70px"
            width="70"
            height="70"
            alt="Logo Blue Sky Group"
          />
          <span>BLUE SKY GROUP</span>
        </Link>
        <nav aria-label="Điều hướng chính">
          {links.map(([href, label]) => (
            <Link href={href} key={href}>
              {label}
            </Link>
          ))}
        </nav>
        <Link className="header-contact" href="/contact">
          Liên hệ ↗
        </Link>
        <button
          ref={trigger}
          className="menu-toggle"
          aria-expanded={open}
          aria-controls="mobile-menu"
          onClick={() => {
            dialog.current?.showModal();
            setOpen(true);
          }}
        >
          Menu ☰
        </button>
        <dialog
          id="mobile-menu"
          ref={dialog}
          onCancel={() => {
            setOpen(false);
            trigger.current?.focus();
          }}
        >
          <button onClick={close} autoFocus>
            Đóng menu ×
          </button>
          <nav aria-label="Điều hướng di động">
            {[...links, ["/contact", "Liên hệ"]].map(([href, label]) => (
              <Link key={href} href={href} onClick={close}>
                {label}
              </Link>
            ))}
          </nav>
        </dialog>
      </div>
    </header>
  );
}
export function Footer() {
  return (
    <footer className="closing">
      <div className="wrap">
        <div className="close-type">
          <h2>Chọn hành trình tiếp theo của bạn.</h2>
          <Link className="btn" href="/tours">
            Tìm tour ↗
          </Link>
        </div>
        <div className="footer-grid">
          <div>
            <strong>BLUE SKY GROUP</strong>
            <p>Du lịch · Sự kiện · Truyền thông</p>
            <p className="demo-note">
              Bản demo — giá, lịch và nội dung cần được xác nhận trước khi bán.
            </p>
          </div>
          <nav aria-label="Thông tin">
            <Link href="/about">Về Blue Sky Group</Link>
            <Link href="/destinations">Điểm đến</Link>
            <Link href="/gallery">Thư viện ảnh</Link>
            <Link href="/faq">Câu hỏi thường gặp</Link>
            <Link href="/contact">Thông tin liên hệ</Link>
          </nav>
          <nav aria-label="Chính sách">
            <Link href="/policies/privacy">Bảo mật thông tin</Link>
            <Link href="/policies/terms">Điều khoản sử dụng</Link>
            <Link href="/policies/payment">Chính sách thanh toán</Link>
            <Link href="/policies/cancellation">Chính sách huỷ tour</Link>
            <Link href="/policies/preparation">Chuẩn bị trước chuyến đi</Link>
          </nav>
        </div>
        <div className="footer-bottom">
          <span>© {new Date().getFullYear()} Blue Sky Group</span>
          <Link href="/credits">Nguồn ảnh và quyền sử dụng</Link>
        </div>
      </div>
    </footer>
  );
}
