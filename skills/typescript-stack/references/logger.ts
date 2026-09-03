import { pino } from "pino";
import type { Config } from "./config.ts";

// Cloud Logging reads `severity` and `message`; pino's defaults (`level`, `msg`) arrive as plain text.
export function newLogger({ ENVIRONMENT, LOG_LEVEL }: Pick<Config, "ENVIRONMENT" | "LOG_LEVEL">) {
  return pino({
    level: LOG_LEVEL,
    messageKey: "message",
    formatters: { level: (label) => ({ severity: label.toUpperCase() }) },
    ...(ENVIRONMENT === "development" ? { transport: { target: "pino-pretty" } } : {}),
  });
}
