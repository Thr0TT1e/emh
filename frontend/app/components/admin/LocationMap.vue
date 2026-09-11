<script setup lang="ts">
    /**
     * LocationMap — выбор координат локации на карте (OpenLayers + тайлы OSM).
     * Клик по карте / перетаскивание маркера → emit('pick', lat, lon).
     * Поиск места — Nominatim (© OpenStreetMap contributors): покрытие всего мира,
     * что важно для зарубежных конфликтов (Сирия, Вьетнам, Куба, ЮАР...).
     */
    import Collection from 'ol/Collection'
    import Feature from 'ol/Feature'
    import Map from 'ol/Map'
    import View from 'ol/View'
    import Point from 'ol/geom/Point'
    import Translate from 'ol/interaction/Translate'
    import TileLayer from 'ol/layer/Tile'
    import VectorLayer from 'ol/layer/Vector'
    import 'ol/ol.css'
    import { fromLonLat, toLonLat } from 'ol/proj'
    import OSM from 'ol/source/OSM'
    import VectorSource from 'ol/source/Vector'
    import CircleStyle from 'ol/style/Circle'
    import Fill from 'ol/style/Fill'
    import Stroke from 'ol/style/Stroke'
    import Style from 'ol/style/Style'

    const props = withDefaults(defineProps<{ lat?: number; lon?: number }>(), { lat: 0, lon: 0 })
    const emit = defineEmits<{ pick: [lat: number, lon: number] }>()

    const mapEl = ref<HTMLDivElement>()
    const query = ref('')
    const searching = ref(false)

    let map: Map | null = null
    let marker: Feature<Point> | null = null
    let ro: ResizeObserver | null = null

    const hasCoords = computed(() => props.lat !== 0 || props.lon !== 0)

    // Стартовый вид — центр РФ; поиск улетает в любую точку мира.
    const FALLBACK_CENTER: [number, number] = [45, 61] // [lon, lat]
    const FALLBACK_ZOOM = 4

    const round = (v: number) => Math.round(v * 1e6) / 1e6

    function setMarker(lonLat: [number, number]) {
        marker?.getGeometry()?.setCoordinates(fromLonLat(lonLat))
    }

    function flyTo(lonLat: [number, number], zoom = 12) {
        map?.getView().animate({ center: fromLonLat(lonLat), zoom, duration: 600 })
    }

    onMounted(() => {
        if (!mapEl.value) return
        const center: [number, number] = hasCoords.value ? [props.lon, props.lat] : FALLBACK_CENTER

        marker = new Feature(new Point(fromLonLat(center)))
        marker.setStyle(new Style({
            image: new CircleStyle({
                radius: 8,
                fill: new Fill({ color: '#bd2f3c' }), // carmine из палитры EMH
                stroke: new Stroke({ color: '#fff', width: 2 }),
            }),
        }))

        map = new Map({
            target: mapEl.value,
            layers: [
                new TileLayer({ source: new OSM() }),
                new VectorLayer({ source: new VectorSource({ features: [marker] }) }),
            ],
            view: new View({ center: fromLonLat(center), zoom: hasCoords.value ? 12 : FALLBACK_ZOOM }),
        })

        // Клик по карте — задаём координаты
        map.on('click', (e) => {
            const [lon, lat] = toLonLat(e.coordinate) as [number, number]
            setMarker([lon, lat])
            emit('pick', round(lat), round(lon))
        })

        // Перетаскивание маркера
        const translate = new Translate({ features: new Collection([marker]) })
        translate.on('translateend', (e) => {
            const geom = (e.features.item(0) as Feature<Point>).getGeometry()!
            const [lon, lat] = toLonLat(geom.getCoordinates()) as [number, number]
            emit('pick', round(lat), round(lon))
        })
        map.addInteraction(translate)

        // Диалог PrimeVue раскрывается после монтирования — следим за размером
        ro = new ResizeObserver(() => map?.updateSize())
        ro.observe(mapEl.value)
    })

    onBeforeUnmount(() => {
        ro?.disconnect()
        map?.dispose()
        map = null
    })

    // Ручной ввод координат в форме — двигаем маркер (без emit обратно)
    watch([() => props.lat, () => props.lon], ([la, lo]) => {
        if (la >= -90 && la <= 90 && lo >= -180 && lo <= 180 && (la || lo)) setMarker([lo, la])
    })

    async function geocode() {
        const q = query.value.trim()
        if (!q || searching.value) return
        searching.value = true
        try {
            const url = `https://nominatim.openstreetmap.org/search?format=jsonv2&limit=1&accept-language=ru&q=${encodeURIComponent(q)}`
            const rows = await (await fetch(url, { headers: { Accept: 'application/json' } })).json()
            if (rows?.[0]) {
                const lon = Number(rows[0].lon)
                const lat = Number(rows[0].lat)
                setMarker([lon, lat])
                flyTo([lon, lat])
                emit('pick', round(lat), round(lon))
            } else {
                query.value = q // оставить как было
            }
        } catch { /* поиск недоступен — молча игнорируем */ }
        finally { searching.value = false }
    }
</script>

<template>
    <div class="locmap">
        <div class="locmap__search">
            <InputText v-model="query" class="w-full" placeholder="Поиск: страна, город, мемориал (например, Алеппо)…"
                @keyup.enter.prevent="geocode" />
            <Button type="button" label="Найти" :loading="searching" @click="geocode" />
        </div>

        <div ref="mapEl" class="locmap__canvas" />

        <div class="locmap__hint">
            Клик по карте или перетаскивание маркера задаёт координаты
        </div>
    </div>
</template>

<style scoped>
    .locmap__search {
        display: flex;
        gap: .6rem;
        margin-bottom: .6rem;
    }

    .locmap__canvas {
        height: 340px;
        border-radius: 8px;
        overflow: hidden;
        border: 1px solid var(--p-surface-200);
    }

    .locmap__hint {
        font-size: .75rem;
        color: var(--p-text-muted-color);
        margin-top: .4rem;
    }
</style>