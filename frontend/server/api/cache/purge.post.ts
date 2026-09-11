import { existsSync } from 'node:fs';
import { readdir, rm } from 'node:fs/promises';
import { join } from 'node:path';

// Путь к ISR-кэшу (соответствует nitro.storage['nitro:routes'].base в nuxt.config)
const CACHE_DIR = process.cwd() + '/.data/nitro-routes';

export default defineEventHandler(async (event) => {
  const body = await readBody<{ paths: string[] }>(event);

  if (!body?.paths || !Array.isArray(body.paths) || body.paths.length === 0) {
    throw createError({
      statusCode: 400,
      message: 'paths must be a non-empty array of strings',
    });
  }

  if (!existsSync(CACHE_DIR)) {
    return {
      success: true,
      purged: [],
      requested: body.paths,
      message: 'ISR cache directory not found — check nitro.storage config and routeRules',
      cacheDir: CACHE_DIR,
    };
  }

  let files: string[] = [];
  try {
    // Рекурсивно читаем все файлы (Nitro может создавать поддиректории)
    files = await readAllFiles(CACHE_DIR);
  } catch (err) {
    return {
      success: false,
      error: `Failed to read cache: ${err}`,
      purged: [],
      requested: body.paths,
    };
  }

  // Паттерны для фильтрации
  const patterns: string[] = [];
  for (const path of body.paths) {
    if (path === '/' || path === '') {
      patterns.push('index');
    } else {
      const clean = path.replace(/^\//, '');
      const heroMatch = clean.match(/^heroes\/([0-9a-f-]+)/i);
      if (heroMatch) {
        const uuidHex = heroMatch[1]?.replace(/-/g, '');
        patterns.push(`heroes${uuidHex?.slice(0, 10)}`);
        patterns.push('_payload');
      } else if (clean === 'heroes') {
        patterns.push('heroes');
        patterns.push('_payload');
      } else {
        patterns.push(clean.replace(/\//g, ''));
      }
    }
  }

  const toDelete = files.filter((f) => patterns.some((p) => f.includes(p)));

  const deleted: string[] = [];
  for (const file of toDelete) {
    try {
      await rm(join(CACHE_DIR, file), { force: true });
      deleted.push(file);
    } catch (err) {
      console.error(`[cache-purge] Failed to delete ${file}:`, err);
    }
  }

  return {
    success: true,
    purged: deleted,
    requested: body.paths,
    patterns,
    totalScanned: files.length,
    cacheDir: CACHE_DIR,
  };
});

// Рекурсивное чтение файлов (с учётом поддиректорий)
async function readAllFiles(dir: string, prefix = ''): Promise<string[]> {
  const entries = await readdir(dir, { withFileTypes: true });
  const files: string[] = [];
  for (const entry of entries) {
    const relPath = prefix ? `${prefix}/${entry.name}` : entry.name;
    if (entry.isDirectory()) {
      files.push(...(await readAllFiles(join(dir, entry.name), relPath)));
    } else {
      files.push(relPath);
    }
  }
  return files;
}
