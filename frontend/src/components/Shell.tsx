"use client";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useRef, useState } from "react";
const links = [
  ["/tours", "Khám phá tour"],
  ["/schools", "Trường học"],
  ["/companies", "Doanh nghiệp"],
  ["/blog", "Cẩm nang"],
  ["/lookup", "Tra cứu đơn"],
];
type HomeContact = { hotline: string; phoneHref: string } | null;

export function SiteHeader() {
  return usePathname() === "/" ? null : <Header />;
}

export function SiteFooter() {
  return usePathname() === "/" ? null : <Footer />;
}

export function Header({ home = false, contact = null }: { home?: boolean; contact?: HomeContact }) {
  const dialog = useRef<HTMLDialogElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const [open, setOpen] = useState(false);
  function close() {
    dialog.current?.close();
    setOpen(false);
    trigger.current?.focus();
  }
  const navigation = home
    ? [links[0], ["/destinations", "Điểm đến"], ...links.slice(1)]
    : links;
  const contactLabel = home ? "Liên hệ" : "Liên hệ ↗";
  return (
    <header className={home ? "site-header nav home-header" : "site-header nav"}>
      <a className="skip-link" href="#main">
        Bỏ qua đến nội dung
      </a>
      {home && (
        <div className="home-support">
          <div className="wrap">
            {contact
              ? <a href={contact.phoneHref}>Hỗ trợ: {contact.hotline}</a>
              : <Link prefetch={false} href="/contact">Hỗ trợ chuyến đi</Link>}
          </div>
        </div>
      )}
      <div className="wrap nav-centre">
        <Link prefetch={false}
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
          {navigation.map(([href, label]) => (
            <Link prefetch={false} href={href} key={href}>
              {label}
            </Link>
          ))}
        </nav>
        <Link prefetch={false} className="header-contact" href="/contact">
          {contactLabel}
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
            {[...navigation, ["/contact", "Liên hệ"]].map(([href, label]) => (
              <Link prefetch={false} key={href} href={href} onClick={close}>
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
          <Link prefetch={false} className="btn" href="/tours">
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
            <Link prefetch={false} href="/about">Về Blue Sky Group</Link>
            <Link prefetch={false} href="/destinations">Điểm đến</Link>
            <Link prefetch={false} href="/gallery">Thư viện ảnh</Link>
            <Link prefetch={false} href="/faq">Câu hỏi thường gặp</Link>
            <Link prefetch={false} href="/contact">Thông tin liên hệ</Link>
          </nav>
          <nav aria-label="Chính sách">
            <Link prefetch={false} href="/policies/privacy">Bảo mật thông tin</Link>
            <Link prefetch={false} href="/policies/terms">Điều khoản sử dụng</Link>
            <Link prefetch={false} href="/policies/payment">Chính sách thanh toán</Link>
            <Link prefetch={false} href="/policies/cancellation">Chính sách huỷ tour</Link>
            <Link prefetch={false} href="/policies/preparation">Chuẩn bị trước chuyến đi</Link>
          </nav>
        </div>
        <div className="footer-bottom">
          <span>© {new Date().getFullYear()} Blue Sky Group</span>
          <Link prefetch={false} href="/credits">Nguồn ảnh và quyền sử dụng</Link>
        </div>
      </div>
    </footer>
  );
}
