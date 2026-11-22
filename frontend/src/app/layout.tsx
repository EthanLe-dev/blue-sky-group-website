import type { Metadata } from "next";
import type { ReactNode } from "react";
import "@fontsource/be-vietnam-pro/vietnamese-400.css";
import "@fontsource/be-vietnam-pro/vietnamese-600.css";
import "@fontsource/be-vietnam-pro/latin-400.css";
import "@fontsource/be-vietnam-pro/latin-600.css";
import "./globals.css";
import Analytics from "@/components/Analytics";
import Script from "next/script";
import { Header, Footer } from "@/components/Shell";
const origin = process.env.SITE_URL || "http://localhost:3000";
export function generateMetadata(): Metadata {
  const verification = process.env.GSC_VERIFICATION || "";
  return {
  verification: /^[A-Za-z0-9_-]{1,200}$/.test(verification) ? { google: verification } : undefined,
  metadataBase: new URL(origin),
  title: {
    default: "Blue Sky Group | Tour Ninh Chữ & Vĩnh Hy",
    template: "%s | Blue Sky Group",
  },
  description:
    "Khám phá tour Ninh Chữ, Vĩnh Hy và hành trình dành cho trường học, doanh nghiệp. Xem lịch, giá và đặt chỗ trực tuyến.",
  openGraph: {
    locale: "vi_VN",
    siteName: "Blue Sky Group",
    type: "website",
    images: ["/images/vinh-hy.jpg"],
  },
  robots: { index: process.env.APP_ENV === "production", follow: true },
  };
}
export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="vi" data-type="service" data-buttons="solid" data-motion="">
      <body>
        <link rel="stylesheet" href="/assets/motion/motion.css" />
        <Script
          type="module"
          src="/assets/motion/motion.js"
          strategy="afterInteractive"
        />
        <Header />
        {children}
        <Footer />
        <Analytics />
      </body>
    </html>
  );
}
