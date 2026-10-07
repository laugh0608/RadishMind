import { spawn } from "node:child_process";
import { createWriteStream } from "node:fs";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { setTimeout as delay } from "node:timers/promises";
import { startPromptProvider } from "./prompt-provider.mjs";

if (process.platform === "win32") {
  throw new Error("Browser regression service cleanup requires POSIX process groups; use Linux, macOS, or WSL.");
}

const webRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const repoRoot = resolve(webRoot, "../..");
const artifactRoot = join(repoRoot, "output/playwright/workflow-e2e");
await mkdir(artifactRoot, { recursive: true });
// Each profile owns a fresh database and releases the shared loopback ports before the next.
// Template metadata uses a deliberately non-callable provider, unlike the Prompt fixture.
const suiteFlags = process.argv.slice(2).filter(value => value.startsWith("--suite="));
const availableSuites = ["templates", "workflow", "rag", "rag-promotion", "rag-application", "http-tool", "offline-projection", "identity"];
if (suiteFlags.length > 1 || (suiteFlags[0] && !availableSuites.includes(suiteFlags[0].slice(8)))) {
  throw new Error("Use at most one --suite=templates|workflow|rag|rag-promotion|rag-application|http-tool|offline-projection|identity selector.");
}
const suites = suiteFlags.length ? [suiteFlags[0].slice(8)] : availableSuites;
const testArguments = process.argv.slice(2).filter(value => !value.startsWith("--suite="));
for (const suite of suites) {
  const code = await runSuite(suite);
  if (code !== 0) { process.exitCode = code; break; }
}

async function runSuite(suite) {
  const output = await mkdtemp(join(artifactRoot, "run-"));
  const runtime = await mkdtemp(join(tmpdir(), "radishmind-workflow-e2e-"));
  const configPath = join(runtime, "platform.json");

  // Pass toolchain settings, not inherited application settings, tokens, or database URLs.
  const environmentKeys = new Set([
    "PATH", "HOME", "USER", "SHELL", "TMPDIR", "LANG", "LC_ALL", "TERM", "CI",
    "GOCACHE", "GOMODCACHE", "GOPATH", "GOROOT", "GOENV", "GOTOOLCHAIN",
    "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY",
    "http_proxy", "https_proxy", "all_proxy", "no_proxy",
    "NODE_EXTRA_CA_CERTS", "SSL_CERT_FILE", "PLAYWRIGHT_BROWSERS_PATH",
  ]);
  const environment = Object.fromEntries(Object.entries(process.env).filter(([key]) => environmentKeys.has(key)));
  environment.NO_PROXY = environment.no_proxy = "127.0.0.1,localhost";
  const groups = [];
  const logs = [];
  let interrupted = false;
  let shutdown;
  let promptProvider;

  function groupExists(pid) {
    try {
      process.kill(-pid, 0);
      return true;
    } catch (error) {
      if (error.code === "ESRCH") return false;
      throw error;
    }
  }

  function signalGroup(pid, signal) {
    try {
      process.kill(-pid, signal);
    } catch (error) {
      if (error.code !== "ESRCH") throw error;
    }
  }

  function stop() {
    shutdown ??= (async () => {
      for (const child of groups) signalGroup(child.pid, "SIGTERM");
      const deadline = Date.now() + 8_000;
      while (groups.some((child) => groupExists(child.pid)) && Date.now() < deadline) await delay(100);
      for (const child of groups) {
        if (groupExists(child.pid)) signalGroup(child.pid, "SIGKILL");
        await child.finished;
      }
      try {
        if (promptProvider) {
          await promptProvider.close();
          await writeFile(join(output, "prompt-provider-observations.json"), JSON.stringify(promptProvider.observations, null, 2));
        }
        for (const log of logs) {
          if (!log.destroyed) await new Promise((accept, reject) => {
            log.once("error", reject);
            log.end(accept);
          });
        }
      } finally {
        await rm(runtime, { recursive: true, force: true });
      }
      console.log(`[workflow-e2e] Stopped owned process groups; removed temporary runtime ${runtime}`);
    })();
    return shutdown;
  }

  function start(command, args, options = {}) {
    const child = spawn(command, args, { cwd: webRoot, env: environment, detached: true, ...options });
    child.finished = new Promise((accept, reject) => {
      child.once("error", reject);
      child.once("close", (code, signal) => accept({ code, signal }));
    });
    // A grandchild may keep stdout open after its launcher exits. Observe exit
    // for liveness, and close only when waiting for cleanup and flushed output.
    child.exited = new Promise((accept, reject) => {
      child.once("error", reject);
      child.once("exit", (code, signal) => accept({ code, signal }));
    });
    child.finished.catch(() => {});
    child.exited.catch(() => {});
    if (child.pid) groups.push(child);
    return child;
  }

  const signalHandlers = new Map();
  for (const signal of ["SIGINT", "SIGTERM"]) {
    const handler = () => {
      interrupted = true;
      void stop().then(() => process.exit(signal === "SIGINT" ? 130 : 143), (error) => {
        console.error(error);
        process.exit(1);
      });
    };
    signalHandlers.set(signal, handler);
    process.once(signal, handler);
  }

  console.log(`[workflow-e2e] ${suite} artifacts: ${output}`);
  let exitCode = 1;
  try {
    await writeFile(configPath, "{}\n", { mode: 0o600 });
    if (suite !== "identity") promptProvider = await startPromptProvider();
    if (interrupted) throw new Error("Interrupted during configuration setup.");
    const launcher = start(suite === "identity" ? process.execPath : "bash", suite === "identity" ? [join(webRoot, "tests/e2e/identity-services.mjs")] : [
      join(repoRoot, "scripts/run-radishmind-web-dev.sh"),
      "--mode", "dev-live", ...(suite === "http-tool" ? ["--workflow-definition-http-tool-local-product"] : suite === "rag-application" ? ["--workflow-rag-application-local-product"] : suite === "rag-promotion" ? ["--workflow-rag-promotion-local-product"] : suite === "rag" ? ["--workflow-rag-dev"] : ["--workflow-definition-local-product", suite === "templates" ? "--workflow-template-local-product" : "--prompt-application-local-product"]), "--no-reuse-existing",
      "--frontend-url", "http://127.0.0.1:4100", "--backend-url", "http://127.0.0.1:17000",
      "--timeout-seconds", "120", "--log-dir", join(output, "services"),
      "--frontend-config", join(webRoot, "tests/e2e/vite.config.ts"),
    ], {
      cwd: repoRoot,
      env: {
        ...environment,
        RADISHMIND_PLATFORM_CONFIG: configPath,
        RADISHMIND_PLATFORM_PROVIDER: "mock",
        ...(promptProvider ? {
          RADISHMIND_MODEL_PROFILE: "prompt-e2e",
          RADISHMIND_MODEL_PROFILE_FALLBACKS: "prompt-e2e",
          RADISHMIND_MODEL_PROFILE_PROMPT_E2E_NAME: "prompt-e2e-model",
          RADISHMIND_MODEL_PROFILE_PROMPT_E2E_BASE_URL: `${promptProvider.url}/v1`,
          RADISHMIND_MODEL_PROFILE_PROMPT_E2E_API_KEY: "prompt-e2e-fixture-only",
          RADISHMIND_MODEL_PROFILE_PROMPT_E2E_API_STYLE: "openai-compatible",
          RADISHMIND_MODEL_PROFILE_PROMPT_E2E_REQUEST_TIMEOUT_SECONDS: "5",
        } : {}),
        RADISHMIND_SQLITE_DEV_DATABASE_PATH: join(runtime, "workflow.db"),
      },
      stdio: ["ignore", "pipe", "pipe"],
    });
    const launcherLog = createWriteStream(join(output, "launcher.log"));
    logs.push(launcherLog);
    const loggingFailure = new Promise((_, reject) => launcherLog.once("error", reject));
    loggingFailure.catch(() => {});
    launcher.stdout.pipe(launcherLog, { end: false });
    launcher.stderr.pipe(launcherLog, { end: false });
    let ready;
    const readiness = new Promise((accept) => { ready = accept; });
    let partialLine = "";
    launcher.stdout.on("data", (chunk) => {
      partialLine += chunk.toString();
      if (partialLine.includes("RadishMind web is ready: http://127.0.0.1:4100")) ready();
      partialLine = partialLine.slice(-2048);
    });
    const startupTimer = new AbortController();
    try {
      await Promise.race([
        readiness,
        loggingFailure,
        launcher.exited.then(({ code, signal }) => { throw new Error(`Service launcher exited before readiness (${code ?? signal}). See ${output}/services.`); }),
        delay(150_000, null, { signal: startupTimer.signal }).then(() => { throw new Error("Isolated service startup timed out."); }),
      ]);
    } finally {
      startupTimer.abort();
    }
    if (interrupted) throw new Error("Interrupted during startup.");
    console.log(`[workflow-e2e] SQLite services ready on 4100 and 17000${promptProvider ? `; Prompt fixture ${promptProvider.url}` : "; identity only, no model provider"}.`);
    const tests = start(process.execPath, [
      join(webRoot, "node_modules/@playwright/test/cli.js"), "test",
      "--config", "tests/e2e/playwright.config.ts", ...testArguments,
    ], {
      env: { ...environment, RADISHMIND_E2E_SUITE: suite, RADISHMIND_E2E_WEB_URL: "http://127.0.0.1:4100", RADISHMIND_E2E_OUTPUT_DIR: output, ...(suite === "identity" ? { RADISHMIND_IDENTITY_E2E_RUNTIME: runtime, PLAYWRIGHT_NO_COPY_PROMPT: "1" } : {}), ...(promptProvider ? { RADISHMIND_E2E_PROVIDER_URL: promptProvider.url } : {}) },
      stdio: "inherit",
    });
    const result = await Promise.race([
      tests.finished,
      loggingFailure,
      launcher.exited.then(() => { throw new Error("Services exited while browser tests were running."); }),
    ]);
    exitCode = result.code ?? 1;
  } catch (error) {
    console.error(`[workflow-e2e] ${error.message}`);
  } finally {
    try { await stop(); } finally {
      for (const [signal, handler] of signalHandlers) process.off(signal, handler);
    }
  }
  return exitCode;
}
