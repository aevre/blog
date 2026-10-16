<script setup lang="ts">
import type { TimerId } from "~/types/timer";

type Placement = "top" | "left" | "right" | "bottom";
type PlacementOption = Placement | "auto";

interface Props {
  placement?: PlacementOption;
  maxWidth?: number | string;
  content?: string;
}

interface TooltipStyle {
  left: string;
  top: string;
  maxWidth: string;
}

const { content = "", placement = "auto", maxWidth = 320 } = defineProps<Props>();

const slots = useSlots();

const contentEmpty = computed(() => content === "" && slots.content === undefined);

const OPEN_DELAY = 80;
const CLOSE_DELAY = 80;

const isOpen = ref(false);

const targetRef = useTemplateRef<HTMLDivElement>("target");
const tooltipRef = useTemplateRef<HTMLDivElement>("tooltip");

const placements: Placement[] = ["top", "bottom", "left", "right"];
const actualPlacement = ref<Placement>("top");

let openTimer: TimerId = null;
let closeTimer: TimerId = null;

const toCssSize = (value: number | string) => (typeof value === "number" ? `${value}px` : value);
const maxWidthCss = computed(() => toCssSize(maxWidth));

const style = ref<TooltipStyle>({
  left: "0",
  top: "0",
  maxWidth: maxWidthCss.value,
});

const computePosition = (
  target: DOMRect,
  tooltipSize: { width: number; height: number },
  placement: Placement,
) => {
  const { width, height } = tooltipSize;
  let left = 0;
  let top = 0;

  switch (placement) {
    case "top":
      left = target.left + target.width / 2 - width / 2;
      top = target.top - height;
      break;
    case "bottom":
      left = target.left + target.width / 2 - width / 2;
      top = target.bottom;
      break;
    case "left":
      left = target.left - width;
      top = target.top + target.height / 2 - height / 2;
      break;
    case "right":
      left = target.right;
      top = target.top + target.height / 2 - height / 2;
      break;
  }

  return {
    left,
    top,
  };
};

const overflowScope = (
  left: number,
  top: number,
  tooltipSize: { width: number; height: number },
) => {
  const vw = window.innerWidth;
  const vh = window.innerHeight;
  const { width, height } = tooltipSize;

  const overLeft = Math.max(0, -left);
  const overRight = Math.max(0, left + width - vw);
  const overTop = Math.max(0, -top);
  const overBottom = Math.max(0, top + height - vh);

  return overLeft + overRight + overTop + overBottom;
};

const choosePlacement = (targetRect: DOMRect, tooltipSize: { width: number; height: number }) => {
  let bestPlacement: Placement = "top";
  let bestScore = Number.MAX_SAFE_INTEGER;
  let bestLeft = 0;
  let bestTop = 0;
  const candidates =
    placement !== "auto"
      ? ([placement, ...placements.filter((p) => p !== placement)] as Placement[])
      : placements;

  for (const p of candidates) {
    const { left, top } = computePosition(targetRect, tooltipSize, p);
    const score = overflowScope(left, top, tooltipSize);
    if (score < bestScore) {
      bestScore = score;
      bestPlacement = p;
      bestLeft = left;
      bestTop = top;
      if (bestScore === 0) break;
    }
  }

  return { placement: bestPlacement, left: bestLeft, top: bestTop };
};

const caculateStyle: () => TooltipStyle | null = () => {
  const targetEl = targetRef.value;
  const tooltipEl = tooltipRef.value;
  if (!targetEl || !tooltipEl) return null;

  const targetRect = targetEl.getBoundingClientRect();
  const tooltipSize = {
    width: tooltipEl.offsetWidth,
    height: tooltipEl.offsetHeight,
  };

  const result = choosePlacement(targetRect, tooltipSize);
  actualPlacement.value = result.placement;

  return {
    left: toCssSize(result.left),
    top: toCssSize(result.top),
    maxWidth: maxWidthCss.value,
  };
};

const clearTimers = () => {
  if (openTimer) {
    clearTimeout(openTimer);
    openTimer = null;
  }
  if (closeTimer) {
    clearTimeout(closeTimer);
    closeTimer = null;
  }
};

let rafId = 0;
const updatePosition = () => {
  if (!isOpen.value) return;
  if (rafId) return;
  rafId = requestAnimationFrame(() => {
    rafId = 0;
    const pos = caculateStyle();
    if (pos) style.value = pos;
  });
};

const openWithDelay = () => {
  clearTimers();
  if (contentEmpty.value) return;
  openTimer = setTimeout(async () => {
    isOpen.value = true;
    await nextTick();
    updatePosition();
  }, OPEN_DELAY);
};

const closeWithDelay = () => {
  clearTimers();
  if (contentEmpty.value) return;
  closeTimer = setTimeout(() => {
    isOpen.value = false;
  }, CLOSE_DELAY);
};

watchEffect((onCleanup) => {
  if (contentEmpty.value) return;
  if (!isOpen.value) return;

  const handler = () => updatePosition();
  window.addEventListener("scroll", handler, true);
  window.addEventListener("resize", handler);

  onCleanup(() => {
    window.removeEventListener("scroll", handler, true);
    window.removeEventListener("resize", handler);
  });
});

onBeforeUnmount(() => {
  clearTimers();
  if (rafId) cancelAnimationFrame(rafId);
});
</script>

<template>
  <div
    ref="target"
    class="tooltip-target"
    @mouseenter="openWithDelay"
    @mouseleave="closeWithDelay"
    @focusin="openWithDelay"
    @focusout="closeWithDelay"
  >
    <slot />
  </div>
  <Teleport to="body">
    <Transition name="tooltip-fader">
      <div
        v-if="isOpen && (content || $slots.content)"
        ref="tooltip"
        class="tooltip"
        :style
        @mouseenter="clearTimers"
        @mouseleave="closeWithDelay"
      >
        <div class="tooltip-content" :data-placement="actualPlacement">
          <slot v-if="$slots.content" name="content" />
          <span v-else>{{ content }}</span>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style lang="less" scoped>
.tooltip {
  position: fixed;
  font-size: 1.4rem;
  line-height: 1.2;
  z-index: 9999;
  transform-origin: bottom center;
  padding: 10px;
}

.tooltip-target {
  cursor: default;
  display: inline-block;
}

.tooltip-content {
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  background-color: var(--tooltip-bg);
  backdrop-filter: blur(var(--blur-elevated));
  border: 1px solid var(--tooltip-border-color, var(--border-default));
  color: var(--text-primary);
  box-shadow: var(--shadow-sm);
  position: relative;
  pointer-events: auto;

  --arrow-size: 8px;
  --arrow-border: calc(var(--arrow-size) + 1px);
  --arrow-color: var(--tooltip-arrow-color, var(--tooltip-bg, var(--bg-card-base)));
  --arrow-border-color: var(
    --tooltip-arrow-border-color,
    var(--tooltip-border-color, var(--border-default))
  );
  --arrow-inset-fix: 1px;
}

.tooltip-content::before,
.tooltip-content::after {
  content: "";
  position: absolute;
  width: 0;
  height: 0;
  border-style: solid;
  pointer-events: none;
  transition: none;
}

.tooltip-content[data-placement="top"]::before {
  left: 50%;
  top: 100%;
  transform: translateX(-50%);
  border-width: var(--arrow-border) var(--arrow-border) 0 var(--arrow-border);
  border-color: var(--arrow-border-color) transparent transparent transparent;
}

.tooltip-content[data-placement="top"]::after {
  left: 50%;
  top: calc(100% - var(--arrow-inset-fix));
  transform: translateX(-50%);
  border-width: var(--arrow-size) var(--arrow-size) 0 var(--arrow-size);
  border-color: var(--arrow-color) transparent transparent transparent;
}

.tooltip-content[data-placement="bottom"]::before {
  left: 50%;
  bottom: 100%;
  transform: translateX(-50%);
  border-width: 0 var(--arrow-border) var(--arrow-border) var(--arrow-border);
  border-color: transparent transparent var(--arrow-border-color) transparent;
}

.tooltip-content[data-placement="bottom"]::after {
  left: 50%;
  bottom: calc(100% - var(--arrow-inset-fix));
  transform: translateX(-50%);
  border-width: 0 var(--arrow-size) var(--arrow-size) var(--arrow-size);
  border-color: transparent transparent var(--arrow-color) transparent;
}

.tooltip-content[data-placement="left"]::before {
  top: 50%;
  left: 100%;
  transform: translateY(-50%);
  border-width: var(--arrow-border) 0 var(--arrow-border) var(--arrow-border);
  border-color: transparent transparent transparent var(--arrow-border-color);
}

.tooltip-content[data-placement="left"]::after {
  top: 50%;
  left: calc(100% - var(--arrow-inset-fix));
  transform: translateY(-50%);
  border-width: var(--arrow-size) 0 var(--arrow-size) var(--arrow-size);
  border-color: transparent transparent transparent var(--arrow-color);
}

.tooltip-content[data-placement="right"]::before {
  top: 50%;
  right: 100%;
  transform: translateY(-50%);
  border-width: var(--arrow-border) var(--arrow-border) var(--arrow-border) 0;
  border-color: transparent var(--arrow-border-color) transparent transparent;
}

.tooltip-content[data-placement="right"]::after {
  top: 50%;
  right: calc(100% - var(--arrow-inset-fix));
  transform: translateY(-50%);
  border-width: var(--arrow-size) var(--arrow-size) var(--arrow-size) 0;
  border-color: transparent var(--arrow-color) transparent transparent;
}

.tooltip-fader-enter-active,
.tooltip-fader-leave-active {
  transition: opacity 0.1s ease;
}

.tooltip-fader-enter-from,
.tooltip-fader-leave-to {
  opacity: 0;
}
</style>
