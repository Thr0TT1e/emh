import type { Api } from '~/lib/api';

declare module '#app' {
  interface NuxtApp {
    $api: Api;
  }
}

export {};
