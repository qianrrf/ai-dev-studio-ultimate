import { readFileSync, readdirSync, existsSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const lock = JSON.parse(readFileSync(path.join(root, "package-lock.json"), "utf8"));
const sections = ["Third-party package license texts from the locked build dependencies.\n"];
for (const [directory, entry] of Object.entries(lock.packages)) {
	if (!directory.startsWith("node_modules/")) continue;
	const location = path.join(root, directory);
	if (!existsSync(location)) continue;
	const names = readdirSync(location).filter(name => /^(licen[cs]e|copying|notice)(\.|$)/i.test(name));
	for (const name of names) {
		const file = path.join(location, name);
		try {
			sections.push(`\n--- ${directory} @ ${entry.version ?? "local"} / ${name} ---\n${readFileSync(file, "utf8")}\n`);
		} catch (error) {
			if (error.code !== "EISDIR") throw error;
		}
	}
}
writeFileSync(path.join(root, "assets", "THIRD_PARTY_LICENSES.txt"), sections.join(""));
