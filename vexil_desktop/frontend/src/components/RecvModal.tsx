import { useState, useCallback } from "react";
import { StartRecv, CancelTransfer, OpenDirDialog } from "../../wailsjs/go/gui/Handler";
import type { TransferView } from "../types";
import { useLanguage } from "../i18n/i18n";

interface Props {
  onClose: () => void;
  view: TransferView | null;
  isActive: boolean;
}

export function RecvModal({ onClose, view, isActive }: Props) {
  const { t } = useLanguage();
  const [port, setPort] = useState(() => {
    try { return JSON.parse(localStorage.getItem("vexil_settings") || "{}").port || "9999"; } catch { return "9999"; }
  });
  const [saveDir, setSaveDir] = useState(() => {
    try { return JSON.parse(localStorage.getItem("vexil_settings") || "{}").saveDir || "./downloads"; } catch { return "./downloads"; }
  });
  const [taskID, setTaskID] = useState("");
  const [error, setError] = useState("");

  const handleStart = useCallback(async () => {
    const p = parseInt(port);
    if (isNaN(p) || p < 1 || p > 65535) { setError(t("device.error_port")); return; }
    setError("");
    try { setTaskID(await StartRecv(p, saveDir)); } catch (e: any) { setError(e?.toString() || t("recv.title")); }
  }, [port, saveDir, t]);

  return (
    <div style={overlayStyle} onClick={isActive ? undefined : onClose}>
      <div style={modalStyle} onClick={e => e.stopPropagation()}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
          <span style={{ fontSize: 18, fontWeight: 600, color: "var(--text-primary)" }}>{t("recv.title")}</span>
          <button onClick={onClose} style={closeBtn}>✕</button>
        </div>

        {!isActive ? (
          <>
            <label style={{ ...labelStyle, marginTop: 16 }}>{t("recv.save_dir")}</label>
            <div style={{ display: "flex", gap: 8 }}>
              <input value={saveDir} onChange={e => setSaveDir(e.target.value)} style={inputStyle} />
              <button onClick={async () => { try { const r = await OpenDirDialog(); if (r) setSaveDir(r); } catch {} }} style={iconBtn}>📂</button>
            </div>
            {error && <p style={{ color: "var(--danger)", fontSize: 13, marginTop: 8 }}>{error}</p>}
            <button onClick={handleStart} className="save-btn" style={primaryBtn}>{t("button.start_recv")}</button>
          </>
        ) : (
          <ProgressView view={view!} taskID={taskID} onCancel={async () => { try { await CancelTransfer(taskID); } catch {} }} t={t} />
        )}
      </div>
    </div>
  );
}

function ProgressView({ view, taskID, onCancel, t }: { view: TransferView; taskID: string; onCancel: () => void; t: (key: string) => string }) {
  const done = view.state === "completed" || view.state === "failed" || view.state === "cancelled";
  return (
    <div style={{ textAlign: "center" }}>
      <p style={{ fontSize: 15, color: "var(--text-secondary)", marginBottom: 16 }}>
        {view.state === "preparing" ? t("progress.preparing") :
         view.state === "connecting" ? t("recv.progress.waiting") :
         view.state === "running" ? t("recv.progress.receiving") :
         view.state === "finalizing" ? t("recv.progress.finalizing") :
         view.state === "completed" ? t("recv.progress.completed") :
         view.state === "failed" ? t("recv.progress.failed") : t("recv.progress.cancelled")}
      </p>
      {view.error && <p style={{ color: "var(--danger)", fontSize: 13, marginBottom: 12 }}>{view.error}</p>}
      <div style={{ height: 6, background: "var(--bg-input)", borderRadius: 3, overflow: "hidden", marginBottom: 12 }}>
        <div style={{ width: `${view.percent}%`, height: "100%", background: done ? "var(--success)" : "linear-gradient(90deg, var(--accent), var(--accent2))", borderRadius: 3, transition: "width 0.3s" }} />
      </div>
      <div style={{ display: "flex", justifyContent: "space-between", fontSize: 13, color: "var(--text-secondary)" }}>
        <span>{view.percent.toFixed(1)}%</span>
        <span>{view.sent} / {view.total}</span>
        <span>{view.speed}</span>
        <span>{view.eta}</span>
      </div>
      {!done && <button onClick={onCancel} style={{ marginTop: 16, padding: "8px 24px", background: "rgba(248,81,73,0.15)", border: "1px solid rgba(248,81,73,0.3)", borderRadius: 8, color: "var(--danger)", fontSize: 14, cursor: "pointer" }}>{t("button.cancel")}</button>}
    </div>
  );
}

const overlayStyle: React.CSSProperties = { position: "fixed", inset: 0, background: "var(--overlay)", display: "flex", alignItems: "center", justifyContent: "center", zIndex: 100 };
const modalStyle: React.CSSProperties = { width: "min(380px, 92vw)", maxHeight: "85vh", overflow: "auto", background: "var(--modal-bg)", border: "1px solid var(--border)", borderRadius: 16, padding: 24 };
const closeBtn: React.CSSProperties = { background: "none", border: "none", color: "var(--text-secondary)", fontSize: 18, cursor: "pointer" };
const labelStyle: React.CSSProperties = { fontSize: 13, color: "var(--text-secondary)", display: "block" };
const inputStyle: React.CSSProperties = { width: "100%", padding: "10px 14px", marginTop: 4, background: "var(--bg-input)", border: "1px solid var(--border-input)", borderRadius: 10, color: "var(--text-primary)", fontSize: 14, outline: "none", boxSizing: "border-box" };
const iconBtn: React.CSSProperties = { width: 44, padding: 0, background: "var(--bg-input)", border: "1px solid var(--border-input)", borderRadius: 10, color: "var(--text-primary)", fontSize: 18, cursor: "pointer" };
const primaryBtn: React.CSSProperties = { width: "100%", marginTop: 20, padding: "12px", background: "linear-gradient(135deg, var(--accent), var(--accent2))", border: "none", borderRadius: 12, color: "#fff", fontSize: 15, fontWeight: 600, cursor: "pointer" };