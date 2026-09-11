export const useCachePurge = () => {
  const purge = async (paths: string | string[]) => {
    const pathsArray = Array.isArray(paths) ? paths : [paths]

    try {
      const result = await $fetch('/api/cache/purge', {
        method: 'POST',
        body: { paths: pathsArray },
      })

      console.log('[cache-purge]', result)
      return result
    } catch (error) {
      console.error('[cache-purge] Failed to purge cache:', error)
      throw error
    }
  }

  const purgeHero = async (heroId: string) => {
    return purge([
      `/heroes/${heroId}`,
      '/heroes',
      '/',
    ])
  }

  return {
    purge,
    purgeHero,
  }
}
