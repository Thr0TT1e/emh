<script setup lang="ts">
  import { type FlexibleDateJson } from '~/sdk/emh/v1/common_pb';
  import { type DatePrecisionJson } from '~/sdk/emh/v1/enums_emh_pb';

  const props = defineProps<{
    modelValue?: FlexibleDateJson;
    label: string;
  }>();

  const emit = defineEmits<{
    'update:modelValue': [value: FlexibleDateJson | undefined];
  }>();

  // ---------------------------------------------------------------------------
  // Хелпер для парсинга ISO даты в строку для UI инпута
  // ---------------------------------------------------------------------------
  const parseAnchorDateForUI = (iso: string | undefined, prec: DatePrecisionJson): string => {
    if (!iso) return '';
    // Для года нужен только YYYY
    if (prec === 'DATE_PRECISION_YEAR') {
      return iso.slice(0, 4);
    }
    // Для месяца нужен YYYY-MM
    if (prec === 'DATE_PRECISION_MONTH') {
      return iso.slice(0, 7);
    }
    // Для EXACT и SEASON нужен полный формат YYYY-MM-DD
    return iso.slice(0, 10);
  };

  // ---------------------------------------------------------------------------
  // Локальное состояние
  // ---------------------------------------------------------------------------
  const precision = ref<DatePrecisionJson>(props.modelValue?.precision ?? 'DATE_PRECISION_UNSPECIFIED');
  const displayText = ref(props.modelValue?.displayText ?? '');
  // Инициализируем строку даты через парсер
  const anchorDateStr = ref(parseAnchorDateForUI(props.modelValue?.anchorDate, precision.value));

  // Правила отображения полей согласно контракту
  const showAnchorDate = computed(() =>
    ['DATE_PRECISION_EXACT', 'DATE_PRECISION_MONTH', 'DATE_PRECISION_YEAR', 'DATE_PRECISION_SEASON'].includes(precision.value)
  );

  const requireDisplayText = computed(() =>
    ['DATE_PRECISION_SEASON', 'DATE_PRECISION_DAY_MONTH', 'DATE_PRECISION_RANGE'].includes(precision.value)
  );

  // Опции для селекта точности
  const precisionOptions = [
    { label: 'Не указано', value: 'DATE_PRECISION_UNSPECIFIED' as DatePrecisionJson },
    { label: 'Точная дата', value: 'DATE_PRECISION_EXACT' as DatePrecisionJson },
    { label: 'Месяц и год', value: 'DATE_PRECISION_MONTH' as DatePrecisionJson },
    { label: 'Только год', value: 'DATE_PRECISION_YEAR' as DatePrecisionJson },
    { label: 'Сезон и год', value: 'DATE_PRECISION_SEASON' as DatePrecisionJson },
    { label: 'День и месяц', value: 'DATE_PRECISION_DAY_MONTH' as DatePrecisionJson },
    { label: 'Диапазон дат', value: 'DATE_PRECISION_RANGE' as DatePrecisionJson },
    { label: 'Дата неизвестна', value: 'DATE_PRECISION_UNKNOWN' as DatePrecisionJson },
  ];

  // ---------------------------------------------------------------------------
  // Watchers
  // ---------------------------------------------------------------------------

  // При смене точности очищаем поле даты, так как форматы инпутов несовместимы
  watch(precision, (newVal, oldVal) => {
    if (newVal !== oldVal) {
      anchorDateStr.value = '';
    }
    if (newVal === 'DATE_PRECISION_UNSPECIFIED') {
      displayText.value = '';
    }
    emitValue();
  });

  watch([displayText, anchorDateStr], () => {
    emitValue();
  });

  // Синхронизация при изменении props извне (например, при загрузке данных)
  watch(() => props.modelValue, (nv) => {
    if (nv) {
      precision.value = nv.precision ?? 'DATE_PRECISION_UNSPECIFIED';
      displayText.value = nv.displayText ?? '';

      // Обновляем anchorDateStr ТОЛЬКО если anchorDate действительно есть в props
      // Это предотвращает сброс поля при неполном вводе года
      if (nv.anchorDate) {
        const newAnchorStr = parseAnchorDateForUI(nv.anchorDate, precision.value);
        if (newAnchorStr !== anchorDateStr.value) {
          anchorDateStr.value = newAnchorStr;
        }
      }
    } else {
      precision.value = 'DATE_PRECISION_UNSPECIFIED';
      displayText.value = '';
      anchorDateStr.value = '';
    }
  }, { deep: true });

  // ---------------------------------------------------------------------------
  // Логика отправки
  // ---------------------------------------------------------------------------
  function emitValue() {
    if (precision.value === 'DATE_PRECISION_UNSPECIFIED') {
      emit('update:modelValue', undefined);
      return;
    }

    const val: FlexibleDateJson = {
      precision: precision.value,
    };

    if (displayText.value.trim()) {
      val.displayText = displayText.value.trim();
    }

    if (showAnchorDate.value && anchorDateStr.value) {
      const raw = anchorDateStr.value;
      // Формируем корректную ISO-строку (RFC3339) только если ввод завершен
      if (precision.value === 'DATE_PRECISION_YEAR' && raw.length === 4) {
        val.anchorDate = `${raw}-01-01T00:00:00Z`;
      } else if (precision.value === 'DATE_PRECISION_MONTH' && raw.length === 7) {
        val.anchorDate = `${raw}-01T00:00:00Z`;
      } else if ((precision.value === 'DATE_PRECISION_EXACT' || precision.value === 'DATE_PRECISION_SEASON') && raw.length === 10) {
        val.anchorDate = `${raw}T00:00:00Z`;
      }
    }

    emit('update:modelValue', val);
  }
</script>

<template>
  <div class="fdi-container mb-4">
    <label class="afield__label mb-2 block">{{ label }}</label>

    <div class="fdi-grid flex flex-col gap-3">
      <!-- Выбор точности -->
      <Select v-model="precision" :options="precisionOptions" option-label="label" option-value="value" class="w-full"
        placeholder="Выберите точность даты" />

      <!-- Поле для anchor_date -->
      <div v-if="showAnchorDate" class="fdi-input-wrapper">
        <!-- Точная дата и Сезон -->
        <input v-if="precision === 'DATE_PRECISION_EXACT' || precision === 'DATE_PRECISION_SEASON'"
          v-model="anchorDateStr" type="date" class="afield__input w-full" placeholder="ДД.ММ.ГГГГ" />
        <!-- Месяц и год -->
        <input v-else-if="precision === 'DATE_PRECISION_MONTH'" v-model="anchorDateStr" type="month"
          class="afield__input w-full" placeholder="ММ.ГГГГ" />
        <!-- Только год -->
        <input v-else-if="precision === 'DATE_PRECISION_YEAR'" v-model="anchorDateStr" type="text" inputmode="numeric"
          pattern="\d{4}" maxlength="4" class="afield__input w-full" placeholder="ГГГГ" />
      </div>

      <!-- Поле для display_text -->
      <div
        v-if="requireDisplayText || displayText || precision === 'DATE_PRECISION_EXACT' || precision === 'DATE_PRECISION_MONTH' || precision === 'DATE_PRECISION_YEAR'">
        <input v-model="displayText" type="text" class="afield__input w-full"
          :class="{ 'border-red-500 focus:border-red-500': requireDisplayText && !displayText.trim() }"
          :placeholder="requireDisplayText ? 'Обязательно: например, «Лето 1989» или «28 июля»' : 'Комментарий к дате (необязательно)'" />
        <p v-if="requireDisplayText && !displayText.trim()" class="text-xs text-red-500 mt-1">
          Для выбранной точности текстовое описание обязательно
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .afield__input {
    width: 100%;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--emh-line);
    border-radius: 4px;
    background: var(--emh-surface);
    color: var(--emh-ink);
    font-family: var(--font-body);
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .afield__input:focus {
    outline: none;
    border-color: var(--emh-bronze);
    box-shadow: 0 0 0 2px rgba(176, 141, 87, 0.15);
  }

  .afield__input.border-red-500 {
    border-color: #ef4444;
  }

  .afield__input.border-red-500:focus {
    border-color: #ef4444;
    box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.15);
  }
</style>