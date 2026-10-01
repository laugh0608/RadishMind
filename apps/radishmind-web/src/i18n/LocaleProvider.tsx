import { createContext, useContext, useEffect, useSyncExternalStore, type ReactNode } from "react";
import { I18nextProvider } from "react-i18next";
import type { i18n } from "i18next";
import { createLocalePreferenceStore, UI_LOCALE_KEY } from "./localePreference.ts";

type PreferenceStore = ReturnType<typeof createLocalePreferenceStore>;
const LocaleContext = createContext<PreferenceStore | null>(null);

export function LocaleProvider({ instance, preference, children }: {
  instance: i18n; preference: PreferenceStore; children: ReactNode;
}) {
  useEffect(() => {
    function onStorage(event: StorageEvent) {
      try { if (event.storageArea !== window.localStorage) return; } catch { return; }
      if (event.key !== UI_LOCALE_KEY && event.key !== null) return;
      void preference.acceptStorageChange(event.key, event.newValue).catch(() => {
        // No remote loader is configured; only a resource initialization defect can fail here.
        console.error("ui_locale_storage_change_failed");
      });
    }
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, [preference]);
  return <LocaleContext.Provider value={preference}><I18nextProvider i18n={instance}>{children}</I18nextProvider></LocaleContext.Provider>;
}

export function useLocalePreference() {
  const store = useContext(LocaleContext);
  if (!store) throw new Error("LocaleProvider is required");
  const snapshot = useSyncExternalStore(store.subscribe, store.getSnapshot, store.getSnapshot);
  return { ...snapshot, select: store.select };
}
