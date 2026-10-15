<script setup lang="ts">
import type { CSSProperties } from "vue";
import type { ToolbarConfig, ToolbarItem, ToolbarPanel } from "./types";

interface Props {
  items: ToolbarItem[];
  config: Required<ToolbarConfig>;
  label?: string;
}

interface Emits {
  (event: "action", item: Extract<ToolbarItem, { type: "action" }>): void;
}

const { items, config, label = "工具栏" } = defineProps<Props>();
const emit = defineEmits<Emits>();
const activePanel = shallowRef<ToolbarPanel | null>(null);
const overflowOpen = shallowRef(false);
const visibleIds = shallowRef<string[]>([]);
const overflowIds = shallowRef<string[]>([]);
const shell = shallowRef<HTMLElement | null>(null);
const measureRow = shallowRef<HTMLElement | null>(null);
let resizeObserver: ResizeObserver | undefined;
let resizeFrame = 0;

const maxWidth = computed(() => Math.max(240, config.maxWidth ?? 480));
const gap = computed(() => Math.max(0, config.gap ?? 6));
const isExpanded = computed(() => Boolean(activePanel.value || overflowOpen.value));

const barWidth = computed(() => {
  const count = visibleItems.value.length + (overflowItems.value.length ? 1 : 0);
  return 16 + count * 40 + Math.max(0, count - 1) * gap.value;
});

const expandedWidth = computed(() => {
  const requested = activePanel.value?.width ?? 0;
  return Math.min(Math.max(240, barWidth.value, requested), Math.max(240, viewportWidth() - 32));
});

const shellStyle = computed<CSSProperties>(() => ({
  "--toolbar-gap": `${gap.value}px`,
  "--toolbar-expanded-width": `${expandedWidth.value}px`,
  "--toolbar-max-width": `${maxWidth.value}px`,
  width: isExpanded.value ? "var(--toolbar-expanded-width)" : "max-content",
}));

const activeComponent = computed(() => activePanel.value?.component);
const visibleItems = computed(() => items.filter((item) => visibleIds.value.includes(item.id)));
const overflowItems = computed(() => items.filter((item) => overflowIds.value.includes(item.id)));

function viewportWidth() {
  if (typeof document === "undefined") return maxWidth.value;
  return document.documentElement.clientWidth;
}

function scheduleLayout() {
  if (typeof window === "undefined") return;
  cancelAnimationFrame(resizeFrame);
  resizeFrame = requestAnimationFrame(() => void calculateLayout());
}

async function calculateLayout() {
  await nextTick();
  if (!measureRow.value) return;

  const availableWidth = Math.min(maxWidth.value, Math.max(0, viewportWidth() - 32));
  const itemGap = gap.value;
  const capacity = Math.max(1, Math.floor((availableWidth - 16 + itemGap) / (40 + itemGap)));
  const candidates = items
    .map((item, index) => ({ item, index }))
    .sort((a, b) => (b.item.priority ?? 0) - (a.item.priority ?? 0) || a.index - b.index);
  const visibleCapacity = candidates.length > capacity ? Math.max(1, capacity - 1) : capacity;
  const visibleSet = new Set(candidates.slice(0, visibleCapacity).map(({ item }) => item.id));

  visibleIds.value = items.filter((item) => visibleSet.has(item.id)).map((item) => item.id);
  overflowIds.value = items.filter((item) => !visibleSet.has(item.id)).map((item) => item.id);
}

function isActive(item: ToolbarItem) {
  return item.type === "panel" && activePanel.value?.id === item.id && !overflowOpen.value;
}

function clickItem(item: ToolbarItem) {
  if (item.type === "action") {
    emit("action", item);
    overflowOpen.value = false;
    return;
  }
  overflowOpen.value = false;
  activePanel.value = activePanel.value?.id === item.id ? null : item;
}

function selectOverflowItem(item: ToolbarItem) {
  if (item.type === "action") emit("action", item);
  else activePanel.value = item;
  overflowOpen.value = false;
}

function toggleOverflow() {
  activePanel.value = null;
  overflowOpen.value = !overflowOpen.value;
}

function closePanel() {
  activePanel.value = null;
  overflowOpen.value = false;
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "Escape" && isExpanded.value) closePanel();
}

function onDocumentPointerDown(event: PointerEvent) {
  if (!isExpanded.value || !shell.value) return;
  if (event.target instanceof Node && !shell.value.contains(event.target)) closePanel();
}

watch(() => items, scheduleLayout, { immediate: true });
watch(() => [config.maxWidth, config.gap], scheduleLayout);

onMounted(() => {
  scheduleLayout();
  if (typeof ResizeObserver !== "undefined") {
    resizeObserver = new ResizeObserver(scheduleLayout);
    if (shell.value) resizeObserver.observe(shell.value);
  }
  window.addEventListener("resize", scheduleLayout, { passive: true });
  document.addEventListener("keydown", onKeydown);
  document.addEventListener("pointerdown", onDocumentPointerDown);
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  if (typeof window !== "undefined") {
    window.removeEventListener("resize", scheduleLayout);
    cancelAnimationFrame(resizeFrame);
  }
  document.removeEventListener("keydown", onKeydown);
  document.removeEventListener("pointerdown", onDocumentPointerDown);
});
</script>

<template>
  <div
    ref="shell"
    class="floating-toolbar"
    :class="{
      'is-expanded': isExpanded,
      'has-panel': Boolean(activePanel),
      'has-overflow': overflowOpen,
    }"
    :style="shellStyle"
    :data-expanded="isExpanded"
  >
    <div class="floating-toolbar-panel-grid" :aria-hidden="!isExpanded">
      <section class="floating-toolbar-panel-clip">
        <div class="floating-toolbar-panel-content">
          <header class="floating-toolbar-panel-header">
            <h2 class="floating-toolbar-panel-title">
              {{ overflowOpen ? "更多操作" : activePanel?.label }}
            </h2>
            <button
              class="floating-toolbar-close"
              type="button"
              aria-label="关闭面板"
              @click="closePanel"
            >
              <span aria-hidden="true">×</span>
            </button>
          </header>

          <div v-if="overflowOpen" class="floating-toolbar-overflow-list">
            <button
              v-for="item in overflowItems"
              :key="item.id"
              class="floating-toolbar-overflow-item"
              type="button"
              @click="selectOverflowItem(item)"
            >
              <Icon
                v-if="item.icon"
                :name="item.icon"
                class="floating-toolbar-icon"
                aria-hidden="true"
              />
              <span class="floating-toolbar-overflow-label">{{ item.label }}</span>
            </button>
          </div>

          <component
            :is="activeComponent"
            v-else-if="activePanel"
            v-bind="activePanel.props"
            class="floating-toolbar-external-panel"
          />
        </div>
      </section>
    </div>

    <nav ref="bar" class="floating-toolbar-bar" :aria-label="label">
      <button
        v-for="item in visibleItems"
        :key="item.id"
        class="floating-toolbar-button"
        :class="{ 'is-active': isActive(item) }"
        type="button"
        :aria-label="item.label"
        :title="item.label"
        :aria-expanded="item.type === 'panel' ? isActive(item) : undefined"
        @click="clickItem(item)"
      >
        <Icon v-if="item.icon" :name="item.icon" class="floating-toolbar-icon" aria-hidden="true" />
        <span v-else class="floating-toolbar-fallback-icon" aria-hidden="true">
          {{ item.label.slice(0, 1) }}
        </span>
      </button>

      <button
        v-if="overflowItems.length"
        class="floating-toolbar-button floating-toolbar__more"
        :class="{ 'is-active': overflowOpen }"
        type="button"
        aria-label="更多操作"
        title="更多操作"
        :aria-expanded="overflowOpen"
        @click="toggleOverflow"
      >
        <span aria-hidden="true" class="floating-toolbar-more-glyph">•••</span>
      </button>
    </nav>

    <div ref="measureRow" class="floating-toolbar-measure" aria-hidden="true">
      <span v-for="item in items" :key="item.id" class="floating-toolbar-measure-button">
        <Icon v-if="item.icon" :name="item.icon" class="floating-toolbar-icon" />
        <span v-else>{{ item.label.slice(0, 1) }}</span>
      </span>
    </div>
  </div>
</template>

<style scoped>
/* 样式保持原样不变 */
.floating-toolbar {
  position: fixed;
  z-index: 1000;
  left: 50%;
  bottom: max(16px, env(safe-area-inset-bottom));
  display: grid;
  grid-template-rows: 0fr auto;
  max-width: calc(100vw - 32px);
  overflow: hidden;
  transform: translateX(-50%);
  color: var(--floating-toolbar-fg, #f5f5f5);
  background: var(--floating-toolbar-bg, rgb(30 30 32 / 94%));
  border: 1px solid var(--floating-toolbar-border, rgb(255 255 255 / 10%));
  border-radius: 18px;
  box-shadow: 0 10px 36px rgb(0 0 0 / 22%);
  backdrop-filter: blur(18px) saturate(1.2);
  transition:
    width 280ms cubic-bezier(0.2, 0.8, 0.2, 1),
    border-radius 220ms ease,
    box-shadow 220ms ease;
}

.floating-toolbar.is-expanded {
  grid-template-rows: minmax(0, 1fr) auto;
  border-radius: 20px;
}

.floating-toolbar-panel-grid {
  min-height: 0;
  overflow: hidden;
}

.floating-toolbar-panel-clip {
  min-height: 0;
  overflow: hidden;
}

.floating-toolbar-panel-content {
  min-height: 0;
  padding: 14px 14px 12px;
  opacity: 0;
  transform: translateY(8px);
  transition:
    opacity 160ms ease,
    transform 240ms cubic-bezier(0.2, 0.8, 0.2, 1);
}

.is-expanded .floating-toolbar-panel-content {
  opacity: 1;
  transform: translateY(0);
  transition-delay: 40ms;
}

.floating-toolbar-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.floating-toolbar-panel-title {
  min-width: 0;
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
}

.floating-toolbar-close,
.floating-toolbar-button {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  padding: 0;
  color: inherit;
  background: transparent;
  border: 0;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

.floating-toolbar-close {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  color: var(--floating-toolbar-muted, #b7b7bb);
  font-size: 22px;
  line-height: 1;
}

.floating-toolbar-button {
  width: 40px;
  height: 40px;
  border-radius: 11px;
  transition:
    background-color 140ms ease,
    color 140ms ease,
    transform 140ms ease;
}

.floating-toolbar-overflow-item {
  display: flex;
  align-items: center;
  gap: 9px;
  min-width: 0;
  padding: 10px;
  color: inherit;
  text-align: left;
  background: transparent;
  border: 0;
  border-radius: 10px;
  cursor: pointer;
}

.floating-toolbar-close:hover,
.floating-toolbar-button:hover,
.floating-toolbar-overflow-item:hover {
  background: var(--floating-toolbar-hover, rgb(255 255 255 / 10%));
}

.floating-toolbar-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--toolbar-gap);
  min-width: 0;
  padding: 8px;
}

.floating-toolbar-button:active {
  transform: scale(0.94);
}

.floating-toolbar-button.is-active {
  color: var(--floating-toolbar-active-fg, #fff);
  background: var(--floating-toolbar-active-bg, rgb(255 255 255 / 16%));
}

.floating-toolbar-icon {
  width: 19px;
  height: 19px;
  flex: 0 0 auto;
}

.floating-toolbar-fallback-icon {
  font-size: 15px;
  font-weight: 600;
}

.floating-toolbar-more-glyph {
  letter-spacing: 1px;
  font-size: 16px;
  transform: translateY(-3px);
}

.floating-toolbar-overflow-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(110px, 1fr));
  gap: 6px;
  max-height: min(45vh, 320px);
  overflow: auto;
}

.floating-toolbar-overflow-label {
  overflow: hidden;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.floating-toolbar-external-panel {
  display: block;
  min-width: 0;
  max-width: 100%;
}

.floating-toolbar-measure {
  position: absolute;
  top: 0;
  left: 0;
  display: flex;
  gap: var(--toolbar-gap);
  width: max-content;
  height: 0;
  overflow: hidden;
  visibility: hidden;
  pointer-events: none;
}

.floating-toolbar-measure-button {
  display: inline-flex;
  width: 40px;
  height: 40px;
  flex: 0 0 40px;
  align-items: center;
  justify-content: center;
}

@media (prefers-reduced-motion: reduce) {
  .floating-toolbar,
  .floating-toolbar-panel-content,
  .floating-toolbar-button {
    transition: none;
  }
}
</style>
