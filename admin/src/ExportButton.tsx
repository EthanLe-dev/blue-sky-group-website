import { useRef, useState } from "react";
import { downloadAuthorized } from "./lib/session";

export default function ExportButton({ path, filename, label }: { path: string; filename: string; label: string }) {
  const [format, setFormat] = useState("xlsx");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const lock = useRef(false);
  return (
    <div className="toolbar">
      <label>Định dạng {label}
        <select aria-label={"Định dạng " + label} value={format} onChange={e => setFormat(e.target.value)} disabled={busy}>
          <option value="xlsx">Excel (.xlsx)</option><option value="csv">CSV (.csv)</option>
        </select>
      </label>
      <button type="button" disabled={busy} onClick={async () => {
        if (lock.current) return;
        lock.current = true;
        setBusy(true);
        setError("");
        try {
          await downloadAuthorized(path + (path.includes("?") ? "&" : "?") + "format=" + format, filename + "." + format);
        } catch (error) {
          setError(error instanceof Error ? error.message : "Không tải được file.");
        } finally {
          lock.current = false;
          setBusy(false);
        }
      }}>{busy ? "Đang tạo file…" : "Tải " + label}</button>
      {error && <p role="alert">{error}</p>}
    </div>
  );
}
