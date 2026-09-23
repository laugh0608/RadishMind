/// <reference path="./resources.d.ts" />
import { createInstance } from "i18next";
import { common as zhCommon } from "./locales/zh-CN/common.ts";
import { common as enCommon } from "./locales/en-US/common.ts";
import { shell as zhShell } from "./locales/zh-CN/shell.ts";
import { shell as enShell } from "./locales/en-US/shell.ts";
import { identity as zhIdentity } from "./locales/zh-CN/identity.ts";
import { identity as enIdentity } from "./locales/en-US/identity.ts";
import { appShell as zhAppShell } from "./locales/zh-CN/appShell.ts";
import { appShell as enAppShell } from "./locales/en-US/appShell.ts";
import type { UiLocale } from "./localePreference.ts";

export function createUiI18n() {
  return createInstance();
}

export async function initializeUiI18n(instance: ReturnType<typeof createUiI18n>, locale: UiLocale) {
  await instance.init({
    lng: locale,
    supportedLngs: ["zh-CN", "en-US"],
    load: "currentOnly",
    fallbackLng: "zh-CN",
    defaultNS: "common",
    ns: ["common", "shell", "identity"],
    resources: { "zh-CN": { common: zhCommon, shell: { ...zhShell, appShell: zhAppShell }, identity: zhIdentity }, "en-US": { common: enCommon, shell: { ...enShell, appShell: enAppShell }, identity: enIdentity } },
    initAsync: false,
    returnNull: false,
    returnEmptyString: false,
    interpolation: { escapeValue: false },
    react: { useSuspense: false },
    parseMissingKeyHandler: key => {
      console.warn("ui_message_missing", key);
      return instance.language === "en-US" ? enCommon.messageUnavailable : zhCommon.messageUnavailable;
    },
  });
}

export const uiI18n = createUiI18n();
