export const UI_LOCALE_KEY = "radishmind.uiLocale.v1";
export const UI_LOCALES = ["zh-CN", "en-US"] as const;
export type UiLocale = typeof UI_LOCALES[number];
export type LocalePreference = { locale: UiLocale; saved: boolean };
export type LocaleStorage = Pick<Storage, "getItem" | "setItem">;

export function isUiLocale(value: unknown): value is UiLocale {
  return value === "zh-CN" || value === "en-US";
}

export function resolveUiLocale(saved: unknown, languages: readonly string[]): UiLocale {
  if (isUiLocale(saved)) return saved;
  for (const language of languages) {
    if (/^zh(?:-|$)/i.test(language)) return "zh-CN";
    if (/^en(?:-|$)/i.test(language)) return "en-US";
  }
  return "zh-CN";
}

export function readLocalePreference(storage: () => LocaleStorage, languages: readonly string[]): LocalePreference {
  try {
    return { locale: resolveUiLocale(storage().getItem(UI_LOCALE_KEY), languages), saved: true };
  } catch {
    return { locale: resolveUiLocale(null, languages), saved: false };
  }
}

/** Stores only the display locale; never receives business data or identity events. */
export function createLocalePreferenceStore(
  storage: () => LocaleStorage,
  languages: readonly string[],
  applyLocale: (locale: UiLocale) => Promise<void>,
) {
  let snapshot = readLocalePreference(storage, languages);
  let revision = 0;
  const listeners = new Set<() => void>();
  function publish(next: LocalePreference) {
    if (next.locale === snapshot.locale && next.saved === snapshot.saved) return;
    snapshot = next;
    for (const listener of listeners) listener();
  }
  async function update(locale: UiLocale, persist: boolean) {
    const current = ++revision;
    await applyLocale(locale);
    if (current !== revision) return;
    let saved = true;
    if (persist) {
      try { storage().setItem(UI_LOCALE_KEY, locale); } catch { saved = false; }
    }
    publish({ locale, saved });
  }
  return {
    getSnapshot: () => snapshot,
    subscribe(listener: () => void) { listeners.add(listener); return () => { listeners.delete(listener); }; },
    select: (locale: UiLocale) => update(locale, true),
    acceptStorageChange: (key: string | null, value: string | null) => {
      if (key !== UI_LOCALE_KEY && key !== null) return Promise.resolve();
      if (value !== null && !isUiLocale(value)) return Promise.resolve();
      const locale = resolveUiLocale(value, languages);
      if (locale === snapshot.locale) {
        publish({ locale, saved: true });
        return Promise.resolve();
      }
      return update(locale, false);
    },
  };
}
