<script setup lang="ts">
  const props = defineProps<{
    totalHeroes: number;
    totalConflicts: number;
  }>();
</script>

<template>
  <div class="home-og">
    <!-- Фоновый градиент -->
    <div class="home-og__bg" />

    <div class="home-og__content">
      <div class="home-og__header">
        <p class="home-og__kicker">
          <span class="home-og__flame" />
          КНИГА ПАМЯТИ
        </p>
        <h1 class="home-og__title">
          Вечная <em>память</em><br />
          героям
        </h1>
      </div>

      <p class="home-og__slogan">
        География разная. Подвиг один.<br />
        <span class="home-og__slogan-accent">Память вечна.</span>
      </p>

      <div class="home-og__stats">
        <div class="home-og__stat">
          <span class="home-og__stat-value">{{ totalHeroes }}</span>
          <span class="home-og__stat-label">имён в реестре</span>
        </div>
        <div class="home-og__stat">
          <span class="home-og__stat-value">{{ totalConflicts }}</span>
          <span class="home-og__stat-label">конфликтов</span>
        </div>
      </div>

      <p class="home-og__lead">
        ВОВ · Афган · Чечня · Вьетнам · Сирия · Африка · Новороссия · Спецоперации
      </p>
    </div>

    <!-- Нижний баннер -->
    <div class="home-og__footer">
      <span>вежливые.рус</span>
      <span class="home-og__motto">Никто не забыт, ничто не забыто</span>
    </div>
  </div>
</template>

<style scoped>
  .home-og {
    width: 1200px;
    height: 630px;
    position: relative;
    font-family: 'Golos Text', sans-serif;
    background: #f2f3f5;
    overflow: hidden;
  }

  .home-og__bg {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(60rem 30rem at 50% -10%, rgba(163, 22, 33, 0.07), transparent 60%),
      radial-gradient(40rem 24rem at 85% 110%, rgba(176, 141, 87, 0.09), transparent 60%),
      linear-gradient(180deg, #f4f5f7 0%, #eef0f2 100%);
  }

  .home-og__content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 60px;
    height: calc(100% - 80px);
  }

  .home-og__header {
    margin-bottom: 32px;
  }

  .home-og__kicker {
    display: inline-flex;
    align-items: center;
    gap: 12px;
    font-size: 16px;
    letter-spacing: 0.2em;
    text-transform: uppercase;
    color: #7a5c2e;
    font-weight: 600;
    margin: 0 0 16px;
  }

  .home-og__flame {
    width: 13px;
    height: 19px;
    background: radial-gradient(ellipse at 50% 78%, #f0a13c, #a31621 74%, transparent);
    border-radius: 50% 50% 46% 46% / 62% 62% 38% 38%;
  }

  .home-og__title {
    font-family: 'Playfair Display', serif;
    font-size: 72px;
    font-weight: 700;
    line-height: 1.02;
    margin: 0;
    color: #1a1c20;
    letter-spacing: -0.01em;
  }

  .home-og__title em {
    font-style: italic;
    color: #a31621;
  }

  .home-og__slogan {
    font-size: 28px;
    font-weight: 500;
    letter-spacing: 0.02em;
    margin: 0 0 40px;
    color: #1a1c20;
    line-height: 1.4;
  }

  .home-og__slogan-accent {
    color: #a31621;
    font-weight: 700;
  }

  .home-og__stats {
    display: flex;
    gap: 80px;
    margin-bottom: 40px;
  }

  .home-og__stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .home-og__stat-value {
    font-family: 'Playfair Display', serif;
    font-size: 56px;
    font-weight: 700;
    color: #1a1c20;
    line-height: 1;
  }

  .home-og__stat-label {
    font-size: 14px;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: #7a5c2e;
    font-weight: 600;
  }

  .home-og__lead {
    font-size: 18px;
    color: #5c616b;
    margin: 0;
    max-width: 800px;
    line-height: 1.6;
  }

  .home-og__footer {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 80px;
    background: rgba(0, 0, 0, 0.05);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 60px;
    font-size: 18px;
    color: #1a1c20;
    letter-spacing: 0.08em;
    border-top: 2px solid #7a5c2e;
  }

  .home-og__motto {
    font-family: 'Playfair Display', serif;
    font-style: italic;
    color: #a31621;
  }
</style>