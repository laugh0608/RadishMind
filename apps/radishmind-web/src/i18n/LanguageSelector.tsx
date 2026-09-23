import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { isUiLocale } from "./localePreference.ts";
import { useLocalePreference } from "./LocaleProvider.tsx";
import "./languageSelector.css";

export function LanguageSelector() {
  const { t } = useTranslation("common");
  const { locale, saved, select } = useLocalePreference();
  const [failed, setFailed] = useState(false);
  const id = useId();
  return <div className="ui-language-selector">
    <label htmlFor={id}>{t($ => $.language)}</label>
    <select id={id} value={locale} onChange={event => {
      const next = event.target.value;
      if (!isUiLocale(next)) return;
      setFailed(false);
      void select(next).catch(() => setFailed(true));
    }}>
      <option value="zh-CN">简体中文</option><option value="en-US">English</option>
    </select>
    <small role="status">{failed ? t($ => $.languageChangeFailed) : !saved ? t($ => $.preferenceNotSaved) : ""}</small>
  </div>;
}
