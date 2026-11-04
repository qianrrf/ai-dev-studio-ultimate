import { readFileSync, writeFileSync } from "node:fs";
import { basename } from "node:path";
import { zipSync } from "fflate";

// Valid ZIP bytes, without host zip commands, temporary copies or local timestamps.
export function writeFixtureZip(source, destination) {
	writeFileSync(destination, zipSync({
		[basename(source)]: [readFileSync(source), { mtime: new Date("2026-01-01T00:00:00Z"), level: 0 }],
	}));
}
