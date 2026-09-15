<script setup lang="ts">
  /**
   * Layout админки под PrimeVue 4.5.5.
   *
   * Адаптивность без хаков:
   *  - Десктоп (≥1024px): статичный <aside class="admin-side"> в потоке документа.
   *  - Мобильные (<1024px): Drawer. В PrimeVue 4 он заменил Sidebar
   *    (в PrimeVue 5 Sidebar вернули — нам он недоступен).
   *
   * Содержимое панели вынесено в AdminSidebarContent.vue, чтобы не дублировать
   * разметку между <aside> и Drawer.
   */
  import AdminSidebarContent from '~/components/admin/AdminSidebarContent.vue';
  import { PublicationStatus } from '~/sdk/emh/v1/enums_emh_pb';
  import HamburgerMenuIcon from '~icons/mdi/hamburger-menu?width=2em&height=2em';

  const { submissionAdmin } = useApi()

  // ---------------------------------------------------------------------------
  // Адаптивность
  // ---------------------------------------------------------------------------
  const drawerVisible = ref(false)
  const isMobile = ref(false)

  const updateMobile = () => {
    isMobile.value = window.innerWidth < 1024
  }

  onMounted(() => {
    updateMobile()
    window.addEventListener('resize', updateMobile)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', updateMobile)
  })

  // При переходе на десктоп гасим drawer, чтобы он не «всплыл» при возврате на мобильный
  watch(isMobile, (mobile) => {
    if (!mobile) drawerVisible.value = false
  })

  // Клик по пункту меню в мобильном drawer закрывает его
  const closeDrawer = () => {
    drawerVisible.value = false
  }

  // ---------------------------------------------------------------------------
  // Счётчик заявок на модерации (DRAFT).
  // Один запрос на layout — в <aside> и Drawer значение передаётся пропом.
  // ---------------------------------------------------------------------------
  const pendingCount = ref(0)

  onMounted(async () => {
    try {
      const res = await submissionAdmin.listSubmissions({
        pagination: { pageSize: 1, cursor: '' },
        status: PublicationStatus.DRAFT,
      })
      // total_count может быть 0, если бэкенд его не считает — тогда пункт без счётчика
      pendingCount.value = Number(res.pagination?.totalCount ?? 0)
    } catch {
      /* игнорим — пункт меню просто будет без счётчика */
    }
  })
</script>

<template>
  <div class="admin-shell">
    <!-- ================================================================== -->
    <!-- Десктоп: статичная боковая панель (без PrimeVue-телепорта)          -->
    <!-- ================================================================== -->
    <aside class="admin-side">
      <AdminSidebarContent :pending-count="pendingCount" @navigate="closeDrawer" />
    </aside>

    <!-- ================================================================== -->
    <!-- Мобильные: Drawer (PrimeVue 4.5.5)                                  -->
    <!-- ================================================================== -->
    <Drawer v-model:visible="drawerVisible" position="left" :modal="true" :dismissable="true" :show-close-icon="true"
      :style="{ width: '280px' }" class="admin-drawer">
      <AdminSidebarContent :pending-count="pendingCount" @navigate="closeDrawer" />
    </Drawer>

    <!-- ================================================================== -->
    <!-- Основная область контента                                          -->
    <!-- ================================================================== -->
    <main class="admin-main">
      <header class="admin-header">
        <Button severity="secondary" class="admin-header__burger" aria-label="Меню" @click="drawerVisible = true">
          <HamburgerMenuIcon />
        </Button>

        <h1 class="admin-header__title">Панель управления</h1>
      </header>

      <div class="admin-content">
        <slot />
      </div>
    </main>

    <!-- Глобальные оверлеи PrimeVue 4.5.5 -->
    <Toast />
    <ConfirmDialog />
  </div>
</template>

<style src="~/assets/css/admin.css"></style>
<style scoped>

  /* ========================================================================= */
  /* Shell                                                                      */
  /* ========================================================================= */
  .admin-shell {
    min-height: 100vh;
    background: var(--p-surface-50);
    color: var(--p-text-color);
  }

  /* ========================================================================= */
  /* Статичный сайдбар — виден только на десктопе                              */
  /* ========================================================================= */
  .admin-side {
    display: none;
  }

  @media (min-width: 1024px) {
    .admin-side {
      display: block;
      position: fixed;
      top: 0;
      left: 0;
      bottom: 0;
      width: 280px;
      background: var(--p-surface-0);
      border-right: 1px solid var(--p-surface-200);
      z-index: 100;
    }
  }

  /* ========================================================================= */
  /* Основная область                                                           */
  /* ========================================================================= */
  .admin-main {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  @media (min-width: 1024px) {
    .admin-main {
      margin-left: 280px;
    }
  }

  .admin-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    height: 3.75rem;
    padding: 0.5rem 1rem;
    background: var(--p-surface-0);
    border-bottom: 1px solid var(--p-surface-200);
    position: sticky;
    top: 0;
    z-index: 10;
  }

  /* Бургер виден только там, где работает Drawer */
  .admin-header__burger {
    display: none;
    flex-shrink: 0;
  }

  @media (max-width: 1023px) {
    .admin-header__burger {
      display: inline-flex;
    }
  }

  .admin-header__title {
    font-size: 1rem;
    font-weight: 600;
    margin: 0;
  }

  .admin-content {
    flex: 1;
    padding: 1.5rem;
  }
</style>

<style>
  .admin-drawer {
    width: 280px;
    max-width: 85vw;
  }

  .admin-drawer .p-drawer-header {
    padding: 0.4rem 0.5rem 0.4rem 0;
    border-bottom: 1px solid var(--p-surface-200);
  }

  .admin-drawer .p-drawer-content {
    padding: 0;
    overflow-y: auto;
  }
</style>