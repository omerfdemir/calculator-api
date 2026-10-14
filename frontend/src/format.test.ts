import { describe, expect, it } from "vitest";
import { formatNumber } from "./format";

describe("formatNumber", () => {
  it.each([
    [0.30000000000000004, "0.3"],
    [1 / 3, "0.333333333333"],
    [2 / 3, "0.666666666667"],
    [1000001, "1000001"],
    [123456789012345, "123456789012000"],
    [1e21, "1e21"],
    [1.5e-7, "1.5e−7"],
    [-3, "−3"],
    [0, "0"],
  ])("formats %d as %s", (value, text) => {
    expect(formatNumber(value)).toBe(text);
  });
});
