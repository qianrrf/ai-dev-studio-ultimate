import type { DesktopTelemetryTransport } from "./desktop-telemetry-controller";
export async function initMainSentry(..._args: unknown[]): Promise<DesktopTelemetryTransport | null> { return null; }
export async function clearMainSentryCache(_root: string): Promise<void> {}
export function sanitizeRendererCapture(_request: unknown): null { return null; }
