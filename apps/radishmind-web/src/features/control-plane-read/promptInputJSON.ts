export type PromptInputJSONIssue = {
  code: "jsonSyntax" | "duplicateProperty" | "numberPrecision" | "jsonDepth" | "unicodeValue";
  line: number;
  column: number;
};

export type PromptInputJSONResult =
  | { ok: true; value: unknown }
  | { ok: false; issue: PromptInputJSONIssue };

// JSON.parse validates syntax. This lexical pass rejects information it would
// otherwise discard, before the parsed value can enter the editor or a request.
export function parsePromptInputJSON(text: string): PromptInputJSONResult {
  let value: unknown;
  try {
    value = JSON.parse(text);
  } catch (error) {
    const position = error instanceof SyntaxError ? /position (\d+)/u.exec(error.message)?.[1] : undefined;
    return failure("jsonSyntax", position ? Number(position) : 0);
  }
  const containers: Array<Set<string> | null> = [];
  for (let cursor = 0; cursor < text.length;) {
    const char = text[cursor];
    if (char === "{" || char === "[") {
      if (containers.length === 64) return failure("jsonDepth", cursor);
      containers.push(char === "{" ? new Set() : null);
      cursor += 1;
    } else if (char === "}" || char === "]") {
      containers.pop();
      cursor += 1;
    } else if (char === '"') {
      const start = cursor++;
      while (text[cursor] !== '"') cursor += text[cursor] === "\\" ? 2 : 1;
      cursor += 1;
      const decoded = JSON.parse(text.slice(start, cursor)) as string;
      if (/[\uD800-\uDFFF]/u.test(decoded)) return failure("unicodeValue", start);
      let next = cursor;
      while (/\s/u.test(text[next] ?? "x")) next += 1;
      const keys = containers.at(-1);
      if (text[next] === ":" && keys) {
        const key = decoded;
        if (keys.has(key)) return failure("duplicateProperty", start);
        keys.add(key);
      }
    } else if (char === "-" || /\d/u.test(char)) {
      const number = /^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/u.exec(text.slice(cursor))![0];
      if (!promptNumberIsLossless(number)) return failure("numberPrecision", cursor);
      cursor += number.length;
    } else {
      cursor += 1;
    }
  }
  return { ok: true, value };

  function failure(code: PromptInputJSONIssue["code"], position: number): PromptInputJSONResult {
    const before = text.slice(0, position).split("\n");
    return { ok: false, issue: { code, line: before.length, column: before.at(-1)!.length + 1 } };
  }
}

export function promptNumberIsLossless(text: string): boolean {
  if (!/^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?$/u.test(text)) return false;
  const number = Number(text);
  if (!Number.isFinite(number) || Object.is(number, -0) || Number.isInteger(number) && !Number.isSafeInteger(number)) return false;
  return decimalIdentity(text) === decimalIdentity(JSON.stringify(number));
}

function decimalIdentity(text: string): string {
  const [mantissa, exponent = "0"] = text.toLowerCase().split("e");
  const [whole, fraction = ""] = mantissa.split(".");
  let digits = `${whole.replace("-", "")}${fraction}`.replace(/^0+/u, "");
  if (!digits) return "0";
  const trailing = /0+$/u.exec(digits)?.[0].length ?? 0;
  digits = digits.slice(0, digits.length - trailing);
  return `${whole.startsWith("-") ? "-" : ""}${digits}e${Number(exponent) - fraction.length + trailing}`;
}
