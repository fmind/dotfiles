import { expect, test } from "vitest";
import { greet } from "./lib.ts";

test("greet renders the name", () => {
  expect(greet({ name: "Fmind" })).toBe("Hello, Fmind!");
});
