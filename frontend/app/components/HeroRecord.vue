<script setup lang="ts">
  import { formatFlexibleYear } from "~/lib/format";
  import type { HeroSummaryJson } from "~/sdk/emh/v1/hero_pb";

  defineProps<{ hero: HeroSummaryJson }>();
</script>
<template>
  <NuxtLink :to="`/heroes/${hero.id}`" class="record">
    <!-- ... фото ... -->
    <div class="record__body">
      <h3 class="record__name">
        {{ hero.lastName }} {{ hero.firstName }} {{ hero.middleName }}
        <span v-if="hero.nickname" class="record__nickname">«{{ hero.nickname }}»</span>
      </h3>

      <p class="record__dates">
        {{ formatFlexibleYear(hero.birthDateInfo, hero.birthDate) }} —
        {{ formatFlexibleYear(hero.deathDateInfo, hero.deathDate) }}
      </p>

      <p v-if="hero.rank || hero.unit" class="record__rank">
        {{ [hero.rank, hero.unit].filter(Boolean).join(" · ") }}
      </p>

      <p v-if="hero.shortBio" class="record__bio">{{ hero.shortBio }}</p>

      <ul v-if="hero?.awardNames?.length" class="record__awards">
        <li v-for="a in hero.awardNames" :key="a" class="record__award">{{ a }}</li>
      </ul>
    </div>

    <span class="record__arrow" aria-hidden="true">→</span>
  </NuxtLink>
</template>

<style scoped>
  .record {
    display: grid;
    grid-template-columns: 96px 1fr auto;
    gap: 1.4rem;
    align-items: center;
    padding: 1.15rem 1.25rem;
    border-bottom: 1px solid var(--emh-line);
    color: var(--emh-ink);
    transition: background-color 0.25s ease, transform 0.25s ease, box-shadow 0.25s ease;
  }

  .record:hover {
    background: #ffffff;
    transform: translateX(6px);
    box-shadow: 0 10px 28px -18px rgba(26, 28, 32, 0.35);
  }

  .record__photo {
    width: 96px;
    height: 118px;
    border: 1px solid var(--emh-bronze);
    padding: 3px;
    background: #fff;
    overflow: hidden;
  }

  .record__photo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    filter: grayscale(0.85) contrast(1.02);
    transition: filter 0.45s ease, transform 0.45s ease;
  }

  .record:hover .record__photo img {
    filter: grayscale(0) contrast(1);
    transform: scale(1.04);
  }

  .record__photo-empty {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    font-size: 0.68rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--emh-muted);
    background: var(--emh-bg);
  }

  .record__name {
    font-family: var(--font-display);
    font-size: 1.45rem;
    font-weight: 700;
    line-height: 1.2;
    margin: 0 0 0.15rem;
  }

  .record__dates {
    margin: 0 0 0.3rem;
    font-size: 0.95rem;
    letter-spacing: 0.06em;
    color: var(--emh-bronze);
    font-weight: 600;
  }

  .record__rank {
    margin: 0 0 0.35rem;
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--emh-muted);
  }

  .record__bio {
    margin: 0;
    font-size: 0.92rem;
    color: var(--emh-muted);
    line-height: 1.5;
  }

  .record__awards {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    margin: 0.55rem 0 0;
    padding: 0;
    list-style: none;
  }

  .record__award {
    font-size: 0.72rem;
    letter-spacing: 0.05em;
    padding: 0.22rem 0.55rem;
    border: 1px solid var(--emh-bronze);
    color: #7a5c2e;
    background: rgba(176, 141, 87, 0.08);
  }

  .record__arrow {
    font-size: 1.4rem;
    color: var(--emh-bronze);
    opacity: 0;
    transform: translateX(-8px);
    transition: opacity 0.25s ease, transform 0.25s ease;
  }

  .record:hover .record__arrow {
    opacity: 1;
    transform: none;
  }

  .record__nickname {
    font-family: var(--font-display);
    font-style: italic;
    font-weight: 500;
    font-size: .85em;
    color: var(--emh-crimson);
  }

  @media (max-width: 640px) {
    .record {
      grid-template-columns: 72px 1fr;
    }

    .record__photo {
      width: 72px;
      height: 90px;
    }

    .record__arrow {
      display: none;
    }
  }
</style>
