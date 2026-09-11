import { existsSync, statSync } from 'node:fs';
import { readdir } from 'node:fs/promises';

const CACHE_DIR = process.cwd() + '/.data/nitro-routes';

export default defineEventHandler(async () => {
  // Storage API
  const storage = useStorage('nitro:routes');
  const storageKeys = await storage.getKeys();

  // Filesystem
  let fileKeys: string[] = [];
  let cacheDirInfo: any = { exists: false };

  if (existsSync(CACHE_DIR)) {
    cacheDirInfo = {
      exists: true,
      writable: (() => {
        try {
          const stat = statSync(CACHE_DIR);
          return (stat.mode & 0o200) !== 0;
        } catch {
          return false;
        }
      })(),
    };
    try {
      fileKeys = await readdir(CACHE_DIR, { recursive: true });
    } catch (err) {
      cacheDirInfo.readError = String(err);
    }
  }

  // Альтернативные пути
  const altPaths = [
    process.cwd() + '/.output/server/cache',
    process.cwd() + '/.data',
    process.cwd() + '/.nuxt',
    process.cwd() + '/tmp',
  ].map((p) => ({
    path: p,
    exists: existsSync(p),
  }));

  // Route rules — проверяем, что реально настроено
  const routeRulesCheck = {
    '/heroes/**': 'should match hero cards',
    '/heroes': 'should match list',
    '/': 'should match home',
  };

  return {
    storage: {
      namespace: 'nitro:routes',
      totalKeys: storageKeys.length,
      sampleKeys: storageKeys.slice(0, 20),
    },
    filesystem: {
      primaryPath: CACHE_DIR,
      ...cacheDirInfo,
      totalFiles: fileKeys.length,
      files: fileKeys.slice(0, 20),
    },
    alternativePaths: altPaths,
    routeRulesCheck,
    cwd: process.cwd(),
    nodeEnv: process.env.NODE_ENV,
  };
});
