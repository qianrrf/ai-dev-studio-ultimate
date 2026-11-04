import { mkdtemp } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { TelemetryPolicyAuthority, nodeTelemetryPolicyFileSystem, resolveDesktopDataDir } from "./telemetry-policy-file";

const dirs: string[] = [];
const makeDir = async () => {
	const dir = await mkdtemp(path.join(os.tmpdir(), "ao-policy-"));
	dirs.push(dir);
	return dir;
};

afterEach(async () => {
	const { rm } = await import("node:fs/promises");
	await Promise.all(dirs.splice(0).map((dir) => rm(dir, { recursive: true, force: true })));
});

describe("TelemetryPolicyAuthority", () => {
	it("supports Windows defaults and durable opt-out replacements", async () => {
		const dataDir = await makeDir();
		const authority = new TelemetryPolicyAuthority({ dataDir, packagedDefault: true, platform: "win32" });
		expect(authority.durabilitySupported).toBe(true);
		expect(await authority.load()).toMatchObject({ eventsEnabled: true, acknowledged: true });
		expect(await authority.setEventsEnabled(false)).toMatchObject({ eventsEnabled: false, acknowledged: true });
		expect(await new TelemetryPolicyAuthority({ dataDir, packagedDefault: true, platform: "win32" }).load()).toMatchObject({ eventsEnabled: false, acknowledged: true });
	});

	it.each(["EPERM", "EBUSY"])("retries a Windows %s replacement without changing the policy generation", async (code) => {
		const dataDir = await makeDir();
		const base = nodeTelemetryPolicyFileSystem;
		let attempts = 0;
		const authority = new TelemetryPolicyAuthority({
			dataDir, packagedDefault: true, platform: "win32",
			fs: { ...base, rename: async (from, to) => {
				attempts++;
				if (attempts === 2) throw Object.assign(new Error("locked"), { code });
				await base.rename(from, to);
			} },
		});
		await authority.load();
		const disabled = await authority.setEventsEnabled(false);
		expect(disabled).toMatchObject({ eventsEnabled: false, acknowledged: true });
		expect(attempts).toBe(3);
		expect((await new TelemetryPolicyAuthority({ dataDir, packagedDefault: true, platform: "win32" }).load()).consentGeneration).toBe(disabled.consentGeneration);
	});

});

it("resolves one absolute data directory against the daemon launch cwd", () => {
	expect(resolveDesktopDataDir({ AO_DATA_DIR: "relative-data" }, "/home/ao", "/work/checkout", false)).toBe(path.resolve("/work/checkout", "relative-data"));
	expect(resolveDesktopDataDir({}, "/home/ao", "/work/checkout", true)).toBe(path.resolve("/home/ao", ".ai-dev-studio-ultimate", "data"));
	expect(resolveDesktopDataDir({}, "/home/ao", "/work/checkout", false)).toBe(path.resolve("/home/ao", ".ai-dev-studio-ultimate", "dev", "data"));
});
