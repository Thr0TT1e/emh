<script setup lang="ts">
  /**
   * AdminSidebarContent — содержимое боковой панели админки:
   * бренд, навигация, профиль и выход.
   *
   * Используется в layouts/admin.vue в двух местах:
   *  - статичный <aside> на десктопе (PrimeVue здесь не нужен вовсе);
   *  - Drawer на мобильных (PrimeVue 4.5.5).
   */
  import { useAuth } from '~/composables/useAuth'

  type AdminNavItem = {
    label: string
    to: string
    icon: string
    exact?: boolean
    badge?: number
  }

  type AdminNavGroup = {
    label: string
    items: AdminNavItem[]
  }

  const props = defineProps<{ pendingCount?: number }>()
  const emit = defineEmits<{ navigate: [] }>()

  const { claims, logout } = useAuth()
  const router = useRouter()
  const route = useRoute()

  // ---------------------------------------------------------------------------
  // Структура меню (группа «Справочники» — Locations / Conflicts / Awards)
  // ---------------------------------------------------------------------------
  const navGroups = computed<AdminNavGroup[]>(() => [
    {
      label: 'Управление',
      items: [
        { label: 'Сводка', to: '/admin', icon: 'pi pi-home', exact: true },
        { label: 'Герои', to: '/admin/heroes', icon: 'pi pi-users' },
        { label: 'API-ключи', to: '/admin/keys', icon: 'pi pi-key' },
        {
          label: 'Заявки',
          to: '/admin/submissions',
          icon: 'pi pi-inbox',
          badge: props.pendingCount || undefined,
        },
      ],
    },
    {
      label: 'Справочники',
      items: [
        { label: 'Локации', to: '/admin/locations', icon: 'pi pi-map-marker' },
        { label: 'Конфликты', to: '/admin/conflicts', icon: 'pi pi-flag' },
        { label: 'Награды', to: '/admin/awards', icon: 'pi pi-trophy' },
        { label: 'LLM-провайдеры', to: '/admin/llm', icon: 'pi pi-bolt' },
      ],
    },
  ])

  // ---------------------------------------------------------------------------
  // Подсветка активного пункта
  // ---------------------------------------------------------------------------
  const isActive = (item: AdminNavItem): boolean => {
    if (item.exact) return route.path === item.to
    return route.path === item.to || route.path.startsWith(`${item.to}/`)
  }

  // ---------------------------------------------------------------------------
  // Выход
  // ---------------------------------------------------------------------------
  const doLogout = async () => {
    await logout();
    router.push("/admin/login");
  };
</script>

<template>
  <div class="admin-sidebar-content">
    <!-- Бренд -->
    <NuxtLink to="/admin" class="admin-sidebar__brand" @click="emit('navigate')">
      <img src="/logo_v5_full_gor_rwb.svg" alt="EMH" class="admin-sidebar__logo" />
      <div class="admin-sidebar__brand-text">
        <small>канцелярия</small>
      </div>
    </NuxtLink>

    <!-- Навигация -->
    <div class="admin-sidebar__nav">
      <div v-for="group in navGroups" :key="group.label" class="admin-sidebar__group">
        <div class="admin-sidebar__group-label">{{ group.label }}</div>

        <ul class="admin-sidebar__menu">
          <li v-for="item in group.items" :key="item.to">
            <NuxtLink :to="item.to" class="admin-sidebar__link" :class="{ 'is-active': isActive(item) }"
              @click="emit('navigate')">
              <i :class="item.icon" class="admin-sidebar__link-icon" />
              <span class="admin-sidebar__link-label">{{ item.label }}</span>
              <span v-if="item.badge" class="admin-sidebar__badge">{{ item.badge }}</span>
            </NuxtLink>
          </li>
        </ul>
      </div>
    </div>

    <!-- Футер: на сайт, профиль, выход -->
    <div class="admin-sidebar__foot">
      <NuxtLink to="/" class="admin-sidebar__link" target="_blank" rel="noopener">
        <i class="pi pi-external-link admin-sidebar__link-icon" />
        <span class="admin-sidebar__link-label">На сайт</span>
      </NuxtLink>

      <div v-if="claims" class="admin-sidebar__user">
        <Avatar :label="(claims.username ?? 'U').slice(0, 2).toUpperCase()" shape="circle" size="normal" />
        <div class="admin-sidebar__user-info">
          <div class="admin-sidebar__user-name">{{ claims.username }}</div>
          <div class="admin-sidebar__user-role">{{ claims.role ?? 'гость' }}</div>
        </div>
      </div>

      <button type="button" class="admin-sidebar__link admin-sidebar__logout" @click="doLogout">
        <i class="pi pi-sign-out admin-sidebar__link-icon" />
        <span class="admin-sidebar__link-label">Выйти</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
  .admin-sidebar-content {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr auto;
    background: var(--p-surface-0);
  }

  /* Бренд */
  .admin-sidebar__brand {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    column-gap: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: 8px;
    transition: background 0.15s;
  }

  .admin-sidebar__brand:hover {
    background: var(--p-surface-100);
  }

  .admin-sidebar__logo {
    width: 100%;
    height: 50px;
    object-fit: contain;
  }

  .admin-sidebar__brand-text {
    display: grid;
    place-items: center;
    line-height: 1.2;
  }

  .admin-sidebar__brand-text small {
    font-size: 1rem;
    color: var(--p-text-muted-color);
    letter-spacing: 0.08em;
    text-transform: lowercase;
  }

  /* Навигация */
  .admin-sidebar__nav {
    overflow-y: auto;
    min-height: 0;
    /* без этого grid-строка не даёт внутреннему скроллу работать */
    padding: 0.5rem 0.75rem;
  }

  .admin-sidebar__group {
    margin-bottom: 1.25rem;
  }

  .admin-sidebar__group-label {
    padding: 0.5rem 0.75rem 0.4rem;
    font-size: 0.68rem;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--p-text-muted-color);
    user-select: none;
  }

  .admin-sidebar__menu {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .admin-sidebar__link {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.55rem 0.75rem;
    color: var(--p-text-color);
    text-decoration: none;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 500;
    background: none;
    border: none;
    width: 100%;
    cursor: pointer;
    text-align: left;
    transition: background 0.15s, color 0.15s;
  }

  .admin-sidebar__link:hover {
    background: var(--p-surface-100);
  }

  .admin-sidebar__link.is-active {
    background: var(--p-primary-50, var(--p-surface-100));
    color: var(--p-primary-color);
  }

  .admin-sidebar__link.is-active .admin-sidebar__link-icon {
    color: var(--p-primary-color);
  }

  .admin-sidebar__link-icon {
    font-size: 1rem;
    width: 1.25rem;
    text-align: center;
    color: var(--p-text-muted-color);
    flex-shrink: 0;
  }

  .admin-sidebar__link-label {
    flex: 1;
  }

  .admin-sidebar__badge {
    margin-left: auto;
    padding: 0.1rem 0.5rem;
    font-size: 0.72rem;
    font-weight: 600;
    color: #fff;
    background: var(--emh-carmine, #bd2f3c);
    border-radius: 999px;
    min-width: 1.25rem;
    text-align: center;
    line-height: 1.3;
  }

  /* Футер */
  .admin-sidebar__foot {
    border-top: 1px solid var(--p-surface-200);
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .admin-sidebar__user {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    padding: 0.55rem 0.75rem;
    border-radius: 6px;
    background: var(--p-surface-50);
    margin-bottom: 0.25rem;
  }

  .admin-sidebar__user-info {
    min-width: 0;
  }

  .admin-sidebar__user-name {
    font-size: 0.88rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .admin-sidebar__user-role {
    font-size: 0.72rem;
    color: var(--p-text-muted-color);
    text-transform: lowercase;
  }

  .admin-sidebar__logout {
    color: var(--emh-carmine, #bd2f3c);
  }

  .admin-sidebar__logout:hover {
    background: rgba(189, 47, 60, 0.08);
  }

  .admin-sidebar__logout .admin-sidebar__link-icon {
    color: inherit;
  }
</style>