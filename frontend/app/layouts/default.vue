<script setup lang="ts">
  const route = useRoute();

  const { data: conflicts } = await useConflictsCatalog();
  const { data: heroesData } = await useHeroesPage(route.path === "/");

  const total = computed(() => heroesData.value?.total ?? 0);

  // Состояние мобильного меню
  const drawerVisible = ref(false);
  const closeDrawer = () => { drawerVisible.value = false; };
</script>

<template>
  <div class="shell">
    <header class="site-header">
      <NuxtLink to="/" class="site-header__brand">
        <img src="/logo_v5_full_gor_rwb.svg" alt="logo" class="shell_logo" fetchpriority="high">
      </NuxtLink>

      <!-- Десктоп: счётчики и навигация -->
      <div class="site-header__status site-header__status--desktop">
        <div>
          <span class="site-header__name">имён в реестре: </span>
          <Badge :value="total ?? 0"></Badge>
        </div>

        <div>
          <span class="site-header__name">конфликтов: </span>
          <Badge :value="conflicts?.length ?? 0"></Badge>
        </div>
      </div>

      <nav class="site-header__nav site-header__nav--desktop">
        <NuxtLink to="/">Книга памяти</NuxtLink>

        <NuxtLink to="/about">О проекте</NuxtLink>

        <NuxtLink to="/submit" class="site-header__submit">
          <i class="pi pi-user-plus" />
          Сообщить о герое
        </NuxtLink>

        <NuxtLink to="/contacts">Контакты</NuxtLink>

        <NuxtLink to="/admin">Служебный вход</NuxtLink>
      </nav>

      <!-- Мобилка: бургер -->
      <button class="site-header__burger" type="button" aria-label="Открыть меню" @click="drawerVisible = true">
        <i class="pi pi-bars" />
      </button>
    </header>

    <!-- Мобильное меню (Drawer) -->
    <Drawer v-model:visible="drawerVisible" position="right" class="site-drawer" header="Меню">
      <div class="site-drawer__stats">
        <div class="site-drawer__stat">
          <span class="site-drawer__stat-label">Имён в реестре</span>
          <Badge :value="total ?? 0" size="large" />
        </div>

        <div class="site-drawer__stat">
          <span class="site-drawer__stat-label">Конфликтов</span>
          <Badge :value="conflicts?.length ?? 0" size="large" />
        </div>
      </div>

      <nav class="site-drawer__nav">
        <NuxtLink to="/" class="site-drawer__link" @click="closeDrawer">
          <i class="pi pi-book" />
          Книга памяти
        </NuxtLink>

        <NuxtLink to="/about" class="site-drawer__link" @click="closeDrawer">
          <i class="pi pi-info-circle" />
          О проекте
        </NuxtLink>

        <NuxtLink to="/submit" class="site-drawer__link" @click="closeDrawer">
          <i class="pi pi-user-plus" />
          Сообщить о герое
        </NuxtLink>

        <NuxtLink to="/contacts" class="site-drawer__link" @click="closeDrawer">
          <i class="pi pi-envelope" />
          Контакты
        </NuxtLink>

        <NuxtLink to="/admin" class="site-drawer__link" @click="closeDrawer">
          <i class="pi pi-lock" />
          Служебный вход
        </NuxtLink>
      </nav>
    </Drawer>

    <slot />

    <footer class="site-footer">
      <hr class="rule" />
      <p class="site-footer__motto">Никто не забыт, ничто не забыто</p>
      <p class="site-footer__years">Глобальные и локальные конфликты · Спецоперации</p>
    </footer>
  </div>

  <Toast />
</template>

<style scoped>
  .shell {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  /* ========================================================================= */
  /* Header                                                                    */
  /* ========================================================================= */
  .site-header {
    display: grid;
    grid-template-columns: 1fr auto auto auto;
    /* brand | status | nav | burger */
    align-items: center;
    gap: 1rem;
    padding: 0.5rem 2rem;
    border-bottom: 1px solid var(--emh-line);
    background: rgba(255, 255, 255, 0.72);
    backdrop-filter: blur(6px);
    position: sticky;
    top: 0;
    z-index: 20;
  }

  .site-header__brand {
    display: inline-flex;
    align-items: center;
    gap: 0.65rem;
    color: var(--emh-ink);
  }

  .site-header__status {
    display: grid;
    grid-template-columns: repeat(2, auto);
    gap: 0.5rem;
  }

  .site-header__name {
    font-family: var(--font-display);
    font-size: 0.92rem;
    font-weight: 700;
    letter-spacing: 0.02em;
    color: var(--emh-muted);
  }

  .site-header__nav {
    display: flex;
    gap: 1.6rem;
    font-size: 0.92rem;
  }

  .site-header__nav a {
    color: var(--emh-muted);
    transition: color 0.2s ease;
  }

  .site-header__nav a:hover,
  .site-header__nav a.router-link-active {
    color: var(--emh-crimson);
  }

  /* Бургер — скрыт на десктопе */
  .site-header__burger {
    display: none;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border: 1px solid var(--emh-line);
    border-radius: 6px;
    background: transparent;
    color: var(--emh-ink);
    font-size: 1.1rem;
    cursor: pointer;
    transition: background 0.2s, border-color 0.2s;
  }

  .site-header__burger:hover {
    background: var(--emh-bg);
    border-color: var(--emh-bronze);
    color: var(--emh-crimson);
  }

  /* ========================================================================= */
  /* Drawer (мобильное меню)                                                   */
  /* ========================================================================= */
  .site-drawer__stats {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    padding: 1rem 0 1.5rem;
    border-bottom: 1px solid var(--emh-line);
    margin-bottom: 1rem;
  }

  .site-drawer__stat {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    align-items: flex-start;
  }

  .site-drawer__stat-label {
    font-family: var(--font-display);
    font-size: 0.78rem;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--emh-bronze);
    font-weight: 600;
  }

  .site-drawer__nav {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .site-drawer__link {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.85rem 0.75rem;
    border-radius: 6px;
    color: var(--emh-ink);
    font-size: 1rem;
    font-weight: 500;
    transition: background 0.2s, color 0.2s;
  }

  .site-drawer__link i {
    color: var(--emh-bronze);
    font-size: 1.05rem;
  }

  .site-drawer__link:hover,
  .site-drawer__link.router-link-active {
    background: rgba(176, 141, 87, 0.08);
    color: var(--emh-crimson);
  }

  .site-drawer__link.router-link-active i {
    color: var(--emh-crimson);
  }

  /* ========================================================================= */
  /* Footer                                                                    */
  /* ========================================================================= */
  .site-footer {
    margin-top: auto;
    padding: 2.5rem 2rem 3rem;
  }

  .site-footer__motto {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 1.15rem;
    color: var(--emh-ink);
    margin: 1.4rem 0 0.3rem;
  }

  .site-footer__years {
    font-size: 0.85rem;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: var(--emh-bronze);
  }

  /* ========================================================================= */
  /* Мобильная адаптация (< 768px)                                             */
  /* ========================================================================= */
  @media (max-width: 768px) {
    .site-header {
      grid-template-columns: 1fr auto;
      /* brand | burger */
      padding: 0.5rem 1rem;
    }

    /* Скрываем десктопные блоки */
    .site-header__status--desktop,
    .site-header__nav--desktop {
      display: none;
    }

    /* Показываем бургер */
    .site-header__burger {
      display: flex;
    }

    .site-footer {
      padding: 2rem 1rem 2.5rem;
    }
  }

  /* Очень маленькие экраны — логотип поменьше */
  @media (max-width: 400px) {
    .shell_logo {
      height: 38px;
    }
  }
</style>