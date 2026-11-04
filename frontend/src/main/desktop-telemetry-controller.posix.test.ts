import { mkdtemp, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { describe, expect, it, vi } from "vitest";

import { DesktopTelemetryController } from "./desktop-telemetry-controller";
import { nodeTelemetryPolicyFileSystem, TelemetryPolicyAuthority } from "./telemetry-policy-file";

describe("DesktopTelemetryController POSIX durability", () => {
	it("retries the durable off generation and preserved desktop purge after a post-rename fsync failure", async () => {
		const dataDir = await mkdtemp(path.join(os.tmpdir(), "ao-controller-policy-"));
		try {
			const enabledGeneration = "7f80c8a9-ec67-4a16-a067-a444ffcc5cca";
			await writeFile(path.join(dataDir, "telemetry_policy.json"), `${JSON.stringify({
				schema_version: 2,
				events_enabled: true,
				consent_generation: enabledGeneration,
				consent_production_enabled: true,
				updated_at: "2026-08-28T10:15:30.000Z",
			})}\n`, { mode: 0o600 });
			let failDirectorySync = false;
			const base = nodeTelemetryPolicyFileSystem;
			const authority = new TelemetryPolicyAuthority({
				dataDir,
				packagedDefault: false,
				platform: "linux",
				fs: {
					...base,
					syncDirectory: async (target) => {
						if (failDirectorySync) {
							failDirectorySync = false;
							throw new Error("directory fsync failed");
						}
						await base.syncDirectory(target);
					},
				},
			});
			const transport = { closeAndDrain: vi.fn(), capture: vi.fn(), clearCache: vi.fn() };
			const clearRendererQueues = vi.fn();
			const controller = new DesktopTelemetryController({
				authority,
				transportFactory: async () => transport,
				environmentAllowsEvents: true,
				productionEnabled: true,
				clearRendererQueues,
			});
			await controller.initialize();
			failDirectorySync = true;

			await expect(controller.setEventsEnabled(false, enabledGeneration)).rejects.toThrow("directory fsync failed");
			const pending = authority.snapshot();
			expect(pending).toMatchObject({ eventsEnabled: false, acknowledged: false });
			expect(transport.clearCache).not.toHaveBeenCalled();
			expect(clearRendererQueues).not.toHaveBeenCalled();

			const retried = await controller.retryPendingCleanup();

			expect(retried).toMatchObject({
				eventsEnabled: false,
				consentGeneration: pending.consentGeneration,
				acknowledged: true,
				state: "applied",
			});
			expect(authority.snapshot()).toMatchObject({ consentGeneration: pending.consentGeneration, acknowledged: true });
			expect(transport.clearCache).toHaveBeenCalledOnce();
			expect(clearRendererQueues).toHaveBeenCalledOnce();
		} finally {
			await rm(dataDir, { recursive: true, force: true });
		}
	});
	it("starts a prior opt-in without renewal and preserves a later opt-out", async () => {
		const dataDir = await mkdtemp(path.join(os.tmpdir(), "ao-controller-prior-choice-"));
		try {
			const generation = "7f80c8a9-ec67-4a16-a067-a444ffcc5cca";
			await writeFile(path.join(dataDir, "telemetry_policy.json"), `${JSON.stringify({
				schema_version: 2,
				events_enabled: true,
				consent_generation: generation,
				consent_production_enabled: false,
				updated_at: "2026-08-28T10:15:30.000Z",
			})}\n`, { mode: 0o600 });
			const transportFactory = vi.fn(async () => ({ closeAndDrain: async () => {}, capture: vi.fn(), clearCache: async () => {} }));
			const boot = () => new DesktopTelemetryController({
				authority: new TelemetryPolicyAuthority({ dataDir, packagedDefault: true, platform: "linux" }),
				transportFactory,
				environmentAllowsEvents: true,
			});
			const first = boot();
			await first.initialize();
			expect(first.snapshot()).toMatchObject({ eventsEnabled: true, state: "applied", acknowledged: true, consentGeneration: generation });
			expect(transportFactory).toHaveBeenCalledOnce();
			expect(first.capture({ consentGeneration: generation, kind: "message", message: "synthetic" })).toBe(true);
			await first.setEventsEnabled(false, generation);
			const second = boot();
			await second.initialize();
			expect(second.snapshot().eventsEnabled).toBe(false);
			expect(transportFactory).toHaveBeenCalledOnce();
		} finally {
			await rm(dataDir, { recursive: true, force: true });
		}
	});
});
