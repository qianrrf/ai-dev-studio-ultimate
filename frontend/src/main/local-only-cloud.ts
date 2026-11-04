// Keep the existing preload contract available without loading hosted auth SDKs.
import { app, ipcMain } from "electron";
import path from "node:path";
import type { CloudAccount } from "../shared/cloud-account";
import { PRODUCT_ID, PRODUCT_STATE_DIRECTORY } from "../shared/product";

const unavailable = () => { throw new Error("Cloud features are unavailable in this local edition."); };
export async function getCloudSession(_dataDir: string): Promise<CloudAccount | null> { return null; }
export async function getCloudAccessToken(_dataDir: string): Promise<string | null> { return null; }
export async function handleCloudDeepLink(_url: string, _dataDir: string): Promise<CloudAccount | null> { return null; }
export async function showCloudSignInFailure(_error: unknown): Promise<void> {}
export function resolveCloudAuthDir(env: NodeJS.ProcessEnv, home: string, cwd: string, packaged: boolean): string {
	return env.AO_DATA_DIR ? path.resolve(cwd, env.AO_DATA_DIR) : path.join(home, PRODUCT_STATE_DIRECTORY, ...(packaged ? [] : ["dev"]));
}
export function registerCloudProtocol(): void {
	if (app.isPackaged) app.setAsDefaultProtocolClient(PRODUCT_ID);
}
export function installCloudIPC(_getDir: () => string, _notify: (account: CloudAccount | null) => void): void {
	ipcMain.handle("cloud:getSession", () => null);
	ipcMain.handle("cloud:signOut", () => undefined);
	ipcMain.handle("cloud:cancelProviderAuth", () => undefined);
	for (const channel of ["cloud:signIn", "cloud:connectProviderAuth"]) ipcMain.handle(channel, unavailable);
}
export function installCloudLocalAuthIPC(_getDir: () => string, _notify: (account: CloudAccount | null) => void): void {
	ipcMain.handle("cloud:localAuthAvailable", () => false);
	for (const channel of ["cloud:localRegister", "cloud:localLogin"]) ipcMain.handle(channel, unavailable);
}
export function createCloudAgentProxy(_getDir: () => string, _resolve: unknown) {
	return { agentBrowserUrl: async (..._args: unknown[]): Promise<string> => unavailable(), isAgentBrowserUrl: (_url: string) => false, close: async () => {} };
}
export function installCloudCpProxy(_getDir: () => string, _resolve: unknown): void {
	for (const channel of ["cloudCp:request", "cloudCp:requestBinary", "cloudCp:cancelBinary", "cloudCp:openStream", "cloudCp:closeStream"]) ipcMain.handle(channel, unavailable);
}
