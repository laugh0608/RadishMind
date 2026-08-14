import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import "@xyflow/react/dist/style.css";
import "./styles/family-ui/tokens.css";
import "./styles/radishmind-aliases.css";
import "./styles.css";

createRoot(document.getElementById("root") as HTMLElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
