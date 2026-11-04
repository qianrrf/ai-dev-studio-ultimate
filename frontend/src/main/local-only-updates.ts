// No release feed exists for this alpha. Never load or fall back to AO's updater.
import type { UpdateSettings, UpdateStatus } from "./update-settings";
import { writeUpdateSettings } from "./update-settings";
export type { UpdateCheckOptions } from "./auto-updater";

const disabled: UpdateSettings = { enabled: false, channel: "latest", nightlyAck: false, feature: null, macDifferentialUpdates: false };
const status: UpdateStatus = { state: "unsupported", message: "Automatic updates are unavailable for this alpha. Install a verified package manually." };
export function getUpdateStatus(): UpdateStatus { return { ...status }; }
export function isUpdateRestartRequested(): boolean { return false; }
export function setRendererSink(_sink: unknown): void {}
export function setUpdateRestartFailureHandler(_handler: () => void): void {}
export function recordUpdateDiagnostic(..._args: unknown[]): void {}
export async function startAutoUpdates(_dir: string): Promise<void> {}
export async function ensureUpdatePrefs(dir: string): Promise<void> { await writeUpdateSettings(dir, disabled); }
export async function setUpdateSettings(dir: string, _settings: UpdateSettings): Promise<void> { await ensureUpdatePrefs(dir); }
export async function setMacDifferentialUpdates(dir: string, _enabled: boolean): Promise<void> { await ensureUpdatePrefs(dir); }
export async function checkForUpdatesNow(_dir: string, _options?: unknown): Promise<void> {}
export async function returnToHome(_dir: string, _requestId?: string): Promise<void> {}
export async function downloadUpdateNow(_requestId?: string): Promise<void> {}
export async function quitAndInstallUpdate(_version?: string): Promise<void> { throw new Error(status.message); }
export async function listFeatureBuilds(): Promise<never[]> { return []; }
export function getActiveFeatureBuild(): null { return null; }
