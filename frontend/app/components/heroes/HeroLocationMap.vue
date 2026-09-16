<script setup lang="ts">
    import type { FeatureLike } from 'ol/Feature'
    import Feature from 'ol/Feature'
    import OLMap from 'ol/Map'
    import View from 'ol/View'
    import Point from 'ol/geom/Point'
    import TileLayer from 'ol/layer/Tile'
    import VectorLayer from 'ol/layer/Vector'
    import 'ol/ol.css'
    import { fromLonLat } from 'ol/proj'
    import VectorSource from 'ol/source/Vector'
    import XYZ from 'ol/source/XYZ'
    import { Fill, Icon, Stroke, Style, Text } from 'ol/style'
    import type { HeroLocationJson } from '~/sdk/emh/v1/hero_pb'

    const ESRI_TILE_URL = 'https://services.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}'
    const ESRI_PROBE_URL = 'https://services.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/1/0/0'
    const PROBE_TIMEOUT_MS = 4000

    const props = defineProps<{ locations: HeroLocationJson[] }>()

    const LOCATION_TYPE_CONFIG: Record<string, { label: string; color: string }> = {
        HERO_LOCATION_TYPE_BIRTH: { label: 'Место рождения', color: '#4a90d9' },
        HERO_LOCATION_TYPE_DEATH: { label: 'Место гибели', color: '#c0392b' },
        HERO_LOCATION_TYPE_BURIAL: { label: 'Место захоронения', color: '#7a5c2e' },
        HERO_LOCATION_TYPE_RESIDENCE: { label: 'Место проживания', color: '#27ae60' },
    }

    const toFiniteNumber = (v: number | string | undefined | null): number | null => {
        if (typeof v === 'number') return Number.isFinite(v) ? v : null
        if (typeof v === 'string') {
            const n = Number(v)
            return Number.isFinite(n) ? n : null
        }
        return null
    }

    const mapContainer = ref<HTMLElement>()
    let map: OLMap | null = null

    // ── Состояние доступности карты ──
    // null = идёт проверка, true = доступна, false = недоступна
    const mapAvailable = ref<boolean | null>(null)

    // ── Pre-flight проверка доступности сервиса тайлов ──
    async function probeTileService(): Promise<boolean> {
        // Safari fallback: AbortSignal.timeout может отсутствовать в старых версиях
        const createAbortSignal = (ms: number): AbortSignal => {
            if (typeof AbortSignal !== 'undefined' && typeof AbortSignal.timeout === 'function') {
                return AbortSignal.timeout(ms)
            }
            const controller = new AbortController()
            setTimeout(() => controller.abort(), ms)
            return controller.signal
        }

        try {
            const response = await fetch(ESRI_PROBE_URL, {
                method: 'HEAD',
                mode: 'cors',
                signal: createAbortSignal(PROBE_TIMEOUT_MS),
            })
            return response.ok
        } catch {
            return false
        }
    }

    // ── SVG-пин с динамическим цветом ──
    const getSvgIconDataUrl = (color: string): string => {
        const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
    <path d="M0 0h32v32H0z" fill="none" />
    <path fill="${color}" d="M16 18a5 5 0 1 1 5-5a5.006 5.006 0 0 1-5 5m0-8a3 3 0 1 0 3 3a3.003 3.003 0 0 0-3-3" />
    <path fill="${color}" d="m16 30l-8.436-9.949a35 35 0 0 1-.348-.451A10.9 10.9 0 0 1 5 13a11 11 0 0 1 22 0a10.9 10.9 0 0 1-2.215 6.597l-.001.003s-.3.394-.345.447ZM8.813 18.395s.233.308.286.374L16 26.908l6.91-8.15c.044-.055.278-.365.279-.366A8.9 8.9 0 0 0 25 13a9 9 0 1 0-18 0a8.9 8.9 0 0 0 1.813 5.395" />
  </svg>`
        return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`
    }

    const iconCache = new Map<string, Icon>()

    function getIcon(color: string): Icon {
        if (!iconCache.has(color)) {
            iconCache.set(
                color,
                new Icon({
                    src: getSvgIconDataUrl(color),
                    scale: 1.2,
                    anchor: [0.5, 0.9375],
                    anchorXUnits: 'fraction',
                    anchorYUnits: 'fraction',
                })
            )
        }
        return iconCache.get(color)!
    }

    function createFeatures(locations: HeroLocationJson[]): Feature<Point>[] {
        return locations.flatMap((hl) => {
            const loc = hl.location
            if (!loc) return []

            const lat = toFiniteNumber(loc.latitude)
            const lon = toFiniteNumber(loc.longitude)
            if (lat === null || lon === null || (lat === 0 && lon === 0)) return []

            const typeStr = (hl.type as string) ?? 'HERO_LOCATION_TYPE_UNSPECIFIED'
            const config = LOCATION_TYPE_CONFIG[typeStr] ?? { label: 'Локация', color: '#95a5a6' }

            return [
                new Feature({
                    geometry: new Point(fromLonLat([lon, lat])),
                    name: loc.name || 'Без названия',
                    historicalName: loc.historicalName || '',
                    typeLabel: config.label,
                    typeColor: config.color,
                }),
            ]
        })
    }

    function createMarkerStyle(feature: FeatureLike): Style {
        const color = feature.get('typeColor') ?? '#7a5c2e'
        const name = feature.get('name') ?? ''

        return new Style({
            image: getIcon(color),
            text: new Text({
                text: name,
                offsetY: -42,
                textBaseline: 'bottom',
                font: '12px "Golos Text", sans-serif',
                fill: new Fill({ color: '#ffffff' }),
                stroke: new Stroke({ color: '#000000', width: 3 }),
                textAlign: 'center',
            }),
        })
    }

    function initMap() {
        if (!mapContainer.value || !props.locations.length) return

        const features = createFeatures(props.locations)
        if (!features.length) return

        const vectorSource = new VectorSource({ features })

        const tileLayer = new TileLayer({
            source: new XYZ({
                url: ESRI_TILE_URL,
                attributions: 'Tiles © Esri',
                crossOrigin: 'anonymous',
            }),
        })

        // Fallback: если тайлы начали падать уже во время работы карты — скрываем её
        tileLayer.getSource()?.on('tileloaderror', () => {
            console.warn('[HeroLocationMap] tile load failed, hiding map')
            mapAvailable.value = false
        })

        const vectorLayer = new VectorLayer({ source: vectorSource, style: createMarkerStyle })

        const firstFeature = features[0]
        const center = firstFeature
            ? firstFeature.getGeometry()!.getCoordinates()
            : fromLonLat([45, 60])

        map = new OLMap({
            target: mapContainer.value,
            layers: [tileLayer, vectorLayer],
            view: new View({
                center,
                zoom: features.length > 1 ? 5 : 10,
                maxZoom: 18,
            }),
        })

        if (features.length > 1) {
            const extent = vectorSource.getExtent()
            if (extent) {
                map.getView().fit(extent, { padding: [50, 50, 50, 50], maxZoom: 8 })
            }
        }
    }

    function setupTooltip() {
        if (!map) return

        map.on('pointermove', (evt) => {
            const feature = map!.forEachFeatureAtPixel(evt.pixel, (f) => f)
            const container = mapContainer.value
            if (!container) return

            const tooltip = container.querySelector('.hero-map-tooltip') as HTMLElement
            if (!tooltip) return

            if (feature) {
                const name = feature.get('name') ?? ''
                const typeLabel = feature.get('typeLabel') ?? ''
                const historicalName = feature.get('historicalName') ?? ''

                let html = `<strong>${name}</strong>`
                if (historicalName) html += `<br><em>${historicalName}</em>`
                html += `<br>${typeLabel}`

                tooltip.innerHTML = html
                tooltip.style.display = 'block'

                const original = evt.originalEvent as PointerEvent
                tooltip.style.left = `${original.offsetX + 12}px`
                tooltip.style.top = `${original.offsetY - 12}px`
            } else {
                tooltip.style.display = 'none'
            }
        })
    }

    onMounted(async () => {
        if (!props.locations.length) return

        // Pre-flight: проверяем доступность сервиса тайлов
        mapAvailable.value = await probeTileService()

        if (mapAvailable.value) {
            // nextTick чтобы контейнер успел отрендериться
            await nextTick()
            initMap()
            setupTooltip()
        }
    })

    onUnmounted(() => {
        if (map) {
            map.setTarget(undefined)
            map = null
        }
    })
</script>

<template>
    <div v-if="locations.length > 0" class="hero-location-map">
        <!-- Skeleton во время pre-flight проверки -->
        <div v-if="mapAvailable === null" class="hero-location-map__skeleton">
            <h3 class="hero-location-map__title">География</h3>
            <div class="hero-location-map__skeleton-box" />
        </div>

        <!-- Карта доступна — рендерим -->
        <template v-else-if="mapAvailable">
            <h3 class="hero-location-map__title">География</h3>
            <div ref="mapContainer" class="hero-location-map__container">
                <div class="hero-map-tooltip" style="display: none" />
            </div>

            <ul class="hero-location-map__legend">
                <li v-for="(config, key) in LOCATION_TYPE_CONFIG" :key="key"
                    v-show="locations.some((hl) => hl.type === key)" class="hero-location-map__legend-item">
                    <svg class="hero-location-map__legend-icon" viewBox="0 0 32 32" :style="{ color: config.color }">
                        <path d="M0 0h32v32H0z" fill="none" />
                        <path fill="currentColor"
                            d="M16 18a5 5 0 1 1 5-5a5.006 5.006 0 0 1-5 5m0-8a3 3 0 1 0 3 3a3.003 3.003 0 0 0-3-3" />
                        <path fill="currentColor"
                            d="m16 30l-8.436-9.949a35 35 0 0 1-.348-.451A10.9 10.9 0 0 1 5 13a11 11 0 0 1 22 0a10.9 10.9 0 0 1-2.215 6.597l-.001.003s-.3.394-.345.447ZM8.813 18.395s.233.308.286.374L16 26.908l6.91-8.15c.044-.055.278-.365.279-.366A8.9 8.9 0 0 0 25 13a9 9 0 1 0-18 0a8.9 8.9 0 0 0 1.813 5.395" />
                    </svg>
                    {{ config.label }}
                </li>
            </ul>
        </template>
        <!-- mapAvailable === false: блок не рендерится вообще -->
    </div>
</template>

<style scoped>
    .hero-location-map {
        margin-top: 2rem;
    }

    .hero-location-map__title {
        font-family: 'Playfair Display', serif;
        font-size: 1.25rem;
        font-weight: 600;
        margin-bottom: 1rem;
        color: var(--emh-text);
    }

    .hero-location-map__container {
        position: relative;
        width: 100%;
        height: 320px;
        border-radius: 8px;
        overflow: hidden;
        border: 1px solid var(--emh-border, #e0e0e0);
        isolation: isolate;
    }

    .hero-map-tooltip {
        position: absolute;
        z-index: 10;
        background: rgba(0, 0, 0, 0.85);
        color: #fff;
        padding: 6px 10px;
        border-radius: 4px;
        font-size: 0.8rem;
        line-height: 1.4;
        pointer-events: none;
        max-width: 220px;
    }

    .hero-location-map__legend {
        display: flex;
        flex-wrap: wrap;
        gap: 0.75rem 1.5rem;
        margin-top: 0.75rem;
        padding: 0;
        list-style: none;
    }

    .hero-location-map__legend-item {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        font-size: 0.8rem;
        color: var(--emh-muted, #666);
    }

    .hero-location-map__legend-icon {
        width: 16px;
        height: 16px;
        flex-shrink: 0;
    }

    /* Skeleton */
    .hero-location-map__skeleton-box {
        width: 100%;
        height: 320px;
        border-radius: 8px;
        background: linear-gradient(90deg,
                var(--emh-surface, #f5f5f5) 0%,
                var(--emh-surface-alt, #e8e8e8) 50%,
                var(--emh-surface, #f5f5f5) 100%);
        background-size: 200% 100%;
        animation: skeleton-pulse 1.5s ease-in-out infinite;
    }

    @keyframes skeleton-pulse {
        0% {
            background-position: 200% 0;
        }

        100% {
            background-position: -200% 0;
        }
    }

    @media (max-width: 768px) {

        .hero-location-map__container,
        .hero-location-map__skeleton-box {
            height: 240px;
        }
    }
</style>