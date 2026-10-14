import { delay, http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { calculateUrl, operationsUrl, server } from "../test/server";
import { calculate, fetchOperations } from "./client";

describe("API client", () => {
  it("stops a request after the timeout and returns NETWORK_ERROR", async () => {
    server.use(
      http.post(calculateUrl, async () => {
        await delay("infinite");
        return HttpResponse.json({});
      }),
    );
    await expect(calculate("2+3", { timeoutMs: 20 })).resolves.toEqual({
      ok: false,
      code: "NETWORK_ERROR",
    });
  });

  it("returns INTERNAL_ERROR for a success body with the wrong shape", async () => {
    server.use(http.post(calculateUrl, () => HttpResponse.json({ result: "five" })));
    await expect(calculate("2+3")).resolves.toEqual({ ok: false, code: "INTERNAL_ERROR" });
  });

  it("returns INTERNAL_ERROR for an unknown error code", async () => {
    server.use(
      http.post(calculateUrl, () =>
        HttpResponse.json({ error: { code: "NEW_CODE", message: "x" } }, { status: 400 }),
      ),
    );
    await expect(calculate("2+3")).resolves.toEqual({ ok: false, code: "INTERNAL_ERROR" });
  });

  it("returns NOT_FOUND for a 404 error", async () => {
    server.use(
      http.post(calculateUrl, () =>
        HttpResponse.json({ error: { code: "NOT_FOUND", message: "x" } }, { status: 404 }),
      ),
    );
    await expect(calculate("2+3")).resolves.toEqual({ ok: false, code: "NOT_FOUND" });
  });

  it("returns INTERNAL_ERROR for an operations list with the wrong shape", async () => {
    server.use(http.get(operationsUrl, () => HttpResponse.json({ operations: [{ name: 1 }] })));
    await expect(fetchOperations()).resolves.toEqual({ ok: false, code: "INTERNAL_ERROR" });
  });

  it("uses the base URL when it is given", async () => {
    let url = "";
    server.use(
      http.post("http://api.example/api/v1/calculate", ({ request }) => {
        url = request.url;
        return HttpResponse.json({ expression: "2 + 3", result: 5 });
      }),
    );
    await expect(calculate("2+3", { baseUrl: "http://api.example" })).resolves.toEqual({
      ok: true,
      value: { expression: "2 + 3", result: 5 },
    });
    expect(url).toBe("http://api.example/api/v1/calculate");
  });
});
