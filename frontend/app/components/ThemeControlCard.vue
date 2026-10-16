<script setup lang="ts">
import type { ThemeMode } from "~/types/theme";

const mode = storeToRefs(useThemeStore()).mode;
const show = defineModel<boolean>("show", { required: true, default: false });

const primaryColorList = [
  "#99a2ff", // periwinkle
  "#7a8b74", // sage
  "#8b7e6b", // warm taupe
  "#7a7a8b", // muted lavender
  "#8b7474", // dusty rose
  "#6b8b8b", // teal
];

const isDark = computed(() => {
  return mode.value === "dark";
});

const setThemeMode = (mode: ThemeMode) => {
  useThemeStore().setMode(mode);
};

const setPrimaryColor = (primaryColor: string) => {
  useThemeStore().setPrimaryColor(primaryColor);
};

const colorInput = useTemplateRef<HTMLInputElement>("colorInput");
const colorInputValue = ref("");

const handleColorChange = () => {
  setPrimaryColor(colorInputValue.value);
};

const { $getLocales, $getLocale, $switchLocale, $ts } = useI18n();

const options = computed(() => {
  return $getLocales().map((l) => {
    return { label: l.name as string, value: l.code as string };
  });
});

const localeRef = ref($getLocale());

onMounted(() => {
  watch(localeRef, () => {
    $switchLocale(localeRef.value);
  });
});
</script>

<template>
  <Transition name="test">
    <div v-if="show" ref="themeControlCard" class="theme-control-card">
      <p class="title">{{ $ts("themeControl.themeMode") }}</p>
      <div :class="{ 'buttons-track': true, dark: isDark }">
        <button :class="{ 'buttons-item': true, active: !isDark }" @click="setThemeMode('light')">
          <Icon name="streamline-plump-color:sun" :size="20" />
          {{ $ts("themeControl.light") }}
        </button>
        <button :class="{ 'buttons-item': true, active: isDark }" @click="setThemeMode('dark')">
          <Icon name="streamline-plump-color:moon-stars" :size="20" />
          {{ $ts("themeControl.dark") }}
        </button>
      </div>
      <p class="title">{{ $ts("themeControl.primaryColor") }}</p>
      <div class="color-picker">
        <button
          v-for="item in primaryColorList"
          :key="item"
          class="color-picker-item"
          :style="{ background: item }"
          @click="setPrimaryColor(item)"
        />
        <button class="color-picker-item" @click="colorInput?.click()">+</button>
        <input
          v-show="false"
          ref="colorInput"
          v-model="colorInputValue"
          type="color"
          @change="handleColorChange"
        />
      </div>
      <p class="title">{{ $ts("themeControl.language") }}</p>
      <InputSelect v-model:value="localeRef" :options />
    </div>
  </Transition>
</template>

<style lang="less" scoped>
.theme-control-card {
  width: 100%;
  max-width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.test-enter-active,
.test-leave-active {
  transition: all 0.3s ease-in-out;
  transform-origin: right bottom;
}

.test-enter-from,
.test-leave-to {
  transform: scale(0.3);
  opacity: 0;
}

.title {
  font-size: 12px;
  font-family: "Maple Mono", monospace;
  font-weight: 500;
  color: var(--text-secondary);
}

.buttons-track {
  width: calc(100% - 10px);
  height: 35px;
  background: var(--bg-interactive-hover);
  margin: 0 auto;
  border-radius: 8px;
  position: relative;
  display: flex;
  border: 1px solid var(--border-subtle);

  &::after {
    content: "";
    position: absolute;
    left: 2px;
    top: 2px;
    width: calc(50% - 3px);
    height: calc(100% - 4px);
    background: var(--brand-primary);
    z-index: 0;
    border-radius: 6px;
    transition: transform 0.3s ease-in-out;
    box-shadow: var(--shadow-sm);
  }

  &.dark::after {
    transform: translateX(calc(100% - 2px));
  }
}

.buttons-item {
  width: 50%;
  background-color: transparent;
  border: none;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  color: var(--text-tertiary);
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 5px;
  position: relative;
  z-index: 1;
  transition: color 0.3s ease;

  &.active {
    color: var(--text-on-brand);
  }
}

.color-picker {
  display: grid;
  height: 32px;
  width: calc(100% - 10px);
  margin: 0 auto;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  justify-content: center;
  gap: 6px;
}

.color-picker-item {
  min-width: 0;
  cursor: pointer;
  border: none;
  box-shadow: var(--shadow-md);
  border-radius: 5px;
  background-color: transparent;
}
</style>
