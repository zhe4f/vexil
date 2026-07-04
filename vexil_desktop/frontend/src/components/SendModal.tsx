import { useState, useCallback, useEffect } from "react";
import { StartSend, CancelTransfer, DiscoverDevices, OpenFileDialog, OpenDirDialog } from "../../wailsjs/go/gui/Handler";
import { OnFileDrop, OnFileDropOff } from "../../wailsjs/runtime/runtime";
import type { TransferView } from "../types";
import { useLanguage } from "../i18n/i18n";

interface DeviceItem {
  name: string;
  ip: string;
  port: number;
  online: boolean;
  source: "discovered" | "history" | "manual";
}

interface Props {
  onClose: () => void;
  view: TransferView | null;
  isActive: boolean;
  historyPeers: string[];
}

export function SendModal({ onClose, view, isActive, historyPeers }: Props) {
  const { t } = useLanguage();
  const [paths, setPaths] = useState<string[]>([]);
  const [devices, setDevices] = useState<DeviceItem[]>([]);
  const [selected, setSelected] = useState<DeviceItem | null>(null);
  const [discovering, setDiscovering] = useState(false);
  const [showManual, setShowManual] = useState(false);
  const [manualIP, setManualIP] = useState("");
  const [manualPort, setManualPort] = useState("9999");
  const [taskID, setTaskID] = useState("");
  const [error, setError] = useState("");
  const [dragOver, setDragOver] = useState(false);

  const startDiscover = useCallback(async () => {
    setDiscovering(true);
    try {
      const result = await DiscoverDevices(3);
      if (Array.isArray(result)) {
        setDevices(prev => {
          const existing = new Map<string, DeviceItem>();
          prev.forEach(d => existing.set(d.ip + ":" + d.port, d));
          result.forEach((d: any) => {
            const key = d.ip + ":" + d.port;
            existing.set(key, {
              name: d.name,
              ip: d.ip,
              port: d.port,
              online: true,
              source: "discovered",
            });
          });
          return Array.from(existing.values());
        });
      }
    } catch {}
    setDiscovering(false);
  }, []);

  useEffect(() => {
    startDiscover();
  }, [startDiscover]);

  useEffect(() => {
    OnFileDrop((_x, _y, files: string[]) => {
      if (files?.length) setPaths(prev => [...prev, ...files]);
    }, true);
    return () => OnFileDropOff();
  }, []);

  const handleSelectFiles = useCallback(async () => {
    try { const r = await OpenFileDialog(); if (r?.length) setPaths(prev => [...prev, ...r]); } catch {}
  }, []);

  const handleSelectDir = useCallback(async () => {
    try { const r = await OpenDirDialog(); if (r) setPaths(prev => [...prev, r]); } catch {}
  }, []);

  const handleAddManual = useCallback(() => {
    if (!manualIP.trim()) return;
    const port = parseInt(manualPort) || 9999;
    const key = manualIP.trim() + ":" + port;
    setDevices(prev => {
      if (prev.some(d => d.ip + ":" + d.port === key)) return prev;
      return [...prev, { name: manualIP.trim(), ip: manualIP.trim(), port, online: false, source: "manual" }];
    });
    setSelected({ name: manualIP.trim(), ip: manualIP.trim(), port, online: false, source: "manual" });
    setShowManual(false);
    setManualIP("");
  }, [manualIP, manualPort]);

  const handleClearDevices = useCallback(() => {
    setDevices(prev => prev.filter(d => d.source === "discovered" && d.online));
    setSelected(null);
  }, []);

  const handleSend = useCallback(async () => {
    if (!selected) { setError(t("device.error_select_target")); return; }
    if (!paths.length) { setError(t("device.error_select_files")); return; }
    setError("");
    try { setTaskID(await StartSend(selected.ip, selected.port, paths, selected.name)); } catch (e: any) { setError(e?.toString() || t("send.title")); }
  }, [selected, paths, t]);

  return (
    <div style={overlayStyle} onClick={isActive ? undefined : onClose}>
      <div style={modalStyle} onClick={e => e.stopPropagation()}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
          <span style={{ fontSize: 18, fontWeight: 600, color: "var(--text-primary)" }}>{t("send.title")}</span>
          <button onClick={onClose} style={closeBtn}>✕</button>
        </div>

        {!isActive ? (
          <>
            <div
              style={{
                padding: paths.length > 0 ? 12 : 32, textAlign: "center", borderRadius: 12, marginBottom: 12,
                border: `2px dashed ${dragOver ? "var(--accent)" : "var(--border-input)"}`,
                background: dragOver ? "rgba(107,140,255,0.06)" : "var(--bg-input)",
                transition: "all 0.2s", "--wails-drop-target": "drop",
              } as React.CSSProperties}
              onDragOver={e => { e.preventDefault(); setDragOver(true); }}
              onDragLeave={() => setDragOver(false)}
              onDrop={e => { e.preventDefault(); setDragOver(false); }}
            >
              <p style={{ fontSize: 14, color: "var(--text-secondary)", margin: 0 }}>
                {paths.length > 0 ? t("send.files_selected", { count: paths.length }) : t("send.drag_placeholder")}
              </p>
              <div style={{ display: "flex", gap: 8, justifyContent: "center", marginTop: 10, flexWrap: "nowrap" }}>
                <button onClick={handleSelectFiles} style={{ ...smBtn, whiteSpace: "nowrap" }}>{t("button.select_file")}</button>
                <button onClick={handleSelectDir} style={{ ...smBtn, whiteSpace: "nowrap" }}>{t("button.select_folder")}</button>
              </div>
            </div>

            {paths.length > 0 && (
              <div style={fileListStyle}>
                {paths.map((p, i) => (
                  <div key={i} style={{ display: "flex", justifyContent: "space-between", fontSize: 13, color: "var(--text-primary)", padding: "3px 0" }}>
                    <span style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", flex: 1 }}>{p}</span>
                    <button onClick={() => setPaths(prev => prev.filter((_, j) => j !== i))} style={rmBtn}>✕</button>
                  </div>
                ))}
              </div>
            )}

            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
              <span style={{ fontSize: 13, color: "var(--text-secondary)" }}>{t("device.select_device")}</span>
              <div style={{ display: "flex", gap: 6 }}>
                <button onClick={startDiscover} disabled={discovering} style={textBtn}>
                  {discovering ? t("device.searching") : t("button.refresh")}
                </button>
                <button onClick={handleClearDevices} style={textBtn}>{t("button.clear")}</button>
              </div>
            </div>

            <div style={{ maxHeight: 160, overflow: "auto", marginBottom: 4 }}>
              {devices.map((d, i) => {
                const key = d.ip + ":" + d.port;
                const isSel = selected?.ip + ":" + selected?.port === key;
                return (
                  <button
                    key={i}
                    onClick={() => setSelected(d)}
                    style={{
                      display: "flex", alignItems: "center", gap: 8, width: "100%",
                      padding: "10px 12px", marginBottom: 4,
                      background: isSel ? "var(--bg-hover)" : "transparent",
                      border: `1px solid ${isSel ? "var(--accent)" : "var(--border)"}`,
                      borderRadius: 8, color: "var(--text-primary)", fontSize: 13, cursor: "pointer", textAlign: "left",
                    }}
                  >
                    <span style={{ fontSize: 10, flexShrink: 0 }}>{d.online ? "🟢" : "⚪"}</span>
                    <span style={{ flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{d.name}</span>
                    <span style={{ color: "var(--text-muted)", fontSize: 12, flexShrink: 0 }}>{d.ip}:{d.port}</span>
                  </button>
                );
              })}
              {devices.length === 0 && (
                <p style={{ color: "var(--text-muted)", fontSize: 13, textAlign: "center", padding: 12 }}>
                  {discovering ? t("device.searching") : t("device.no_device")}
                </p>
              )}
            </div>

            {showManual ? (
              <div style={{ display: "flex", gap: 6, marginTop: 4, alignItems: "stretch" }}>
                <input value={manualIP} onChange={e => setManualIP(e.target.value)} placeholder={t("device.ip_placeholder")} style={inputStyle} />
                <input value={manualPort} onChange={e => setManualPort(e.target.value)} placeholder={t("device.port_placeholder")} style={{ ...inputStyle, width: 70 }} />
                <button onClick={handleAddManual} style={{ ...smBtn, whiteSpace: "nowrap" }}>{t("button.add")}</button>
                <button onClick={() => setShowManual(false)} style={{ ...smBtn, color: "var(--text-muted)", whiteSpace: "nowrap" }}>{t("button.cancel_small")}</button>
              </div>
            ) : (
              <button onClick={() => setShowManual(true)} style={{ ...textBtn, width: "100%", textAlign: "center", marginTop: 4 }}>
                {t("button.manual_add")}
              </button>
            )}

            {error && <p style={{ color: "var(--danger)", fontSize: 13, marginTop: 10 }}>{error}</p>}
            <button onClick={handleSend} className="save-btn" style={primaryBtn}>{t("button.start_send")}</button>
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
        {view.state === "running" ? t("send.progress.transferring") :
         view.state === "finalizing" ? t("send.progress.finalizing") :
         view.state === "completed" ? t("send.progress.completed") :
         view.state === "failed" ? t("send.progress.failed") :
         view.state === "cancelled" ? t("send.progress.cancelled") : t("send.progress.waiting")}
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
const modalStyle: React.CSSProperties = { width: "min(440px, 94vw)", maxHeight: "88vh", overflow: "auto", background: "var(--modal-bg)", border: "1px solid var(--border)", borderRadius: 16, padding: 24 };
const closeBtn: React.CSSProperties = { background: "none", border: "none", color: "var(--text-secondary)", fontSize: 18, cursor: "pointer" };
const inputStyle: React.CSSProperties = { flex: 1, padding: "8px 12px", background: "var(--bg-input)", border: "1px solid var(--border-input)", borderRadius: 8, color: "var(--text-primary)", fontSize: 14, outline: "none" };
const smBtn: React.CSSProperties = { padding: "6px 14px", background: "var(--bg-input)", border: "1px solid var(--border-input)", borderRadius: 8, color: "var(--text-primary)", fontSize: 13, cursor: "pointer" };
const textBtn: React.CSSProperties = { padding: "4px 8px", background: "none", border: "none", color: "var(--text-secondary)", fontSize: 12, cursor: "pointer" };
const primaryBtn: React.CSSProperties = { width: "100%", marginTop: 16, padding: "12px", background: "linear-gradient(135deg, var(--accent), var(--accent2))", border: "none", borderRadius: 12, color: "#fff", fontSize: 15, fontWeight: 600, cursor: "pointer" };
const fileListStyle: React.CSSProperties = { padding: "8px 12px", background: "var(--bg-input)", border: "1px solid var(--border)", borderRadius: 8, maxHeight: 100, overflow: "auto" };
const rmBtn: React.CSSProperties = { background: "none", border: "none", color: "var(--danger)", fontSize: 14, cursor: "pointer" };