import { expect, test } from "vitest";
import { loadConfig } from "./config.ts";

test("loadConfig applies defaults", () => {
  expect(loadConfig({})).toMatchObject({ ENVIRONMENT: "development", PORT: 8080 });
});

test("loadConfig rejects an unknown environment", () => {
  expect(() => loadConfig({ ENVIRONMENT: "staging" })).toThrow(/invalid environment/);
});
