import type { UiLocale } from "./localePreference.ts";

export function formatDisplayDate(value: string | null | undefined, locale: UiLocale, options: Pick<Intl.DateTimeFormatOptions, "dateStyle" | "timeStyle" | "timeZone"> = {}): string | null {
  if (!value) return null;
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return null;
  return new Intl.DateTimeFormat(locale, { dateStyle: "medium", timeStyle: "short", ...options }).format(date);
}

export function formatDisplayNumber(value: number | bigint | null | undefined, locale: UiLocale, options: Intl.NumberFormatOptions = {}): string | null {
  if (value === null || value === undefined || (typeof value === "number" && !Number.isFinite(value))) return null;
  return new Intl.NumberFormat(locale, options).format(value);
}
