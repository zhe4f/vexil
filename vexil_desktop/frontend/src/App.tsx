import { useState, useEffect, useCallback } from "react";
import { GetHistory, DeleteHistory, ClearHistory } from "../wailsjs/go/gui/Handler";
import { SendModal } from "./components/SendModal";
import { RecvModal } from "./components/RecvModal";
import { HistoryModal } from "./components/HistoryModal";
import { HistoryList } from "./components/HistoryList";
import { SettingsDrawerContent } from "./components/SettingsDrawer";
import { useTransfer } from "./hooks/useTransfer";
import type { HistoryEntry } from "./types";
import { AnimatePresence, motion } from "framer-motion";
import { GuideModal } from "./components/GuideModal";
import "./App.css";
import type { DeviceInfo } from "./types";
import { GetDeviceInfo, SetDeviceName } from "../wailsjs/go/gui/Handler";
import { useLanguage } from "./i18n/i18n";

function App() {
  const { t, lang, switchLang } = useLanguage();
  const [activeTab, setActiveTab] = useState<"send" | "recv" | null>(null);
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [showSettings, setShowSettings] = useState(false);
  const [showHistoryAll, setShowHistoryAll] = useState(false);
  const [selectedEntry, setSelectedEntry] = useState<HistoryEntry | null>(null);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [theme, setTheme] = useState<"light" | "dark">(() => {
    return (localStorage.getItem("vexil_theme") as "dark" | "light") || "dark";
  });
  const [curtain, setCurtain] = useState(false);
  const { view, isActive } = useTransfer();

  const [editingName, setEditingName] = useState(false);
  const [newName, setNewName] = useState("");

  const handleSaveName = async () => {
    if (newName.trim()) {
      await SetDeviceName(newName.trim());
      setDeviceInfo(prev => prev ? { ...prev, name: newName.trim() } : null);
    }
    setEditingName(false);
  };

  const [deviceInfo, setDeviceInfo] = useState<DeviceInfo | null>(null);
  useEffect(() => {
    GetDeviceInfo().then(setDeviceInfo);
  }, []);

  const [showGuide, setShowGuide] = useState(() => {
    return localStorage.getItem("vexil_guide_shown") !== "1";
  });

  const closeGuide = () => {
    setShowGuide(false);
    localStorage.setItem("vexil_guide_shown", "1");
  };

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
    localStorage.setItem("vexil_theme", theme);
  }, [theme]);

  const switchTheme = () => {
    setCurtain(true);
    setTimeout(() => {
      document.documentElement.classList.add("theme-switching");
      setTheme(t => t === "dark" ? "light" : "dark");
    }, 400);
    setTimeout(() => {
      document.documentElement.classList.remove("theme-switching");
    }, 900);
    setTimeout(() => {
      setCurtain(false);
    }, 700);
  };

  const loadHistory = useCallback(async () => {
    try { const e = await GetHistory(20); if (Array.isArray(e)) setHistory(e); } catch {}
  }, []);

  useEffect(() => { loadHistory(); }, [loadHistory]);
  useEffect(() => { if (!isActive) { loadHistory(); } }, [isActive, loadHistory]);

  const handleDeleteAll = useCallback(async () => {
    try { await ClearHistory(); setHistory([]); } catch {}
  }, []);

  const themeTitle = theme === "dark" ? t("theme.switch_light") : t("theme.switch_dark");
  const settingsTitle = t("settings.title");

  return (
    <div id="app">
      <div className="bg-glow bg-glow-1" />
      <div className="bg-glow bg-glow-2" />

      <button
        onClick={switchTheme}
        style={{
          position: "fixed", top: 20, right: 68, zIndex: 10,
          width: 40, height: 40, borderRadius: 10,
          background: "var(--bg-input)", border: "1px solid var(--border-input)",
          color: "var(--text-secondary)", fontSize: 18, cursor: "pointer",
        }}
        title={themeTitle}
      >
        {theme === "dark" ? "☀" : "🌙"}
      </button>

      <button
        onClick={() => setShowSettings(true)}
        style={{
          position: "fixed", top: 20, right: 20, zIndex: 10,
          width: 40, height: 40, borderRadius: 10,
          background: "var(--bg-input)", border: "1px solid var(--border-input)",
          color: "var(--text-secondary)", cursor: "pointer",
          display: "flex", alignItems: "center", justifyContent: "center",
        }}
        title={settingsTitle}
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
          <circle cx="12" cy="12" r="3"/>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
        </svg>
      </button>

      <div className="logo">
        <h1 className="logo-text">{t("app.title")}</h1>
        <div className="logo-line" />
      </div>

      <div style={{ position: "fixed", bottom: 20, right: 20, zIndex: 10, textAlign: "right" }}>
        {editingName ? (
          <div style={{ display: "flex", alignItems: "center", gap: 4 }}>
            <input
              value={newName}
              onChange={e => setNewName(e.target.value)}
              onKeyDown={e => e.key === "Enter" && handleSaveName()}
              placeholder={t("device.select_device")}
              autoFocus
              style={{
                width: 120,
                padding: "4px 8px",
                background: "var(--bg-input)",
                border: "1px solid var(--border-input)",
                borderRadius: 6,
                color: "var(--text-primary)",
                fontSize: 14,
                textAlign: "right",
              }}
            />
            <button onClick={handleSaveName} style={{
              background: "none",
              border: "none",
              color: "var(--accent)",
              cursor: "pointer",
              fontSize: 14,
            }}>✓</button>
            <button onClick={() => setEditingName(false)} style={{
              background: "none",
              border: "none",
              color: "var(--text-muted)",
              cursor: "pointer",
              fontSize: 14,
            }}>✕</button>
          </div>
        ) : (
          <div
            onClick={() => {
              setNewName(deviceInfo?.name ?? "");
              setEditingName(true);
            }}
            style={{ cursor: "pointer" }}
          >
            <div style={{ fontSize: 14, color: "var(--text-secondary)", lineHeight: 1.8 }}>
              {deviceInfo?.name ?? "..."}
            </div>
            <div style={{ fontSize: 14, color: "var(--text-secondary)", lineHeight: 1.8 }}>
              {deviceInfo?.ip ?? "..."}
            </div>
          </div>
        )}
      </div>

      {isActive && view ? (
        <div style={{ width: "min(340px, 85vw)", textAlign: "center", marginTop: 20 }}>
          <p style={{ fontSize: 15, color: "var(--text-secondary)", marginBottom: 16 }}>
            {view.state === "preparing" ? t("progress.preparing") :
             view.state === "connecting" ? t("progress.connecting") :
             view.state === "running" ? t("progress.running") :
             view.state === "finalizing" ? t("progress.finalizing") :
             view.state === "completed" ? t("progress.completed") :
             view.state === "failed" ? t("progress.failed") : t("progress.cancelled")}
          </p>
          <div style={{ height: 6, background: "var(--bg-input)", borderRadius: 3, overflow: "hidden", marginBottom: 12 }}>
            <div style={{ width: `${view.percent}%`, height: "100%", background: view.state === "completed" ? "var(--success)" : "linear-gradient(90deg, var(--accent), var(--accent2))", borderRadius: 3, transition: "width 0.3s" }} />
          </div>
          <div style={{ display: "flex", justifyContent: "space-between", fontSize: 13, color: "var(--text-secondary)", marginBottom: 4 }}>
            <span>{view.percent.toFixed(1)}%</span>
            <span>{view.sent} / {view.total}</span>
          </div>
          <div style={{ fontSize: 13, color: "var(--text-muted)" }}>{view.speed} · {view.eta}</div>
          {view.error && <p style={{ color: "var(--danger)", fontSize: 13, marginTop: 8 }}>{view.error}</p>}
        </div>
      ) : (
        <>
          <div className="main-cards">
            <button className="card card-send" onClick={() => setActiveTab("send")}>
            <div className="card-icon">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M12 19V5M5 12l7-7 7 7"/></svg>
            </div>
            <span className="card-title">{t("send.title")}</span>
            </button>
            <button className="card card-recv" onClick={() => setActiveTab("recv")}>
            <div className="card-icon">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M12 5v14M5 12l7 7 7-7"/></svg>
            </div>
            <span className="card-title">{t("recv.title")}</span>
            </button>
          </div>

          <div className="recent-section">
            <div className="divider" />
            <HistoryList entries={history} max={3} onOpen={(entry) => {
              const idx = history.indexOf(entry);
              setSelectedEntry(entry);
              setSelectedIndex(idx >= 0 ? idx + 1 : 1);
            }} />
            {history.length > 3 && (
              <button onClick={() => setShowHistoryAll(true)} style={{ marginTop: 25, padding: "6px 16px", background: "none", border: "1px solid var(--border-input)", borderRadius: 8, color: "var(--text-secondary)", fontSize: 12, cursor: "pointer" }}>
                {t("button.view_all")} ({history.length})
              </button>
            )}
          </div>
        </>
      )}

      {activeTab === "send" && <SendModal onClose={() => setActiveTab(null)} view={view} isActive={isActive} historyPeers={history.map(e => e.peer)} />}
      {activeTab === "recv" && <RecvModal onClose={() => setActiveTab(null)} view={view} isActive={isActive} />}
      {selectedEntry && <HistoryModal entry={selectedEntry} index={selectedIndex} onClose={() => setSelectedEntry(null)} onDeleted={loadHistory} />}

      {showHistoryAll && (
        <div style={{ position: "fixed", inset: 0, background: "var(--overlay)", zIndex: 100, display: "flex", alignItems: "center", justifyContent: "center" }} onClick={() => setShowHistoryAll(false)}>
          <div style={{ width: "min(420px, 92vw)", maxHeight: "80vh", background: "var(--modal-bg)", border: "1px solid var(--border)", borderRadius: 16, display: "flex", flexDirection: "column", overflow: "hidden" }} onClick={e => e.stopPropagation()}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "20px 24px 12px", flexShrink: 0 }}>
              <span style={{ fontSize: 18, fontWeight: 600, color: "var(--text-primary)" }}>{t("history.title")}</span>
              <button onClick={() => setShowHistoryAll(false)} style={{ background: "none", border: "none", color: "var(--text-secondary)", fontSize: 18, cursor: "pointer" }}>✕</button>
            </div>

            <div style={{ flex: 1, overflow: "auto", padding: "0 24px" }}>
              <HistoryList
                entries={history}
                onDelete={async (i) => { try { await DeleteHistory(i); loadHistory(); } catch {} }}
                onOpen={(entry) => {
                  const idx = history.indexOf(entry);
                  setSelectedEntry(entry);
                  setSelectedIndex(idx >= 0 ? idx + 1 : 1);
                }}
              />
            </div>

            {history.length > 0 && (
              <div style={{ padding: "12px 24px 20px", flexShrink: 0 }}>
                <button onClick={handleDeleteAll} style={{ padding: "8px", width: "100%", background: "rgba(248,81,73,0.1)", border: "1px solid rgba(248,81,73,0.2)", borderRadius: 8, color: "var(--danger)", fontSize: 13, cursor: "pointer" }}>{t("button.delete_all")}</button>
              </div>
            )}
          </div>
        </div>
      )}

      <AnimatePresence>
        {showSettings && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              style={{ position: "fixed", inset: 0, background: "var(--overlay)", zIndex: 99 }}
              onClick={() => setShowSettings(false)}
            />
            <motion.div
              initial={{ x: "100%" }}
              animate={{ x: 0 }}
              exit={{ x: "100%" }}
              transition={{ duration: 0.3, ease: [0.32, 0.72, 0, 1] }}
              style={{ position: "fixed", top: 0, right: 0, bottom: 0, width: "min(320px, 85vw)", background: "var(--modal-bg)", borderLeft: "1px solid var(--border)", zIndex: 100, padding: 24, overflow: "auto" }}
            >
              <SettingsDrawerContent onClose={() => setShowSettings(false)} />
            </motion.div>
          </>
        )}
      </AnimatePresence>

      <div style={{
        position: "fixed", inset: 0, zIndex: 9999,
        background: `linear-gradient(180deg, var(--bg-primary), var(--bg-secondary))`,
        opacity: curtain ? 1 : 0,
        transform: curtain ? "translateY(0)" : "translateY(-100%)",
        transition: "transform 0.5s cubic-bezier(0.76, 0, 0.24, 1), opacity 0.3s ease",
        pointerEvents: "none",
      }} />

      {showGuide && <GuideModal onClose={closeGuide} />}
    </div>
  );
}

export default App;