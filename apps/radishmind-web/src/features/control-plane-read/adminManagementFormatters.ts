import type { UiLocale } from "../../i18n/localePreference.ts";
import { formatDisplayDate, formatDisplayNumber } from "../../i18n/formatters.ts";

export function adminDisplayDate(value: string, locale: UiLocale): string {
  const date = formatDisplayDate(value, locale, { timeZone: "UTC" });
  return date ? `${date} UTC` : value;
}

// Rates remain exact micro-USD integers in forms and requests. Displaying a rate
// must not round away a micro-unit through floating-point currency conversion.
export function adminDisplayRate(value: number, locale: UiLocale): string {
  const micros = BigInt(value);
  return `USD ${formatDisplayNumber(micros / 1_000_000n, locale)}.${String(micros % 1_000_000n).padStart(6, "0")}`;
}
