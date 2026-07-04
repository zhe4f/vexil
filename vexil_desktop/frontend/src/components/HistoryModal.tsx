import { OpenPath, DeleteHistory } from "../../wailsjs/go/gui/Handler";
import type { HistoryEntry } from "../types";
import { useLanguage } from "../i18n/i18n";

interface Props {
  entry: HistoryEntry;
  onClose: () => void;
  index: number;
  onDeleted: () => void;
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + " MB";
  return (bytes / 1024 / 1024 / 1024).toFixed(1) + " GB";
}

export function HistoryModal({ entry, onClose, index, onDeleted }: Props) {
  const { t } = useLanguage();

  return (
    <div style={overlayStyle} onClick={onClose}>
      <div style={modalStyle} onClick={e => e.stopPropagation()}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
          <span style={{ fontSize: 18, fontWeight: 600, color: "var(--text-primary)" }}>{t("history.detail_title")}</span>
          <button onClick={onClose} style={{ background: "none", border: "none", color: "var(--text-secondary)", fontSize: 18, cursor: "pointer" }}>✕</button>
        </div>

        <div style={{ textAlign: "center", marginBottom: 20 }}>
          <span style={{ fontSize: 40, color: entry.success ? "var(--success)" : "var(--danger)" }}>
            {entry.success ? "✓" : "✗"}
          </span>
          <p style={{ fontSize: 16, color: "var(--text-primary)", marginTop: 8 }}>
            {entry.direction === "send" ? t("history.detail.send_label") : t("history.detail.recv_label")} · {entry.success ? t("history.detail.success") : t("history.detail.failed")}
          </p>
        </div>

        <div style={{ background: "var(--bg-input)", borderRadius: 10, padding: 14 }}>
          {[
            [t("history.detail.files"), entry.file_names?.join(", ") || entry.files + "个文件"],
            [t("history.detail.size"), formatSize(entry.size)],
            [t("history.detail.speed"), entry.speed_mbps.toFixed(1) + " MB/s"],
            [t("history.detail.duration"), entry.duration_sec.toFixed(1) + "s"],
            [t("history.detail.peer_name"), entry.peer_name],
            [t("history.detail.peer_ip"), entry.peer],
            [t("history.detail.time"), entry.time],
            [t("history.detail.save_path"), entry.save_path || "-"],
          ].map(([label, value]) => (
            <div key={label} style={{ display: "flex", justifyContent: "space-between", padding: "6px 0", fontSize: 13, borderBottom: "1px solid var(--border)" }}>
              <span style={{ color: "var(--text-secondary)" }}>{label}</span>
              <span style={{ color: "var(--text-primary)", textAlign: "right", maxWidth: "60%", overflow: "hidden", textOverflow: "ellipsis" }}>{value}</span>
            </div>
          ))}
        </div>

        <div style={{ display: "flex", gap: 10, marginTop: 16 }}>
          <button onClick={() => { if (entry.save_path) OpenPath(entry.save_path); }} style={actionBtn}>{t("button.open_file")}</button>
          <button onClick={async () => { try { await DeleteHistory(index); onDeleted(); onClose(); } catch {} }} style={{ ...actionBtn, color: "var(--danger)", borderColor: "rgba(248,81,73,0.3)" }}>{t("button.delete_file")}</button>
        </div>
      </div>
    </div>
  );
}

const overlayStyle: React.CSSProperties = { position: "fixed", inset: 0, background: "var(--overlay)", display: "flex", alignItems: "center", justifyContent: "center", zIndex: 110 };
const modalStyle: React.CSSProperties = { width: "min(380px, 92vw)", background: "var(--modal-bg)", border: "1px solid var(--border)", borderRadius: 16, padding: 24 };
const actionBtn: React.CSSProperties = { flex: 1, padding: "10px", background: "var(--bg-input)", border: "1px solid var(--border-input)", borderRadius: 10, color: "var(--text-primary)", fontSize: 14, cursor: "pointer", textAlign: "center" };