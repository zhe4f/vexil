import { useState, useEffect, useCallback } from "react";
import { GetConfig, UpdateConfig, OpenDirDialog } from "../../wailsjs/go/gui/Handler";
import { useLanguage } from "../i18n/i18n";

interface Props {
  onClose: () => void;
}

export function SettingsDrawerContent({ onClose }: Props) {
  const { t, lang, switchLang } = useLanguage();
  const [port, setPort] = useState("9999");
  const [saveDir, setSaveDir] = useState("./downloads");
  const [numConns, setNumConns] = useState("8");
  const [maxChunkMB, setMaxChunkMB] = useState("16");
  const [windowSizeMB, setWindowSizeMB] = useState("256");
  const [tlsEnabled, setTlsEnabled] = useState(true);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    const s = localStorage.getItem("vexil_settings");
    if (s) {
      try {
        const p = JSON.parse(s);
        if (p.port) setPort(p.port);
        if (p.saveDir) setSaveDir(p.saveDir);
      } catch {}
    }
    GetConfig()
      .then(c => {
        if (c) {
          setNumConns(String(c.num_conns || 8));
          setMaxChunkMB(String(c.max_chunk_mb || 16));
          setWindowSizeMB(String(c.window_size_mb || 256));
          if (c.tls_enabled !== undefined) setTlsEnabled(c.tls_enabled);
        }
      })
      .catch(() => {});
  }, []);

  const handleSave = useCallback(async () => {
    localStorage.setItem("vexil_settings", JSON.stringify({ port, saveDir }));
    try {
      await UpdateConfig({
        num_conns: parseInt(numConns) || 8,
        max_chunk_mb: parseInt(maxChunkMB) || 16,
        window_size_mb: parseInt(windowSizeMB) || 256,
        tls_enabled: tlsEnabled,
      });
    } catch {}
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  }, [port, saveDir, numConns, maxChunkMB, windowSizeMB, tlsEnabled]);

  return (
    <>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <span style={{ fontSize: 18, fontWeight: 600, color: "var(--text-primary)" }}>{t("settings.title")}</span>
        <button onClick={onClose} style={{ background: "none", border: "none", color: "var(--text-secondary)", fontSize: 18, cursor: "pointer" }}>✕</button>
      </div>

      <label style={lbl}>{t("settings.port")}</label>
      <input value={port} onChange={e => setPort(e.target.value)} style={inp} />

      <label style={{ ...lbl, marginTop: 16 }}>{t("settings.save_dir")}</label>
      <div style={{ display: "flex", gap: 8 }}>
        <input value={saveDir} onChange={e => setSaveDir(e.target.value)} style={inp} />
        <button onClick={async () => { try { const r = await OpenDirDialog(); if (r) setSaveDir(r); } catch {} }} style={iconBtn}>📂</button>
      </div>

      <div style={{ marginTop: 24, paddingTop: 16, borderTop: "1px solid var(--border)" }}>
        <p style={{ fontSize: 12, color: "var(--text-muted)", marginBottom: 12 }}>{t("settings.advanced")}</p>
        <label style={lbl}>{t("settings.conns")}</label>
        <input value={numConns} onChange={e => setNumConns(e.target.value)} style={inp} />
        <label style={{ ...lbl, marginTop: 12 }}>{t("settings.chunk_mb")}</label>
        <input value={maxChunkMB} onChange={e => setMaxChunkMB(e.target.value)} style={inp} />
        <label style={{ ...lbl, marginTop: 12 }}>{t("settings.window_mb")}</label>
        <input value={windowSizeMB} onChange={e => setWindowSizeMB(e.target.value)} style={inp} />

        <label style={{ ...lbl, marginTop: 12 }}>{t("settings.tls")}</label>
        <div style={{ display: "flex", alignItems: "center", gap: 10, marginTop: 4 }}>
          <button
            onClick={() => setTlsEnabled(!tlsEnabled)}
            style={{
              width: 44,
              height: 24,
              borderRadius: 12,
              background: tlsEnabled ? "var(--success)" : "var(--border)",
              border: "none",
              cursor: "pointer",
              position: "relative",
              transition: "background 0.2s",
            }}
          >
            <span
              style={{
                position: "absolute",
                top: 2,
                left: tlsEnabled ? 22 : 2,
                width: 20,
                height: 20,
                borderRadius: "50%",
                background: "#fff",
                transition: "left 0.2s",
              }}
            />
          </button>
          <span style={{ fontSize: 13, color: "var(--text-secondary)" }}>
            {tlsEnabled ? t("settings.tls_enabled") : t("settings.tls_disabled")}
          </span>
        </div>

        <label style={{ ...lbl, marginTop: 16 }}>{t("settings.language")}</label>
        <select
          value={lang}
          onChange={(e) => switchLang(e.target.value as "zh" | "en")}
          style={{ ...inp, marginTop: 4 }}
        >
          <option value="zh">{t("settings.lang_zh")}</option>
          <option value="en">{t("settings.lang_en")}</option>
        </select>
      </div>

      <button
        onClick={handleSave}
        className="save-btn"
        style={{
          width: "100%",
          marginTop: 24,
          padding: 12,
          background: saved ? "var(--success)" : "linear-gradient(135deg, var(--accent), var(--accent2))",
          border: "none",
          borderRadius: 12,
          color: "#fff",
          fontSize: 15,
          fontWeight: 600,
          cursor: "pointer",
        }}
      >
        {saved ? t("button.saved") : t("button.save")}
      </button>

      <div style={{ marginTop: 24, paddingTop: 16, borderTop: '1px solid var(--border)', textAlign: 'center' }}>
        <p style={{ fontSize: 13, color: 'var(--text-muted)', lineHeight: 1.8 }}>
          Vexil v1.0.0<br/>
          Made with ❤️ by <a href="https://github.com/zhe4f" target="_blank" style={{ color: 'var(--accent)' }}>zhe4f</a><br/>
          <span style={{ fontSize: 12 }}>MIT License</span>
        </p>
      </div>
    </>
  );
}

const lbl: React.CSSProperties = { fontSize: 13, color: "var(--text-secondary)", display: "block" };
const inp: React.CSSProperties = { width: "100%", padding: "10px 14px", marginTop: 4, background: "var(--bg-input)", border: "1px solid var(--border-input)", borderRadius: 10, color: "var(--text-primary)", fontSize: 14, outline: "none", boxSizing: "border-box" };
const iconBtn: React.CSSProperties = { width: 44, padding: 0, background: "var(--bg-input)", border: "1px solid var(--border-input)", borderRadius: 10, color: "var(--text-primary)", fontSize: 18, cursor: "pointer" };