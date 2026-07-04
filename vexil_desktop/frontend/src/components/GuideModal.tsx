import { useLanguage } from "../i18n/i18n";

interface Props {
  onClose: () => void;
}

export function GuideModal({ onClose }: Props) {
  const { t } = useLanguage();
  return (
    <div style={{
      position: "fixed", inset: 0, background: "var(--overlay)",
      display: "flex", alignItems: "center", justifyContent: "center", zIndex: 200,
    }} onClick={onClose}>
      <div style={{
        width: "min(360px, 90vw)", background: "var(--modal-bg)",
        border: "1px solid var(--border)", borderRadius: 16, padding: 28,
      }} onClick={e => e.stopPropagation()}>
        <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--text-primary)", textAlign: "center", marginBottom: 20 }}>
          {t("guide.welcome")}
        </h2>
        <div style={{ color: "var(--text-secondary)", fontSize: 14, lineHeight: 2 }}>
          <div>{t("guide.send_desc")}</div>
          <div>{t("guide.recv_desc")}</div>
          <div>{t("guide.resume")}</div>
          <div>{t("guide.checksum")}</div>
          <div>{t("guide.theme")}</div>
          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" style={{ flexShrink: 0 }}>
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            {t("guide.settings")}
          </div>
        </div>
        <button
          onClick={onClose}
          style={{
            width: "100%", marginTop: 20, padding: 12,
            background: "linear-gradient(135deg, var(--accent), var(--accent2))",
            border: "none", borderRadius: 12, color: "#fff",
            fontSize: 15, fontWeight: 600, cursor: "pointer",
          }}
        >
          {t("button.got_it")}
        </button>
      </div>
    </div>
  );
}