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

defineOptions({ inheritAttrs: false });

const collapsed = defineModel<boolean>("collapsed", {
  default: false,
});

const hovered = ref(false);
const focused = ref(false);
const isCollapsed = computed(
  () => collapsed.value && !hovered.value && !focused.value && !isExpanded.value,
);

const { items, config, label = "" } = defineProps<Props>();
const emit = defineEmits<Emits>();
const activePanelId = shallowRef<ToolbarItem["id"] | null>(null);
const activePanel = computed(() =>
  items.find(
    (item): item is ToolbarPanel => item.type === "panel" && item.id === activePanelId.value,
  ),
);

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

const hoverZoneStyle = computed<CSSProperties>(() => ({
  width: `${isExpanded.value ? expandedWidth.value : barWidth.value}px`,
  maxWidth: `min(${maxWidth.value}px, calc(100vw - 32px))`,
  height: `calc(${buttonSize.value + 18}px + max(16px, env(safe-area-inset-bottom)))`,
}));

const shellStyle = computed<CSSProperties>(() => ({
  "--toolbar-gap": `${gap.value}px`,
  "--toolbar-expanded-width": `${expandedWidth.value}px`,
  "--toolbar-max-width": `${maxWidth.value}px`,
  "--toolbar-button-size": `${buttonSize.value}px`,
  "--toolbar-icon-size": `${iconSize.value}px`,
  width: isCollapsed.value
    ? "44px"
    : isExpanded.value
      ? "var(--toolbar-expanded-width)"
      : `${barWidth.value}px`,
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
  // await nextTick();

  windowWidth.value = document.documentElement.clientWidth;
  const available = availableWidth.value;

  const candidates = items
    .map((item, index) => ({ item, index }))
    .sort((a, b) => (b.item.priority ?? 0) - (a.item.priority ?? 0) || a.index - b.index);

  const capacity = Math.max(
    1,
    Math.floor((available - 16 + gap.value) / (buttonSize.value + gap.value)),
  );

  const needOverflow = candidates.length > capacity;
  const visibleCapacity = needOverflow ? Math.max(0, capacity - 1) : capacity;

  const visibleSet = new Set(candidates.slice(0, visibleCapacity).map(({ item }) => item.id));

  visibleIds.value = items.filter((item) => visibleSet.has(item.id)).map((item) => item.id);
  overflowIds.value = items.filter((item) => !visibleSet.has(item.id)).map((item) => item.id);
}

function isActive(item: ToolbarItem) {
  return item.type === "panel" && activePanelId.value === item.id && !overflowOpen.value;
}

function clickItem(item: ToolbarItem) {
  if (item.type === "action") {
    emit("action", item);
    closePanel();
    return;
  }
  if (isActive(item)) {
    closePanel();
    return;
  }
  activePanelId.value = item.id;
  overflowOpen.value = false;
}

function toggleOverflow() {
  if (overflowOpen.value) {
    closePanel();
    return;
  }
  activePanelId.value = null;
  overflowOpen.value = true;
}

function closePanel() {
  activePanelId.value = null;
  overflowOpen.value = false;
  const activeElement = document.activeElement;
  if (activeElement instanceof HTMLElement && shell.value?.contains(activeElement)) {
    activeElement.blur();
  }
  focused.value = false;
}

function handlePointerEnter(event: PointerEvent) {
  if (event.pointerType === "mouse") {
    hovered.value = true;
  }
}

function handlePointerLeave(event: PointerEvent) {
  if (event.pointerType === "mouse") {
    hovered.value = false;
  }
}

async function expandToolbar() {
  collapsed.value = false;
  await nextTick();
  shell.value
    ?.querySelector<HTMLButtonElement>(".floating-toolbar-bar button")
    ?.focus({ preventScroll: true });
}

function handleFocusOut(event: FocusEvent) {
  focused.value =
    event.relatedTarget instanceof Node && Boolean(shell.value?.contains(event.relatedTarget));
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "Escape" && isExpanded.value) closePanel();
}

function onDocumentPointerDown(event: PointerEvent) {
  if (!isExpanded.value || !shell.value) return;
  if (event.target instanceof Node && !shell.value.contains(event.target)) closePanel();
}

function observePanelHeight(element: HTMLElement, onCleanup: (cleanup: () => void) => void) {
  const measure = () => {
    panelHeight.value = element.offsetHeight;
  };

  measure();

  if (typeof ResizeObserver === "undefined") return;
  const observer = new ResizeObserver(measure);
  observer.observe(element);
  onCleanup(() => observer.disconnect());
}

// 测量内部自然高度，让外层在明确的像素高度之间过渡；异步内容变化也会更新。
watch(
  panelClip,
  (panel, _previous, onCleanup) => {
    if (panel) observePanelHeight(panel, onCleanup);
  },
  { flush: "post" },
);

watch(() => items.map(({ id, priority, label }) => [id, priority, label]), scheduleLayout, {
  immediate: true,
});
watch(() => [config.maxWidth, config.gap, config.iconSize], scheduleLayout);

onMounted(() => {
  scheduleLayout();
  if (typeof ResizeObserver !== "undefined") {
    resizeObserver = new ResizeObserver(scheduleLayout);
    resizeObserver.observe(document.documentElement);
  }
  document.addEventListener("keydown", onKeydown);
  document.addEventListener("pointerdown", onDocumentPointerDown);
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  cancelAnimationFrame(resizeFrame);
  document.removeEventListener("keydown", onKeydown);
  document.removeEventListener("pointerdown", onDocumentPointerDown);
});
</script>

<template>
  <div
    v-if="visibleItems.length > 0 || overflowItems.length > 0"
    class="floating-toolbar-hover-zone"
    :class="{ 'is-hovered': hovered }"
    :style="hoverZoneStyle"
    @pointerenter="handlePointerEnter"
    @pointerleave="handlePointerLeave"
  >
    <div
      ref="shell"
      v-bind="$attrs"
      class="floating-toolbar"
      :class="{
        'is-collapsed': isCollapsed,
        'is-expanded': isExpanded,
        'has-panel': Boolean(activePanel),
        'has-overflow': overflowOpen,
      }"
      :style="shellStyle"
      :data-expanded="isExpanded"
      @focusout="handleFocusOut"
    >
      <div
        class="floating-toolbar-panel-grid"
        :style="{ height: `${isExpanded ? panelHeight : 0}px` }"
        :aria-hidden="!isExpanded"
        :inert="!isExpanded"
        @focusin="focused = true"
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

      <nav
        ref="bar"
        class="floating-toolbar-bar"
        :aria-label="label"
        :aria-hidden="isCollapsed"
        :inert="isCollapsed"
        @focusin="focused = true"
      >
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

      <button
        class="floating-toolbar-expand"
        type="button"
        aria-label="展开工具栏"
        :aria-expanded="!isCollapsed"
        :aria-hidden="!isCollapsed"
        :inert="!isCollapsed"
        @click="expandToolbar"
      >
        <Icon name="iconamoon:arrow-up-2-fill" aria-hidden="true" />
      </button>
    </div>
  </div>
</template>

<style scoped>
/* 命中区域覆盖展开后的按钮栏和底部空隙，不随收起动画改变尺寸。 */
.floating-toolbar-hover-zone {
  position: fixed;
  z-index: 1000;
  left: 0;
  right: 0;
  bottom: 0;
  margin-inline: auto;
  pointer-events: none;
}

/* 先进入可见工具栏才启用整个区域，收起时不遮挡两侧页面。 */
.floating-toolbar-hover-zone.is-hovered {
  pointer-events: auto;
}

.floating-toolbar {
  --toolbar-surface: rgb(255 255 255 / 58%);

  position: fixed;
  z-index: 1000;
  left: 50%;
  bottom: max(16px, env(safe-area-inset-bottom));
  display: grid;
  grid-template-rows: auto auto;
  box-sizing: border-box;
  pointer-events: auto;
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
    bottom 230ms cubic-bezier(0.2, 0.8, 0.2, 1),
    border-radius 220ms ease,
    box-shadow 220ms ease;
}

.floating-toolbar.is-collapsed {
  bottom: 0;
  border-radius: 44px 44px 0 0;
}

.floating-toolbar-expand {
  box-sizing: border-box;
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 0 0 env(safe-area-inset-bottom);
  border: 0;
  border-radius: inherit;
  color: inherit;
  background: transparent;
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  transition:
    opacity 140ms ease,
    visibility 0s 140ms;
}

.is-collapsed .floating-toolbar-expand {
  opacity: 1;
  visibility: visible;
  pointer-events: auto;
  transition-delay: 90ms, 0s;
}

.floating-toolbar-expand:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: -4px;
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
  min-width: max-content;
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
  overscroll-behavior: contain;
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
  gap: 9px;
  align-items: center;
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
  box-sizing: border-box;
  height: calc(var(--toolbar-button-size) + 16px);
  min-height: 0;
  opacity: 1;
  visibility: visible;
  transition:
    height 230ms cubic-bezier(0.2, 0.8, 0.2, 1),
    padding 230ms ease,
    opacity 140ms ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--toolbar-gap);
  min-width: 0;
  padding: 8px;
}

.is-collapsed .floating-toolbar-bar {
  height: calc(24px + env(safe-area-inset-bottom));
  padding: 0;
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
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
  --toolbar-fallback-icon-size: calc(var(--toolbar-icon-size) - 6px);

  font-size: var(--toolbar-fallback-icon-size);
  font-weight: 600;
  display: inline-block;
  line-height: var(--toolbar-fallback-icon-size);
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

[data-theme="dark"] .floating-toolbar {
  --toolbar-surface: rgb(34 34 38 / 62%);
}

@media (prefers-reduced-motion: reduce) {
  .floating-toolbar,
  .floating-toolbar-panel-grid,
  .floating-toolbar-panel-content,
  .floating-toolbar-button,
  .floating-toolbar-bar,
  .floating-toolbar-expand,
  .toolbar-panel-enter-active,
  .toolbar-panel-leave-active {
    transition: none;
  }
}
</style>
