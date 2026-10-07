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

// Keep six decimal places from the integer micro-USD contract without a
// floating-point division. Currency is fixed by the existing Gateway schema.
export function formatMicroUSD(value: number | bigint, locale: UiLocale): string | null {
  if (typeof value === "number" && (!Number.isSafeInteger(value) || value < 0)) return null;
  const micros = BigInt(value);
  if (micros < 0n) return null;
  return `USD ${formatDisplayNumber(micros / 1_000_000n, locale)}.${String(micros % 1_000_000n).padStart(6, "0")}`;
}
