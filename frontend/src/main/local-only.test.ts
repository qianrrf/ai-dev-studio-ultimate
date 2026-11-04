import { afterEach, describe, expect, it, vi } from "vitest";
import { mkdtemp, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";

const handlers = vi.hoisted(() => new Map<string, (...args: unknown[]) => unknown>());
vi.mock("electron", () => ({
	app: { isPackaged: true, setAsDefaultProtocolClient: vi.fn() },
	ipcMain: { handle: (channel: string, handler: (...args: unknown[]) => unknown) => handlers.set(channel, handler) },
}));
import { getCloudSession, installCloudIPC, installCloudLocalAuthIPC, installCloudCpProxy } from "./local-only-cloud";
import { getUpdateStatus, setUpdateSettings, quitAndInstallUpdate } from "./local-only-updates";
import { readUpdateSettings } from "./update-settings";
import { initMainSentry, sanitizeRendererCapture } from "./local-only-telemetry";

afterEach(() => { handlers.clear(); vi.unstubAllGlobals(); });

describe("local edition boundaries", () => {
	it("rejects hosted auth and proxy requests without making network calls", async () => {
		const fetch = vi.fn();
		vi.stubGlobal("fetch", fetch);
		installCloudIPC(() => "unused", () => {});
		installCloudLocalAuthIPC(() => "unused", () => {});
		installCloudCpProxy(() => "unused", null);
		expect(await getCloudSession("unused")).toBeNull();
		expect(handlers.get("cloud:getSession")!()).toBeNull();
		expect(handlers.get("cloud:localAuthAvailable")!()).toBe(false);
		for (const channel of ["cloud:signIn", "cloud:connectProviderAuth", "cloud:localRegister", "cloud:localLogin", "cloudCp:request", "cloudCp:requestBinary", "cloudCp:openStream"]) {
			expect(() => handlers.get(channel)!({}, { baseUrl: "https://example.test" })).toThrow(/unavailable/);
		}
		expect(fetch).not.toHaveBeenCalled();
	});

	it("cannot enable updates or install an upstream update", async () => {
		const dir = await mkdtemp(path.join(os.tmpdir(), "ultimate-updates-"));
		try {
			await setUpdateSettings(dir, { enabled: true, channel: "nightly", nightlyAck: true, feature: { pr: 1 } });
			expect(await readUpdateSettings(dir)).toMatchObject({ enabled: false, channel: "latest", feature: null });
			expect(getUpdateStatus().state).toBe("unsupported");
			await expect(quitAndInstallUpdate()).rejects.toThrow(/unavailable/);
		} finally { await rm(dir, { recursive: true, force: true }); }
	});

	it("never creates an error-reporting transport", async () => {
		expect(await initMainSentry()).toBeNull();
		expect(sanitizeRendererCapture({ kind: "exception" })).toBeNull();
	});
});
