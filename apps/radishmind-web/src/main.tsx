import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import { uiI18n, initializeUiI18n } from "./i18n/instance.ts";
import { createLocalePreferenceStore } from "./i18n/localePreference.ts";
import { LocaleProvider } from "./i18n/LocaleProvider.tsx";
import "@xyflow/react/dist/style.css";
import "./styles/family-ui/tokens.css";
import "./styles/radishmind-aliases.css";
import "./styles.css";

const root = createRoot(document.getElementById("root") as HTMLElement);
async function start() {
  const instance = uiI18n;
  const preference = createLocalePreferenceStore(() => window.localStorage, navigator.languages, async locale => {
    await instance.changeLanguage(locale);
    document.documentElement.lang = locale;
  });
  const { locale } = preference.getSnapshot();
  await initializeUiI18n(instance, locale);
  document.documentElement.lang = locale;
  root.render(<StrictMode><LocaleProvider instance={instance} preference={preference}><App /></LocaleProvider></StrictMode>);
}
void start().catch(() => {
  root.render(<main role="alert"><h1>界面初始化失败 / Interface initialization failed</h1>
    <p>请刷新重试。 / Reload to try again.</p><button type="button" onClick={() => window.location.reload()}>重试 / Retry</button></main>);
});
