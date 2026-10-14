import { expect, test } from "bun:test";
import { fixedWindowLimiter } from "../src/ratelimit";

test("fixedWindowLimiter allows the limit per key per window", () => {
  let now = 1_000;
  const allow = fixedWindowLimiter(2, 60_000, () => now);
  expect([allow("a"), allow("a"), allow("a"), allow("b")]).toEqual([true, true, false, true]);
  now += 59_999;
  expect(allow("a")).toBe(false);
  now += 1;
  expect(allow("a")).toBe(true);
});
