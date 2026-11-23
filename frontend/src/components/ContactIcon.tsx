export default function ContactIcon({ channel }: { channel: "phone" | "zalo" | "facebook" }) {
  if (channel === "zalo") return <span className="contact-zalo-mark" aria-hidden="true">Zalo</span>;
  if (channel === "facebook") return <span className="contact-facebook-mark" aria-hidden="true">f</span>;
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="m7 3 3 5-2.5 2a15 15 0 0 0 6.5 6.5L16 14l5 3v3a2 2 0 0 1-2 2C9.6 21.4 2.6 14.4 2 5a2 2 0 0 1 2-2Z" />
    </svg>
  );
}
