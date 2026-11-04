import { describe, expect, it } from "vitest";
import { chmod, mkdtemp, readFile, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { readRemotes, removeRemote, RemotesFilePermissionError, updateRemote } from "./remotes-store";

async function tempFile(contents?: string, mode = 0o600): Promise<string> {
	const dir = await mkdtemp(join(tmpdir(), "ao-remotes-"));
	const path = join(dir, "remotes.json");
	if (contents !== undefined) {
		await writeFile(path, contents, "utf8");
		await chmod(path, mode);
	}
	return path;
}

const TWO_HOSTS = '{"remotes":[{"label":"workbox","url":"http://192.0.2.1:1","password":"old"},{"label":"mini","url":"http://192.0.2.9:9","password":"m"}]}';
describe("POSIX credential mode rejection", () => {
	it("refuses a file readable by others, naming the fix", async () => {
		const path = await tempFile('{"remotes":[]}', 0o644);
		await expect(readRemotes(path)).rejects.toBeInstanceOf(RemotesFilePermissionError);
		await expect(readRemotes(path)).rejects.toThrow(/chmod 600/);
	});
	it("refuses to touch a file others can read", async () => {
		const path = await tempFile(TWO_HOSTS, 0o644);
		await expect(updateRemote(path, "http://192.0.2.1:1", { password: "rotated" })).rejects.toBeInstanceOf(
			RemotesFilePermissionError,
		);
		// And left the passwords exactly where they were.
		expect(await readFile(path, "utf8")).toBe(TWO_HOSTS);
	});
	it("refuses to touch a file others can read", async () => {
		const path = await tempFile(TWO_HOSTS, 0o644);
		await expect(removeRemote(path, "http://192.0.2.1:1")).rejects.toBeInstanceOf(RemotesFilePermissionError);
		expect(await readFile(path, "utf8")).toBe(TWO_HOSTS);
	});
});
