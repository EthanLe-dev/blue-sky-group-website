import type { Metadata } from "next";
import type { ReactNode } from "react";
import "@fontsource-variable/nunito-sans/wght.css";
import "./globals.css";
import "./conversion.css";
import Analytics from "@/components/Analytics";
import Script from "next/script";
import { resolveContact } from "@/lib/contact";
import { readCompanySettings } from "@/lib/company-profile";
import { SiteHeader, SiteFooter, SiteSupport } from "@/components/Shell";
const origin = process.env.SITE_URL || "http://localhost:3000";
export function generateMetadata(): Metadata {
  const verification = process.env.GSC_VERIFICATION || "";
  return {
  verification: /^[A-Za-z0-9_-]{1,200}$/.test(verification) ? { google: verification } : undefined,
  metadataBase: new URL(origin),
  title: {
    default: "Blue Sky Group | Du lịch & Trải nghiệm",
    template: "%s | Blue Sky Group",
  },
  description:
    "Tour du lịch cho gia đình, trường học và doanh nghiệp. Xem lịch, giá và đặt chỗ trực tuyến.",
  openGraph: {
    locale: "vi_VN",
    siteName: "Blue Sky Group",
    type: "website",
    images: ["/images/vinh-hy.jpg"],
  },
  robots: { index: process.env.APP_ENV === "production", follow: true },
  };
}
export default async function RootLayout({ children }: { children: ReactNode }) {
  const settings = await readCompanySettings();
  const contact = resolveContact(settings?.company_profile);
  return (
    <html lang="vi" data-type="service" data-buttons="solid" data-motion="">
      <body>
        <link rel="stylesheet" href="/assets/motion/motion.css" />
        <Script
          type="module"
          src="/assets/motion/motion.js"
          strategy="afterInteractive"
        />
        <SiteHeader contact={contact} />
        {children}
        <SiteFooter />
        <SiteSupport contact={contact} />
        <Analytics />
      </body>
    </html>
  );
}
