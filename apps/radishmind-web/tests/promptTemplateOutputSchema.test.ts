import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import { formatPromptOutputSchema, parsePromptOutputSchema, validatePromptOutputSchema } from "../src/features/control-plane-read/promptTemplateOutputSchema.ts";

const scalar = (type = "string") => ({ type, additionalProperties: false });
const object = (properties: Record<string, unknown> = {}, required: string[] = []) => ({ type: "object", properties, required, additionalProperties: false });

test("Prompt output schema round-trips nested objects, arrays and every supported scalar", () => {
  const value = object({
    diagnosis: scalar(), evidence: { ...scalar("array"), items: scalar() },
    detail: object({ count: scalar("integer"), ratio: scalar("number"), known: scalar("boolean") }, ["known"]),
  }, ["diagnosis", "evidence"]);
  const parsed = parsePromptOutputSchema(JSON.stringify(value));
  assert.equal(parsed.error, "");
  assert.deepEqual(JSON.parse(formatPromptOutputSchema(parsed.schema)), value);
  assert.equal(parsePromptOutputSchema(formatPromptOutputSchema()).error, "");
  assert.equal(validatePromptOutputSchema({ ...object(), additionalProperties: true }), "");
});

test("Prompt output schema rejects malformed JSON and unsupported or inconsistent structures", () => {
  assert.match(parsePromptOutputSchema('{"private_unfinished":').error, /完整有效的 JSON/);
  const cases: [unknown, RegExp][] = [
    [[], /根类型/], [scalar(), /根类型/],
    [{ ...object(), $ref: "external" }, /关键字/],
    [{ ...object(), additionalProperties: null }, /布尔值/],
    [{ ...object(), properties: [] }, /properties/],
    [object({}, ["missing"]), /必须在 properties/],
    [object({ value: scalar() }, ["value", "value"]), /无重复/],
    [object({ "bad-name": scalar() }), /字段名/],
    [object({}, ["bad-name"]), /合法字段名/],
    [{ ...object(), items: scalar() }, /不能声明 items/],
    [object({ list: scalar("array") }), /schema.properties.list.items/],
    [object({ value: { ...scalar(), items: scalar() } }), /标量类型/],
    [object({ value: { ...scalar(), additionalProperties: true } }), /非 object/],
    [object({ value: { ...scalar(), properties: { nested: scalar() } } }), /非 object/],
    [object({ value: scalar("null") }), /类型仅支持/],
  ];
  for (const [value, error] of cases) {
    const parsed = parsePromptOutputSchema(JSON.stringify(value));
    assert.equal(parsed.schema, undefined);
    assert.match(parsed.error, error);
  }
});

test("The existing diagnostics trial schema imports without manually filling scalar defaults", () => {
  const trial = JSON.parse(readFileSync(new URL("../../../docs/features/user-workspace/prompt-application-dev-test-usage-guide.parts/diagnostics-trial-source.json", import.meta.url), "utf8"));
  const parsed = parsePromptOutputSchema(JSON.stringify(trial.output_contract.json_schema));
  assert.equal(parsed.error, "");
  assert.equal(parsed.schema?.properties?.diagnosis?.additionalProperties, false);
  assert.equal(parsed.schema?.properties?.evidence?.items?.additionalProperties, false);
  assert.deepEqual(parsed.schema?.required, trial.output_contract.json_schema.required);
  assert.match(validatePromptOutputSchema(trial.output_contract.json_schema), /布尔值/);
});

test("Prompt output schema enforces root-inclusive depth and cumulative property budgets", () => {
  let nested: unknown = scalar();
  for (let depth = 1; depth < 8; depth++) nested = object({ nested });
  assert.equal(validatePromptOutputSchema(nested), "");
  assert.match(validatePromptOutputSchema(object({ nested })), /8 层/);
  const fields = Object.fromEntries(Array.from({ length: 127 }, (_, index) => [`value${index}`, scalar()]));
  assert.equal(validatePromptOutputSchema(object({ nested: object(fields) })), "");
  assert.match(validatePromptOutputSchema(object({ nested: object({ ...fields, extra: scalar() }) })), /128 个字段/);
  let items: unknown = scalar();
  for (let level = 0; level < 6; level++) items = { ...scalar("array"), items };
  const wide = Object.fromEntries(Array.from({ length: 128 }, (_, index) => [`field${index}`, items]));
  assert.match(validatePromptOutputSchema(object(wide)), /32 KiB/);
});
