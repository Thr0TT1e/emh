<script setup lang="ts">
  import { useAuth } from "~/composables/useAuth";
  import { isRateLimited, isUnauthenticated } from "~/lib/errors";

  definePageMeta({ layout: false });

  const { isAuthenticated, login } = useAuth();
  const router = useRouter();
  const route = useRoute();

  if (isAuthenticated.value) {
    const next = typeof route.query.next === 'string' ? route.query.next : '/admin';
    router.replace(next);
  }

  const username = ref("");
  const password = ref("");
  const busy = ref(false);
  const errorMsg = ref("");
  const shakeKey = ref(0);

  // ---------------------------------------------------------------------------
  // Блокировка формы на N секунд после rate-limit (429).
  // ---------------------------------------------------------------------------
  const lockUntil = ref(0);
  const lockRemaining = ref(0);
  let lockTimer: ReturnType<typeof setInterval> | undefined;

  const isLocked = computed(() => lockRemaining.value > 0);

  const startLockTimer = (ms: number) => {
    lockUntil.value = Date.now() + ms;
    lockRemaining.value = Math.ceil(ms / 1000);
    clearInterval(lockTimer);
    lockTimer = setInterval(() => {
      const left = Math.max(0, Math.ceil((lockUntil.value - Date.now()) / 1000));
      lockRemaining.value = left;
      if (left <= 0) clearInterval(lockTimer);
    }, 250);
  };

  onBeforeUnmount(() => clearInterval(lockTimer));

  // Сообщение в query (?reason=expired) после редиректа interceptor'а.
  onMounted(() => {
    if (route.query.reason === 'expired') {
      errorMsg.value = "Сессия истекла. Войдите снова.";
      shakeKey.value++;
    }
  });

  const submit = async () => {
    if (isLocked.value) return;
    busy.value = true;
    errorMsg.value = "";
    try {
      await login(username.value, password.value);
      const next = typeof route.query.next === 'string' ? route.query.next : '/admin';
      router.push(next);
    } catch (err) {
      if (isRateLimited(err)) {
        errorMsg.value = "Слишком много попыток входа. Подождите 1 минуту.";
        startLockTimer(60_000);
      } else if (isUnauthenticated(err)) {
        errorMsg.value = "Неверный логин или пароль.";
      } else {
        errorMsg.value = "Ошибка соединения. Попробуйте ещё раз.";
      }
      shakeKey.value++;
    } finally {
      busy.value = false;
    }
  };

  useHead({ title: "Служебный вход — Вечная память героям" });
</script>

<template>
  <div class="login">
    <div class="login__panel">
      <Image class="login__logo" src="/logo_v5_full_vert_wbr.svg" alt="Вечная память героям" />
      <div class="login__kicker">Канцелярия</div>
      <div class="login__quote">
        «Никто не забыт, <br />и ничто не забыто»
      </div>
    </div>

    <div class="login__form-wrap">
      <Form @submit="submit">
        <h1 class="login__form-title">Служебный вход</h1>

        <p v-if="errorMsg" :key="shakeKey" class="login__error shake">{{ errorMsg }}</p>

        <div>
          <label for="login-name" class="afield__label">Имя пользователя</label>
          <InputText id="login-name" v-model="username" class="w-full" autocomplete="username" required
            :disabled="isLocked" />
        </div>

        <div>
          <label class="afield__label">Пароль</label>
          <Password v-model="password" autocomplete="current-password" required :disabled="isLocked" toggleMask fluid />
        </div>

        <Button type="submit" :loading="busy" :disabled="isLocked" class="w-full mt-3">
          {{ isLocked ? `Подождите ${lockRemaining} сек…` : 'Войти' }}
        </Button>

        <NuxtLink to="/" class="login__back mt-4">← Вернуться на сайт</NuxtLink>
      </Form>
    </div>
  </div>
</template>

<style scoped>
  .afield__pass {
    display: grid;
  }

  /* стили без изменений — те же, что были */
  .login {
    display: grid;
    grid-template-columns: 1fr 1fr;
    min-height: 100vh;
  }

  .login__panel {
    width: 100%;
    height: 100%;
    display: grid;
    grid-auto-rows: max-content;
    align-content: center;
    justify-items: center;
    background: radial-gradient(46rem 26rem at 20% 110%, rgba(163, 22, 33, .28), transparent 60%), radial-gradient(30rem 20rem at 80% -10%, rgba(176, 141, 87, .16), transparent 60%), #14161b;
    color: #f2f3f5;
  }

  .login__logo {
    width: 300px;
  }

  .login__kicker {
    height: 60px;
    align-content: center;
    font-size: 1rem;
    letter-spacing: .3em;
    text-transform: uppercase;
    color: var(--emh-bronze);
  }

  .login__quote {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 1.15rem;
    color: #9aa0ab;
  }

  .login__form-wrap {
    display: grid;
    grid-template-columns: minmax(300px, 31.25rem);
    padding: 0 20px;
    place-self: center;
    background: var(--emh-bg);
  }

  .login__form-title {
    font-size: 1.5rem;
    margin-bottom: 1.4rem;
  }

  .login__error {
    color: var(--emh-crimson);
    font-size: .9rem;
    background: rgba(163, 22, 33, .07);
    border: 1px solid rgba(163, 22, 33, .25);
    padding: .55rem .8rem;
    margin-bottom: 1rem;
  }

  .login__back {
    font-size: .88rem;
    color: var(--emh-muted);
  }

  @media (max-width: 820px) {
    .login {
      grid-template-columns: 1fr;
    }

    .login__panel {
      padding: 3rem 2rem;
    }

    .login__logo {
      width: 200px;
    }
  }
</style>