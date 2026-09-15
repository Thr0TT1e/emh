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
    import OSM from 'ol/source/OSM'
    import VectorSource from 'ol/source/Vector'
    import { Fill, Icon, Stroke, Style, Text } from 'ol/style'
    import type { HeroLocationJson } from '~/sdk/emh/v1/hero_pb'

    const props = defineProps<{
        locations: HeroLocationJson[]
    }>()

    const LOCATION_TYPE_CONFIG: Record<string, { label: string; color: string }> = {
        HERO_LOCATION_TYPE_BIRTH: { label: 'Место рождения', color: '#4a90d9' },
        HERO_LOCATION_TYPE_DEATH: { label: 'Место гибели', color: '#c0392b' },
        HERO_LOCATION_TYPE_BURIAL: { label: 'Место захоронения', color: '#7a5c2e' },
        HERO_LOCATION_TYPE_RESIDENCE: { label: 'Место проживания', color: '#27ae60' },
    }

    // ── Хелпер: безопасное приведение double из protobuf-JSON к number ──
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

    // ── Генерация SVG-пина с нужным цветом ──
    const getSvgIconDataUrl = (color: string): string => {
        const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
            <path d="M0 0h32v32H0z" fill="none" />
            <path fill="${color}" d="M16 18a5 5 0 1 1 5-5a5.006 5.006 0 0 1-5 5m0-8a3 3 0 1 0 3 3a3.003 3.003 0 0 0-3-3" />
            <path fill="${color}" d="m16 30l-8.436-9.949a35 35 0 0 1-.348-.451A10.9 10.9 0 0 1 5 13a11 11 0 0 1 22 0a10.9 10.9 0 0 1-2.215 6.597l-.001.003s-.3.394-.345.447ZM8.813 18.395s.233.308.286.374L16 26.908l6.91-8.15c.044-.055.278-.365.279-.366A8.9 8.9 0 0 0 25 13a9 9 0 1 0-18 0a8.9 8.9 0 0 0 1.813 5.395" />
        </svg>`
        // encodeURIComponent безопасен как в браузере, так и при SSR
        return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`
    }

    // Кэш иконок, чтобы не пересоздавать base64 и DOM-объекты при каждом вызове style-функции
    const iconCache = new Map<string, Icon>()

    function getIcon(color: string): Icon {
        if (!iconCache.has(color)) {
            iconCache.set(
                color,
                new Icon({
                    src: getSvgIconDataUrl(color),
                    scale: 1.2, // Увеличиваем пин для лучшей видимости (38x38px)
                    // Острие пина находится на y=30 из 32 (30/32 = 0.9375). 
                    // Привязываем маркер к карте именно за острие.
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
            // Пропускаем локации без координат
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

    // ── Стиль маркера и подписи ──
    function createMarkerStyle(feature: FeatureLike): Style {
        const color = feature.get('typeColor') ?? '#7a5c2e'
        const name = feature.get('name') ?? ''

        return new Style({
            image: getIcon(color),
            text: new Text({
                text: name,
                offsetY: -42, // Сдвигаем текст выше пина (высота пина ~38px)
                textBaseline: 'bottom', // Выравниваем низ текста по указанной Y-координате
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

        const tileLayer = new TileLayer({ source: new OSM(), opacity: 0.85 })
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

        // Подгоняем вид ко всем маркерам
        if (features.length > 1) {
            const extent = vectorSource.getExtent()
            if (extent) {
                map.getView().fit(extent, { padding: [50, 50, 50, 50], maxZoom: 12 })
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

    onMounted(() => {
        initMap()
        setupTooltip()
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
        <h3 class="hero-location-map__title">География</h3>

        <div ref="mapContainer" class="hero-location-map__container">
            <div class="hero-map-tooltip" style="display: none" />
        </div>

        <ul class="hero-location-map__legend">
            <li v-for="(config, key) in LOCATION_TYPE_CONFIG" :key="key"
                v-show="locations.some((hl) => hl.type === key)" class="hero-location-map__legend-item">
                <!-- Используем тот же SVG для консистентности с маркерами на карте -->
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

    @media (max-width: 768px) {
        .hero-location-map__container {
            height: 240px;
        }
    }
</style>