import { expect, it, vi } from "vitest";

const getBootstrap = vi.hoisted(() => vi.fn());
vi.mock("./bridge", () => ({ aoBridge: { telemetry: { getBootstrap } } }));
import { initTelemetry } from "./telemetry";

it("the product cannot initialize analytics even when a bootstrap is available", async () => {
	getBootstrap.mockResolvedValue({ distinctId: "synthetic", appVersion: "0.1.0-alpha.1" });
	expect(await initTelemetry()).toBe(false);
	expect(getBootstrap).not.toHaveBeenCalled();
});
