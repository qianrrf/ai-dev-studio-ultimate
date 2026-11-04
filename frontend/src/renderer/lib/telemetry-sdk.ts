import type posthog from "posthog-js/dist/module.full.no-external";

// Retain the telemetry algorithms without shipping an analytics SDK in the alpha.
export async function loadPostHog(): Promise<typeof posthog> {
	throw new Error("Telemetry is unavailable in this local edition.");
}
