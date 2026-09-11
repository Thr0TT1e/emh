import type { Api } from '~/lib/api';

// Типизированный доступ к Connect-клиентам.
export const useApi = (): Api => useNuxtApp().$api;
