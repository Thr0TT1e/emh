<script setup lang="ts">
  import RibbonIcon from '~icons/mdi/ribbon?width=1.25em&height=1.25em';
  import type { HeroAwardJson } from '~/sdk/emh/v1/hero_pb';

  // Награды приходят уже отфильтрованными (только с лентой) и отсортированными
  // по старшинству — см. ribbonAwardsOf() из ~/lib/awards.
  defineProps<{ awards: HeroAwardJson[] }>();

  // Минимальный срез API Popover: PrimeVue-компонент автоимпортируется,
  // ссылаться на его тип в type-позиции нельзя.
  interface PopoverInstance {
    toggle: (event: Event) => void;
  }

  const popover = ref<PopoverInstance | null>(null);
  const expanded = ref(false);

  const toggle = (event: MouseEvent) => popover.value?.toggle(event);
</script>

<template>
  <div class="arb">
    <ul class="arb__bar">
      <li v-for="a in awards" :key="a.awardId" class="arb__cell">
        <img
          class="arb__ribbon"
          :src="a.ribbonImageUrl"
          :alt="a.awardName"
          width="36"
          height="12"
          loading="lazy"
          decoding="async"
        />
      </li>
    </ul>

    <button
      type="button"
      class="arb__toggle"
      :aria-expanded="expanded"
      aria-haspopup="dialog"
      @click="toggle"
    >
      <RibbonIcon aria-hidden="true" />
      <span>Все награды ({{ awards.length }})</span>
    </button>

    <Popover ref="popover" class="arb__popover" @show="expanded = true" @hide="expanded = false">
      <ul class="arb__list">
        <li v-for="a in awards" :key="a.awardId" class="arb__item">
          <img
            v-if="a.imageUrl"
            class="arb__sign"
            :src="a.imageUrl"
            :alt="a.awardName"
            width="144"
            height="48"
            loading="lazy"
            decoding="async"
          />
          <span class="arb__name">{{ a.awardName }}</span>
          <span v-if="a.decreeNumber" class="arb__decree">{{ a.decreeNumber }}</span>
        </li>
      </ul>
    </Popover>
  </div>
</template>

<style scoped>
  .arb {
    margin: 0 0 1.8rem;
  }

  /* Планки идут сплошным блоком: ленты смыкаются краями, как на колодке. */
  .arb__bar {
    list-style: none;
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
    margin: 0 0 0.7rem;
    padding: 4px;
    background: var(--emh-surface);
    border: 1px solid var(--emh-line);
  }

  .arb__cell {
    display: flex;
  }

  .arb__ribbon {
    display: block;
    width: 36px;
    height: 12px;
    object-fit: cover;
  }

  .arb__toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.25rem 0;
    border: 0;
    background: none;
    color: var(--emh-crimson);
    font: inherit;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
  }

  .arb__toggle:hover {
    color: var(--emh-crimson-dark);
  }

  .arb__list {
    list-style: none;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(9rem, 1fr));
    gap: 0.9rem;
    max-width: 32rem;
    max-height: 60vh;
    margin: 0;
    padding: 0;
    overflow-y: auto;
  }

  .arb__item {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }

  .arb__sign {
    width: 100%;
    max-width: 144px;
    height: auto;
    object-fit: contain;
  }

  .arb__name {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--emh-ink);
    line-height: 1.3;
  }

  .arb__decree {
    font-size: 0.78rem;
    color: var(--emh-muted);
  }
</style>
