import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./App.css";
import "../wailsjs/runtime/runtime";
import { LanguageProvider } from "./i18n/i18n";

const container = document.getElementById("root");
const root = createRoot(container!);
root.render(
  <React.StrictMode>
    <LanguageProvider>
      <App />
    </LanguageProvider>
  </React.StrictMode>
);