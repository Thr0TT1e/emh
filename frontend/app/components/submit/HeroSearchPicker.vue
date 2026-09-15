<script setup lang="ts">
    import { useApi } from '~/composables/useApi'
    import { toPlain } from '~/lib/pb'
    import { HeroSummarySchema, type HeroSummaryJson } from '~/sdk/emh/v1/hero_pb'

    // Props
    const heroId = defineModel<string>('heroId', { default: '' })
    const heroName = defineModel<string>('heroName', { default: '' })

    // API
    const { hero } = useApi()

    // State
    const heroSearch = ref('')
    const heroResults = ref<HeroSummaryJson[]>([])
    const searching = ref(false)
    const showHeroPicker = ref(false)

    let searchTimer: ReturnType<typeof setTimeout> | undefined

    // Debounced search
    watch(heroSearch, () => {
        clearTimeout(searchTimer)

        if (!heroSearch.value.trim()) {
            heroResults.value = []
            showHeroPicker.value = false
            return
        }

        searchTimer = setTimeout(async () => {
            searching.value = true
            try {
                const res = await hero.listHeroes({
                    pagination: { pageSize: 10 },
                    searchQuery: heroSearch.value,
                })
                heroResults.value = (res.heroes ?? []).map((h) => toPlain(HeroSummarySchema, h))
                showHeroPicker.value = true
            } finally {
                searching.value = false
            }
        }, 400)
    })

    // Select hero
    const selectHero = (h: HeroSummaryJson) => {
        heroId.value = h.id ?? ''
        heroName.value = [h.lastName, h.firstName, h.middleName].filter(Boolean).join(' ')
        showHeroPicker.value = false
        heroSearch.value = ''
    }

    // Clear selection
    const clearHero = () => {
        heroId.value = ''
        heroName.value = ''
    }
</script>

<template>
    <div class="hero-search-picker">
        <div class="afield">
            <label class="afield__label">Дополнить существующую запись</label>
            <InputText v-model="heroSearch" placeholder="Введите ФИО для поиска..." class="w-full" />
            <ProgressSpinner v-if="searching" class="mt-2" style="width: 24px; height: 24px" />
        </div>

        <!-- Selected hero display -->
        <div v-if="heroId" class="selected-hero mt-4 p-3 bg-surface-50 dark:bg-surface-800 rounded">
            <div class="flex items-center justify-between">
                <div>
                    <i-mdi-account class="text-primary mr-2" />
                    <span class="font-medium">{{ heroName }}</span>
                </div>
                <Button icon="i-mdi-close" severity="secondary" text size="small" @click="clearHero" />
            </div>
        </div>

        <!-- Search results picker -->
        <div v-if="showHeroPicker && heroResults.length > 0" class="hero-results mt-3">
            <div class="text-sm text-muted mb-2">Найдено записей:</div>
            <div class="space-y-2">
                <div v-for="h in heroResults" :key="h.id"
                    class="hero-result-item p-3 border border-surface-200 dark:border-surface-700 rounded cursor-pointer hover:bg-surface-50 dark:hover:bg-surface-800"
                    @click="selectHero(h)">
                    <div class="font-medium">
                        {{ [h.lastName, h.firstName, h.middleName].filter(Boolean).join(' ') }}
                    </div>
                    <div v-if="h.rank" class="text-sm text-muted mt-1">{{ h.rank }}</div>
                    <div v-if="h.shortBio" class="text-sm text-muted mt-1 line-clamp-2">{{ h.shortBio }}</div>
                </div>
            </div>
        </div>

        <div v-else-if="showHeroPicker && heroResults.length === 0 && !searching" class="mt-3 text-muted text-sm">
            Ничего не найдено
        </div>
    </div>
</template>

<style scoped>
    .hero-result-item {
        transition: background-color 0.15s ease;
    }
</style>