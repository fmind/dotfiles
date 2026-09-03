#!/usr/bin/env node
import { loadConfig } from "./config.ts";
import { greet } from "./lib.ts";
import { newLogger } from "./logger.ts";

// Service entry point: `node src/main.ts` in development, `node dist/main.js` in production.
function main(): void {
  const config = loadConfig();
  const logger = newLogger(config);
  logger.info({ port: config.PORT }, greet({ name: "world" }));
}

main();
