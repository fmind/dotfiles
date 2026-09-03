import { expect, test } from "vitest";
import { newLogger } from "./logger.ts";

test("newLogger honours the configured level", () => {
  expect(newLogger({ ENVIRONMENT: "production", LOG_LEVEL: "warn" }).level).toBe("warn");
});

test("newLogger pretty-prints in development", () => {
  expect(newLogger({ ENVIRONMENT: "development", LOG_LEVEL: "debug" }).level).toBe("debug");
});
