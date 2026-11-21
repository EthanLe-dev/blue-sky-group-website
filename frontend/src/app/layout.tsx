import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Blue Sky Group",
  description: "Du lịch và trải nghiệm Blue Sky Group",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="vi">
      <body>{children}</body>
    </html>
  );
}
