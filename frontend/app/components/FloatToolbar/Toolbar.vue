<script setup lang="ts">
import type { CSSProperties } from "vue";
import type { ToolbarConfig, ToolbarItem, ToolbarPanel } from "./types";

interface Props {
  items: ToolbarItem[];
  config: ToolbarConfig;
  label?: string;
}

interface Emits {
  (event: "action", item: Extract<ToolbarItem, { type: "action" }>): void;
}

const { items, config, label = "" } = defineProps<Props>();
const emit = defineEmits<Emits>();
const activePanel = shallowRef<ToolbarPanel | null>(null);
const overflowOpen = shallowRef(false);
const overflowPanelKey = Symbol("toolbar-overflow");
const visibleIds = shallowRef<ToolbarItem["id"][]>([]);
const overflowIds = shallowRef<ToolbarItem["id"][]>([]);
const shell = useTemplateRef("shell");
const panelClip = useTemplateRef("panelClip");
const panelHeight = shallowRef(0);
const windowWidth = shallowRef<number | null>(null);
let resizeObserver: ResizeObserver | undefined;
let resizeFrame = 0;

const iconSize = computed(() => Math.max(12, config.iconSize ?? 19));
const buttonSize = computed(() => Math.max(40, iconSize.value + 20));

const maxWidth = computed(() => Math.max(240, config.maxWidth ?? 480));
const gap = computed(() => Math.max(0, config.gap ?? 6));
const isExpanded = computed(() => Boolean(activePanel.value || overflowOpen.value));

const barWidth = computed(() => {
  const count = visibleItems.value.length + (overflowItems.value.length ? 1 : 0);
  return 16 + count * buttonSize.value + Math.max(0, count - 1) * gap.value;
});

const availableWidth = computed(() =>
  Math.min(maxWidth.value, Math.max(0, (windowWidth.value ?? maxWidth.value) - 32)),
);
const expandedWidth = computed(() => {
  const requested = activePanel.value?.width ?? 0;
  return Math.min(Math.max(240, barWidth.value, requested), availableWidth.value);
});

const shellStyle = computed<CSSProperties>(() => ({
  "--toolbar-gap": `${gap.value}px`,
  "--toolbar-expanded-width": `${expandedWidth.value}px`,
  "--toolbar-max-width": `${maxWidth.value}px`,
  "--toolbar-button-size": `${buttonSize.value}px`,
  width: isExpanded.value ? "var(--toolbar-expanded-width)" : `${barWidth.value}px`,
}));

const activeComponent = computed(() => activePanel.value?.component);
const activePanelProps = computed(() =>
  Object.fromEntries(
    Object.entries(activePanel.value?.props ?? {}).map(([key, value]) => [key, toValue(value)]),
  ),
);
const visibleItems = computed(() => items.filter((item) => visibleIds.value.includes(item.id)));
const overflowItems = computed(() => items.filter((item) => overflowIds.value.includes(item.id)));

function scheduleLayout() {
  if (typeof window === "undefined") return;
  cancelAnimationFrame(resizeFrame);
  resizeFrame = requestAnimationFrame(() => void calculateLayout());
}

async function calculateLayout() {
  await nextTick();

  windowWidth.value = document.documentElement.clientWidth;
  const itemGap = gap.value;
  const capacity = Math.max(
    1,
    Math.floor((availableWidth.value - 16 + itemGap) / (buttonSize.value + itemGap)),
  );
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

// 测量内部自然高度，让外层在明确的像素高度之间过渡；异步内容变化也会更新。
watch(
  panelClip,
  (panel, _previous, onCleanup) => {
    if (!panel) return;
    const measure = () => {
      panelHeight.value = panel.offsetHeight;
    };
    measure();
    if (typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(measure);
    observer.observe(panel);
    onCleanup(() => observer.disconnect());
  },
  { flush: "post" },
);

watch(() => items, scheduleLayout, { immediate: true });
watch(() => [config.maxWidth, config.gap, config.iconSize], scheduleLayout);

onMounted(() => {
  scheduleLayout();
  if (typeof ResizeObserver !== "undefined") {
    resizeObserver = new ResizeObserver(scheduleLayout);
    resizeObserver.observe(document.documentElement);
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
    v-if="visibleItems.length > 0 || overflowItems.length > 0"
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
    <div
      v-if="visibleItems.length > 0"
      class="floating-toolbar-panel-grid"
      :style="{ height: `${isExpanded ? panelHeight : 0}px` }"
      :aria-hidden="!isExpanded"
    >
      <section ref="panelClip" class="floating-toolbar-panel-clip">
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

          <div class="floating-toolbar-panel-body">
            <Transition
              name="toolbar-panel"
              @before-enter="(element) => element.removeAttribute('inert')"
              @before-leave="(element) => element.setAttribute('inert', '')"
            >
              <div v-if="isExpanded" :key="overflowOpen ? overflowPanelKey : activePanel?.id">
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
                      :size="iconSize"
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
                  v-bind="activePanelProps"
                  class="floating-toolbar-external-panel"
                />
              </div>
            </Transition>
          </div>
        </div>
      </section>
    </div>

    <nav v-if="visibleItems.length > 0" ref="bar" class="floating-toolbar-bar" :aria-label="label">
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
        <Icon
          v-if="item.icon"
          :size="iconSize"
          :name="item.icon"
          class="floating-toolbar-icon"
          aria-hidden="true"
        />
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
  </div>
</template>

<style scoped>
.floating-toolbar {
  --toolbar-surface: rgb(255 255 255 / 58%);
  --toolbar-panel-surface: rgb(246 245 244 / 46%);

  position: fixed;
  z-index: 1000;
  left: 50%;
  bottom: max(16px, env(safe-area-inset-bottom));
  display: grid;
  grid-template-rows: auto auto;
  max-width: min(var(--toolbar-max-width), calc(100vw - 32px));
  overflow: hidden;
  transform: translateX(-50%);
  color: var(--text-primary);
  background: var(--toolbar-surface);
  border: 1px solid var(--border-glass);
  border-radius: 18px;
  box-shadow: var(--shadow-lg);
  backdrop-filter: blur(var(--blur-elevated));
  transition:
    width 230ms cubic-bezier(0.2, 0.8, 0.2, 1),
    border-radius 220ms ease,
    box-shadow 220ms ease;
}

.floating-toolbar.is-expanded {
  background: var(--toolbar-surface);
  border-radius: 20px;
}

.floating-toolbar-panel-grid {
  min-height: 0;
  height: 0;
  overflow: hidden;
  background: transparent;
  transition: height 230ms cubic-bezier(0.2, 0.8, 0.2, 1);
}

.floating-toolbar-panel-clip {
  display: flex;
  min-height: 0;
  max-height: min(60vh, 420px);
  flex-direction: column;
}

.floating-toolbar-panel-content {
  display: flex;
  min-height: 0;
  flex-direction: column;
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
  flex: 0 0 auto;
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

.floating-toolbar-panel-body {
  position: relative;
  min-height: 0;
  overflow: auto;
  scrollbar-width: none;
}

.floating-toolbar-panel-body::-webkit-scrollbar {
  display: none;
}

.toolbar-panel-enter-active,
.toolbar-panel-leave-active {
  transition: opacity 230ms cubic-bezier(0.2, 0.8, 0.2, 1);
}

/* 旧内容退出文档流，测量新内容高度时不会把两份内容相加。 */
.toolbar-panel-leave-active {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  max-height: 100%;
  overflow: hidden;
  pointer-events: none;
}

.toolbar-panel-enter-from,
.toolbar-panel-leave-to {
  opacity: 0;
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
  color: var(--text-secondary);
  font-size: 22px;
  line-height: 1;
}

.floating-toolbar-button {
  width: var(--toolbar-button-size);
  height: var(--toolbar-button-size);
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
  background: var(--bg-interactive-hover);
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
  color: var(--text-primary);
  background: var(--bg-interactive-active);
}

.floating-toolbar-icon {
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
  width: 100%;
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

[data-theme="dark"] .floating-toolbar {
  --toolbar-surface: rgb(34 34 38 / 62%);
  --toolbar-panel-surface: rgb(26 26 30 / 50%);
}

@media (prefers-reduced-motion: reduce) {
  .floating-toolbar,
  .floating-toolbar-panel-grid,
  .floating-toolbar-panel-content,
  .floating-toolbar-button,
  .toolbar-panel-enter-active,
  .toolbar-panel-leave-active {
    transition: none;
  }
}
</style>
