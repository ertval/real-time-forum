import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, test } from 'vitest';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const spaRoot = path.resolve(testDir, '../../../');
const forbiddenImportFragment = '../../../web/static/';

async function collectJavaScriptFiles(rootDir) {
	const entries = await readdir(rootDir, { withFileTypes: true });
	const nested = await Promise.all(
		entries.map(async (entry) => {
			const absolutePath = path.join(rootDir, entry.name);
			if (entry.isDirectory()) {
				return collectJavaScriptFiles(absolutePath);
			}

			if (!entry.isFile() || !absolutePath.endsWith('.js')) {
				return [];
			}

			return [absolutePath];
		}),
	);

	return nested.flat();
}

async function findForbiddenImports() {
	const files = await collectJavaScriptFiles(spaRoot);
	const offenders = [];
	const importPattern =
		/(?:import\s+(?:[^'"]+?\s+from\s+)?|export\s+[^'"]*?\s+from\s+)['"]([^'"]+)['"]/g;

	for (const filePath of files) {
		const source = await readFile(filePath, 'utf8');
		const matches = source.matchAll(importPattern);

		for (const match of matches) {
			const specifier = match[1] ?? '';
			if (!specifier.includes(forbiddenImportFragment)) {
				continue;
			}

			offenders.push({
				file: path.relative(spaRoot, filePath),
				specifier,
			});
		}
	}

	return offenders;
}

describe('SPA import boundaries', () => {
	test('does not import browser-invalid modules from ../../../web/static/', async () => {
		const offenders = await findForbiddenImports();

		expect(offenders).toEqual([]);
	});
});
