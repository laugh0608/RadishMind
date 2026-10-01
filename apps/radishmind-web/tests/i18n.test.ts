import assert from "node:assert/strict";
import test from "node:test";
import { readdir } from "node:fs/promises";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { I18nextProvider, useTranslation } from "react-i18next";
import { createLocalePreferenceStore, resolveUiLocale, UI_LOCALE_KEY } from "../src/i18n/localePreference.ts";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { formatDisplayDate, formatDisplayNumber } from "../src/i18n/formatters.ts";

function messages(value: object, prefix = ""): Record<string, string> {
  return Object.fromEntries(Object.entries(value).flatMap(([key, entry]) => typeof entry === "string"
    ? [[prefix + key, entry]] : Object.entries(messages(entry, prefix + key + "."))));
}

test("UI locale resolves explicit preference, browser priority and unsupported languages", () => {
  assert.equal(resolveUiLocale("en-US", ["zh-CN"]), "en-US");
  assert.equal(resolveUiLocale("invalid", ["fr-FR", "en-GB", "zh-TW"]), "en-US");
  assert.equal(resolveUiLocale(null, ["zh-TW"]), "zh-CN");
  assert.equal(resolveUiLocale(null, ["french", "english"]), "zh-CN");
  assert.equal(resolveUiLocale({}, []), "zh-CN");
});

test("UI locale persists only the enum and cross-tab changes never write back", async () => {
  const writes: unknown[] = [], applied: unknown[] = [];
  const store = createLocalePreferenceStore(() => ({getItem: () => null, setItem: (...args) => writes.push(args)}), ["zh"], async locale => { applied.push(locale); });
  let notifications = 0;
  const unsubscribe = store.subscribe(() => notifications++);
  await store.select("en-US");
  assert.deepEqual(writes, [[UI_LOCALE_KEY, "en-US"]]);
  await store.acceptStorageChange("identity", "zh-CN");
  await store.acceptStorageChange(UI_LOCALE_KEY, "unexpected");
  await store.acceptStorageChange(UI_LOCALE_KEY, "en-US");
  assert.equal(applied.length, 1);
  await store.acceptStorageChange(UI_LOCALE_KEY, null);
  assert.deepEqual(store.getSnapshot(), {locale: "zh-CN", saved: true});
  assert.equal(writes.length, 1); assert.equal(notifications, 2);
  unsubscribe(); await store.select("en-US");
  assert.equal(notifications, 2);
});

test("UI locale storage failures retain memory choice and expose unsaved state", async () => {
  const store = createLocalePreferenceStore(() => { throw new Error("denied"); }, ["en"], async () => {});
  assert.deepEqual(store.getSnapshot(), {locale: "en-US", saved: false});
  await store.select("zh-CN");
  assert.deepEqual(store.getSnapshot(), {locale: "zh-CN", saved: false});
});

test("failed locale application does not persist a changed preference", async () => {
  let writes = 0;
  const store = createLocalePreferenceStore(() => ({getItem: () => "zh-CN", setItem: () => { writes++; }}), [], async () => { throw new Error("unavailable"); });
  await assert.rejects(store.select("en-US"));
  assert.equal(store.getSnapshot().locale, "zh-CN"); assert.equal(writes, 0);
});

test("late locale completion cannot replace a more recent choice", async () => {
  let release!: () => void;
  const store = createLocalePreferenceStore(() => ({getItem: () => "zh-CN", setItem: () => {}}), [], locale => locale === "en-US" ? new Promise<void>(resolve => { release = resolve; }) : Promise.resolve());
  const old = store.select("en-US"); await store.select("zh-CN"); release(); await old;
  assert.equal(store.getSnapshot().locale, "zh-CN");
});

test("all shipped resources have matching nonempty keys and interpolation contracts", async () => {
  const base = new URL("../src/i18n/locales/", import.meta.url);
  const enFiles = (await readdir(new URL("en-US/", base))).sort();
  assert.deepEqual(enFiles, (await readdir(new URL("zh-CN/", base))).sort());
  for (const file of enFiles) {
    const en = messages(await import(new URL(`en-US/${file}`, base).href));
    const zh = messages(await import(new URL(`zh-CN/${file}`, base).href));
    assert.deepEqual(Object.keys(en).sort(), Object.keys(zh).sort(), file);
    for (const key of Object.keys(en)) {
      const label = `${file}:${key}`;
      assert.ok(en[key].trim(), label); assert.ok(zh[key].trim(), label);
      const parameters = (value: string) => [...value.matchAll(/\{\{\s*(\w+)\s*\}\}/g)].map(match => match[1]).sort();
      assert.deepEqual(parameters(en[key]), parameters(zh[key]), label);
    }
  }
});

test("isolated translation instances switch messages with an explicit missing-key fallback", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "zh-CN");
  assert.equal(instance.t($ => $.language), "界面语言");
  await instance.changeLanguage("en-US");
  assert.equal(instance.t($ => $.language), "Interface language");
  assert.equal(instance.exists("missing"), false);
  assert.equal(instance.t("missing" as never), "This message is unavailable. Try again or check the diagnostic reference.");
  assert.equal(instance.hasResourceBundle("zh-CN", "shell"), true);
});

test("display formatting preserves UTC, zero, missing values and large integers", () => {
  assert.equal(formatDisplayDate("invalid", "zh-CN"), null);
  assert.equal(formatDisplayDate(null, "en-US"), null);
  assert.match(formatDisplayDate("2026-09-23T00:00:00Z", "en-US", {timeZone: "UTC"})!, /Sep 23, 2026/);
  assert.equal(formatDisplayNumber(0, "en-US"), "0");
  assert.equal(formatDisplayNumber(null, "en-US"), null);
  assert.equal(formatDisplayNumber(NaN, "en-US"), null);
  assert.equal(formatDisplayNumber(9007199254740993n, "en-US"), "9,007,199,254,740,993");
  assert.equal(formatDisplayNumber(.125, "en-US", {style: "percent", maximumFractionDigits: 1}), "12.5%");
});

test("denied writes recover from a matching cross-tab preference without writing back", async () => {
  let writes = 0;
  const store = createLocalePreferenceStore(() => ({getItem: () => "zh-CN", setItem: () => { writes++; throw new Error("quota"); }}), ["en"], async () => {});
  await store.select("en-US");
  assert.deepEqual(store.getSnapshot(), {locale: "en-US", saved: false});
  await store.acceptStorageChange(UI_LOCALE_KEY, "en-US");
  assert.deepEqual(store.getSnapshot(), {locale: "en-US", saved: true});
  await store.acceptStorageChange(null, null);
  assert.equal(store.getSnapshot().locale, "en-US");
  assert.equal(writes, 1);
});

test("localized interpolated user text stays escaped at the React render boundary", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "common", { untrustedDisplayTest: "Value: {{value}}" });
  function Display() {
    const { t } = useTranslation();
    return createElement("p", null, t("untrustedDisplayTest" as never, {value: '<img src=x onerror="alert(1)">'}));
  }
  const html = renderToStaticMarkup(createElement(I18nextProvider, {i18n: instance}, createElement(Display)));
  assert.ok(html.includes("&lt;img"));
  assert.ok(!html.includes("<img"));
  assert.ok(!html.includes("&amp;lt;"));
});

test("schema diagnostics preserve stable codes and field paths for both display languages", async () => {
  const { parsePromptOutputSchema } = await import("../src/features/control-plane-read/promptTemplateOutputSchema.ts");
  assert.deepEqual(parsePromptOutputSchema("{").issue, {code: "json", path: "schema"});
  assert.deepEqual(parsePromptOutputSchema(JSON.stringify({type: "object", properties: {score: {type: "string", minimum: 0}}})).issue,
    {code: "keyword", path: "schema.properties.score"});
  const parsed = parsePromptOutputSchema('{"type":"object","properties":{"title":{"type":"string"}}}');
  assert.equal(parsed.error, "");
  assert.equal(parsed.issue, undefined);
  assert.deepEqual(parsed.schema, {type: "object", additionalProperties: false, properties: {title: {type: "string", additionalProperties: false}}});
});
