import { mkdirSync } from 'node:fs';
import { join } from 'node:path';

// ═══════════════════════════════════════════════════════════
// Server plugin: гарантирует, что директория ISR-кэша существует
// до первого запроса. Nitro иногда не создаёт её сам при fs-драйвере.
// ═══════════════════════════════════════════════════════════
export default defineNitroPlugin(() => {
  const cacheDir = join(process.cwd(), '.data', 'nitro-routes');
  try {
    mkdirSync(cacheDir, { recursive: true });
    console.log(`[cache] Directory ensured: ${cacheDir}`);
  } catch (err) {
    console.error(`[cache] Failed to create directory ${cacheDir}:`, err);
  }
});
