import { describe, expect, it, vi } from "vitest";
import { chmod, mkdtemp, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { removeSavedRemote } from "./remotes-ipc";


const TWO_HOSTS =
	'{"remotes":[{"label":"workbox","url":"http://192.0.2.1:1","password":"old"},{"label":"mini","url":"http://192.0.2.9:9","password":"m"}]}';

async function tempFile(contents = TWO_HOSTS, mode = 0o600): Promise<string> {
	const dir = await mkdtemp(join(tmpdir(), "ao-remotes-ipc-"));
	const path = join(dir, "remotes.json");
	await writeFile(path, contents, "utf8");
	await chmod(path, mode);
	return path;
}


const dropped = () => vi.fn<(url: string) => Promise<void>>().mockResolvedValue(undefined);

describe("POSIX saved-host permissions", () => {
	it("refuses to touch a file others can read", async () => {
		const path = await tempFile(TWO_HOSTS, 0o644);
		const disconnect = dropped();
		await expect(removeSavedRemote(path, "http://192.0.2.1:1", disconnect)).rejects.toThrow(/chmod 600/);
		expect(disconnect).not.toHaveBeenCalled();
	});
});
