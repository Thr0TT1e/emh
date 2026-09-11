<script setup lang="ts">
  // Props из defineOgImage
  const props = defineProps<{
    fullName: string;
    birthYear?: string;
    deathYear?: string;
    rank?: string;
    conflicts?: string;
    photo?: string;
    nickname?: string;
  }>();
</script>

<template>
  <div class="hero-og">
    <!-- Фоновый градиент: гранит + отблеск пламени -->
    <div class="hero-og__bg" />

    <div class="hero-og__content">
      <!-- Левая колонка: фото -->
      <div v-if="photo" class="hero-og__photo">
        <img :src="photo" :alt="fullName" />
      </div>
      <div v-else class="hero-og__photo-empty">
        <span>ПОРТРЕТ НЕ СОХРАНИЛСЯ</span>
      </div>

      <!-- Правая колонка: информация -->
      <div class="hero-og__info">
        <p class="hero-og__kicker">
          <span class="hero-og__flame" />
          ВЕЧНАЯ ПАМЯТЬ ГЕРОЯМ
        </p>

        <h1 class="hero-og__name">{{ fullName }}</h1>

        <p v-if="nickname" class="hero-og__nickname">
          позывной «{{ nickname }}»
        </p>

        <p v-if="rank" class="hero-og__rank">{{ rank }}</p>

        <p v-if="birthYear || deathYear" class="hero-og__dates">
          <template v-if="birthYear">{{ birthYear }}</template>
          <template v-if="birthYear && deathYear"> — </template>
          <template v-if="deathYear">{{ deathYear }}</template>
        </p>

        <p v-if="conflicts" class="hero-og__conflicts">{{ conflicts }}</p>
      </div>
    </div>

    <!-- Нижний баннер с URL -->
    <div class="hero-og__footer">
      <span>вежливые.рус</span>
      <span class="hero-og__motto">Никто не забыт</span>
    </div>
  </div>
</template>

<style scoped>
  .hero-og {
    width: 1200px;
    height: 630px;
    position: relative;
    font-family: 'Golos Text', sans-serif;
    background: #1a1c20;
    overflow: hidden;
  }

  .hero-og__bg {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse at 20% 30%, rgba(163, 22, 33, 0.15), transparent 50%),
      radial-gradient(ellipse at 80% 70%, rgba(176, 141, 87, 0.12), transparent 50%),
      linear-gradient(135deg, #1a1c20 0%, #2a2d33 100%);
  }

  .hero-og__content {
    position: relative;
    z-index: 1;
    display: flex;
    gap: 60px;
    padding: 60px;
    height: calc(100% - 80px);
  }

  .hero-og__photo {
    flex-shrink: 0;
    width: 360px;
    height: 450px;
    border: 3px solid #7a5c2e;
    padding: 8px;
    background: #fff;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  .hero-og__photo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    filter: grayscale(0.3);
  }

  .hero-og__photo-empty {
    flex-shrink: 0;
    width: 360px;
    height: 450px;
    border: 3px solid #7a5c2e;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.05);
    color: #5c616b;
    font-size: 20px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    text-align: center;
    padding: 1rem;
  }

  .hero-og__info {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    color: #f2f3f5;
  }

  .hero-og__kicker {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 16px;
    letter-spacing: 0.2em;
    text-transform: uppercase;
    color: #7a5c2e;
    font-weight: 600;
    margin: 0 0 24px;
  }

  .hero-og__flame {
    width: 13px;
    height: 19px;
    background: radial-gradient(ellipse at 50% 78%, #f0a13c, #a31621 74%, transparent);
    border-radius: 50% 50% 46% 46% / 62% 62% 38% 38%;
  }

  .hero-og__name {
    font-family: 'Playfair Display', serif;
    font-size: 58px;
    font-weight: 700;
    line-height: 1.1;
    margin: 0 0 16px;
    color: #fff;
  }

  .hero-og__nickname {
    font-family: 'Playfair Display', serif;
    font-style: italic;
    font-size: 22px;
    color: #7a5c2e;
    margin: 0 0 16px;
  }

  .hero-og__rank {
    font-size: 20px;
    color: #7a5c2e;
    margin: 0 0 16px;
    letter-spacing: 0.05em;
  }

  .hero-og__dates {
    font-family: 'Playfair Display', serif;
    font-size: 32px;
    font-weight: 600;
    color: #a31621;
    margin: 0 0 24px;
    letter-spacing: 0.04em;
  }

  .hero-og__conflicts {
    font-size: 18px;
    color: #9aa0ac;
    margin: 0;
    line-height: 1.4;
  }

  .hero-og__footer {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 80px;
    background: rgba(0, 0, 0, 0.4);
    backdrop-filter: blur(10px);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 60px;
    font-size: 18px;
    color: #f2f3f5;
    letter-spacing: 0.08em;
    border-top: 2px solid #7a5c2e;
  }

  .hero-og__motto {
    font-family: 'Playfair Display', serif;
    font-style: italic;
    color: #7a5c2e;
  }
</style>