<script setup lang="ts">
  /**
   * AdminSidebarContent — содержимое боковой панели админки:
   * бренд, навигация, профиль и выход.
   *
   * Используется в layouts/admin.vue в двух местах:
   *  - статичный <aside> на десктопе (PrimeVue здесь не нужен вовсе);
   *  - Drawer на мобильных (PrimeVue 4.5.5).
   */
  import { useAuth } from '~/composables/useAuth';
  import ExternalLinkIcon from '~icons/mdi/external-link?width=1.25em&height=1.25em';
  import LogoutIcon from '~icons/mdi/logout?width=1.25em&height=1.25em';

  type AdminNavItem = {
    label: string
    to: string
    viewBox: string
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
        {
          label: 'Сводка',
          to: '/admin',
          viewBox: '0 0 32 32',
          icon: `<path d="M16.612 2.214a1.01 1.01 0 0 0-1.242 0L1 13.419l1.243 1.572L4 13.621V26a2.004 2.004 0 0 0 2 2h20a2.004 2.004 0 0 0 2-2V13.63L29.757 15L31 13.428ZM18 26h-4v-8h4Zm2 0v-8a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v8H6V12.062l10-7.79l10 7.8V26Z" class="cuyn6t"/>`,
          exact: true
        },
        {
          label: 'Герои',
          to: '/admin/heroes',
          viewBox: '0 0 24 24',
          icon: `<path fill="currentColor" d="M12 5a3.5 3.5 0 0 0-3.5 3.5A3.5 3.5 0 0 0 12 12a3.5 3.5 0 0 0 3.5-3.5A3.5 3.5 0 0 0 12 5m0 2a1.5 1.5 0 0 1 1.5 1.5A1.5 1.5 0 0 1 12 10a1.5 1.5 0 0 1-1.5-1.5A1.5 1.5 0 0 1 12 7M5.5 8A2.5 2.5 0 0 0 3 10.5c0 .94.53 1.75 1.29 2.18c.36.2.77.32 1.21.32s.85-.12 1.21-.32c.37-.21.68-.51.91-.87A5.42 5.42 0 0 1 6.5 8.5v-.28c-.3-.14-.64-.22-1-.22m13 0c-.36 0-.7.08-1 .22v.28c0 1.2-.39 2.36-1.12 3.31c.12.19.25.34.4.49a2.48 2.48 0 0 0 1.72.7c.44 0 .85-.12 1.21-.32c.76-.43 1.29-1.24 1.29-2.18A2.5 2.5 0 0 0 18.5 8M12 14c-2.34 0-7 1.17-7 3.5V19h14v-1.5c0-2.33-4.66-3.5-7-3.5m-7.29.55C2.78 14.78 0 15.76 0 17.5V19h3v-1.93c0-1.01.69-1.85 1.71-2.52m14.58 0c1.02.67 1.71 1.51 1.71 2.52V19h3v-1.5c0-1.74-2.78-2.72-4.71-2.95M12 16c1.53 0 3.24.5 4.23 1H7.77c.99-.5 2.7-1 4.23-1"/>`
        },
        {
          label: 'API-ключи',
          to: '/admin/keys',
          viewBox: '0 0 24 24',
          icon: `<path fill="currentColor" d="M12.66 13.67c-.34.33-.73.62-1.16.83V21l-2 2l-2-2l2-1.71L8 18l1.5-1.29l-2-1.71v-.5a4.42 4.42 0 0 1-2.5-4C5 8 7 6 9.5 6h.11c-.02.07-.07.12-.11.18c-.27.61-.42 1.25-.47 1.9A1.498 1.498 0 0 0 9.5 11h.1c.64 1.25 1.74 2.2 3.06 2.67M16 6c0-.63-.1-1.25-.28-1.82c1.34.38 2.49 1.37 3.01 2.78c.6 1.66.16 3.43-.98 4.63L20 17.68l-1.22 2.57l-2.56-1.2l1.28-2.29l-1.84-.7l.97-1.72l-2.47-.93l-.16-.46a4.48 4.48 0 0 1-3.73-2.91a4.51 4.51 0 0 1 2.69-5.77c.18-.06.37-.1.54-.14A3.95 3.95 0 0 0 10 2C7.79 2 6 3.79 6 6c0 .09 0 .17.03.26c-.33.27-.63.56-.88.89C5.06 6.78 5 6.4 5 6c0-2.76 2.24-5 5-5s5 2.24 5 5c0 1.16-.4 2.21-1.06 3.06C16.08 8.88 16 6 16 6m-3.19 2.1c.06.17.15.31.25.44c.56-.66.91-1.5.94-2.43c-.11.02-.2.04-.3.07c-.78.29-1.2 1.15-.89 1.92"/>`
        },
        {
          label: 'Заявки',
          to: '/admin/submissions',
          viewBox: '0 0 24 24',
          icon: `<path fill="currentColor" d="M19 3a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2zM5 10v2h4.4c-.6-.53-1.06-1.22-1.27-2zm14 2v-2h-3.13c-.21.78-.67 1.47-1.27 2zm0-4V5H5v3h5v1c0 1.07.93 2 2 2s2-.93 2-2V8zm2 11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4h7v1c0 1.07.93 2 2 2s2-.93 2-2v-1h7zM5 17v2h4.4c-.6-.53-1.06-1.22-1.27-2zm14 2v-2h-3.13c-.21.78-.67 1.47-1.27 2z"/>`,
          badge: props.pendingCount || undefined,
        },
      ],
    },
    {
      label: 'Справочники',
      items: [
        {
          label: 'Локации',
          to: '/admin/locations',
          viewBox: '0 0 32 32',
          icon: `<path fill="currentColor" d="m16 24l-6.09-8.6A8.14 8.14 0 0 1 16 2a8.08 8.08 0 0 1 8 8.13a8.2 8.2 0 0 1-1.8 5.13Zm0-20a6.07 6.07 0 0 0-6 6.13a6.2 6.2 0 0 0 1.49 4L16 20.52L20.63 14A6.24 6.24 0 0 0 22 10.13A6.07 6.07 0 0 0 16 4"/><circle cx="16" cy="9" r="2" fill="currentColor"/><path fill="currentColor" d="M28 12h-2v2h2v14H4V14h2v-2H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h24a2 2 0 0 0 2-2V14a2 2 0 0 0-2-2"/>`
        },
        {
          label: 'Конфликты',
          to: '/admin/conflicts',
          viewBox: '0 0 24 24',
          icon: `<path fill="currentColor" d="M7 5h16v4h-1v1h-6a1 1 0 0 0-1 1v1a2 2 0 0 1-2 2H9.62c-.38 0-.73.22-.9.56l-2.45 4.89c-.17.34-.51.55-.89.55H2s-3 0 1-6c0 0 3-4-1-4V5h1l.5-1h3zm7 7v-1a1 1 0 0 0-1-1h-1s-1 1 0 2a2 2 0 0 1-2-2a1 1 0 0 0-1 1v1a1 1 0 0 0 1 1h3a1 1 0 0 0 1-1"/>`
        },
        {
          label: 'Награды',
          to: '/admin/awards',
          viewBox: '0 0 24 24',
          icon: `<path fill="currentColor" d="M18 2c-.9 0-2 1-2 2H8c0-1-1.1-2-2-2H2v9c0 1 1 2 2 2h2.2c.4 2 1.7 3.7 4.8 4v2.08C8 19.54 8 22 8 22h8s0-2.46-3-2.92V17c3.1-.3 4.4-2 4.8-4H20c1 0 2-1 2-2V2zM6 11H4V4h2zm10 .5c0 1.93-.58 3.5-4 3.5c-3.41 0-4-1.57-4-3.5V6h8zm4-.5h-2V4h2z"/>`
        },
        {
          label: 'LLM-провайдеры',
          to: '/admin/llm',
          viewBox: '0 0 32 32',
          icon: `<path d="M17 11h3v10h-3v2h8v-2h-3V11h3V9h-8zm-4-2H9c-1.103 0-2 .897-2 2v12h2v-5h4v5h2V11c0-1.103-.897-2-2-2m-4 7v-5h4v5z" class="cuyn6t"/>`
        },
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
              <svg xmlns="http://www.w3.org/2000/svg" width="1.25em" height="1.25em" :viewBox="item.viewBox"
                v-html="item.icon" />
              <!-- <i :class="item.icon" class="admin-sidebar__link-icon" /> -->
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
        <ExternalLinkIcon />
        <span class="admin-sidebar__link-label">На сайт</span>
      </NuxtLink>

      <div v-if="claims" class="admin-sidebar__user">
        <Avatar :label="(claims.username ?? 'U').slice(0, 2).toUpperCase()" shape="circle" size="normal" />

        <div class="admin-sidebar__user-info">
          <div class="admin-sidebar__user-name">{{ claims.username }}</div>
          <div class="admin-sidebar__user-role">{{ claims.role ?? 'гость' }}</div>
        </div>
      </div>

      <Button @click="doLogout">
        <LogoutIcon />
        <span>Выйти</span>
      </Button>
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