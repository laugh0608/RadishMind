import { spawn } from "node:child_process";
import { createServer } from "node:net";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { setTimeout as delay } from "node:timers/promises";

// run.mjs owns this process group, both children, logs and the temporary database.
// No evaluation runner, provider fixture or external identity issuer is required.
const webRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const repoRoot = resolve(webRoot, "../..");
if (!process.env.RADISHMIND_PLATFORM_CONFIG || !process.env.RADISHMIND_SQLITE_DEV_DATABASE_PATH) {
  throw new Error("Identity services require the isolated browser runner configuration.");
}

for (const port of [4100, 17000]) {
  const probe = createServer();
  await new Promise((accept, reject) => {
    probe.once("error", reject);
    probe.listen(port, "127.0.0.1", accept);
  });
  await new Promise((accept, reject) => probe.close(error => error ? reject(error) : accept()));
}

function start(command, args, env, cwd = webRoot) {
  const child = spawn(command, args, { cwd, env, stdio: "inherit" });
  return new Promise((_, reject) => {
    child.once("error", reject);
    child.once("exit", (code, signal) => reject(new Error(`Identity service exited (${code ?? signal}).`)));
  });
}

const backend = start("go", ["test", "-tags=identity_browser_test", "./internal/httpapi", "-run", "^TestLocalIdentityBrowserServer$", "-count=1", "-timeout=10m"], {
  ...process.env,
  GOCACHE: process.env.GOCACHE ?? "/tmp/radishmind-go-build-cache",
  RADISHMIND_IDENTITY_BROWSER_TEST: "1",
}, join(repoRoot, "services/platform"));
const frontend = start(process.execPath, [join(webRoot, "node_modules/vite/bin/vite.js"), "--host", "127.0.0.1", "--port", "4100", "--config", "tests/e2e/vite.config.ts"], {
  ...process.env,
  VITE_RADISHMIND_READ_SOURCE: "dev-live-http",
  VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL: "http://127.0.0.1:17000",
  VITE_RADISHMIND_READ_AUTH_MODE: "local_session_dev_test",
  VITE_RADISHMIND_LOCAL_IDENTITY_MODE: "local_identity_dev",
  VITE_RADISHMIND_LOCAL_IDENTITY_BASE_URL: "http://127.0.0.1:17000",
});

async function ready() {
  for (;;) {
    try {
      const health = await fetch("http://127.0.0.1:17000/healthz", { signal: AbortSignal.timeout(1000) });
      const web = await fetch("http://127.0.0.1:4100", { signal: AbortSignal.timeout(1000) });
      if (health.ok && web.ok) return;
    } catch (error) {
      if (!(error instanceof TypeError) && error.name !== "TimeoutError") throw error;
    }
    await delay(200);
  }
}

await Promise.race([backend, frontend, ready()]);
console.log("RadishMind web is ready: http://127.0.0.1:4100");
await Promise.race([backend, frontend]);
