// Public surface of the package: one barrel so `exports` in package.json stays a single entry.
export { type Config, loadConfig } from "./config.ts";
export { type Greeting, greet } from "./lib.ts";
export { newLogger } from "./logger.ts";
