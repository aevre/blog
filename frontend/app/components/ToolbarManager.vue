<script setup lang="ts">
import type { MessageType } from "~/types/ui/message";
import type { ToolbarContext, ToolbarItem } from "./FloatToolbar/types";

const toolbar = useToolbar();
const message = useMessage();
const { $ts } = useI18n();

const items = computed<ToolbarItem[]>(() =>
  toolbar.items.value
    .filter((item) => item.visible !== false)
    .map((item) => ({
      ...item,
      label: $ts(item.label),
    })),
);

const scrollCollapsed = ref(false);

function handleAction(item: Extract<ToolbarItem, { type: "action" }>) {
  const context: ToolbarContext = {
    notify: (msg, options) => {
      const type = options?.type ?? "info";
      message.create(type as MessageType, $ts(msg));
    },
  };
  void item.handler(context);
}

let lastY = 0;
let distance = 0;

watch(scrollCollapsed, () => {
  distance = 0;
});

function getScrollY() {
  const maxY = Math.max(0, document.documentElement.scrollHeight - window.innerHeight);

  return Math.min(maxY, Math.max(0, window.scrollY));
}

function handleScroll() {
  const y = getScrollY();
  const delta = y - lastY;
  lastY = y;

  if (y <= 64) {
    scrollCollapsed.value = false;
    distance = 0;
    return;
  }

  if (delta === 0) return;

  distance = Math.sign(delta) === Math.sign(distance) ? distance + delta : delta;

  if (Math.abs(distance) >= 12) {
    scrollCollapsed.value = distance > 0;
    distance = 0;
  }
}

onMounted(() => {
  lastY = getScrollY();
  window.addEventListener("scroll", handleScroll, { passive: true });
});

onBeforeUnmount(() => {
  window.removeEventListener("scroll", handleScroll);
});
</script>

<template>
  <FloatToolbar
    v-model:collapsed="scrollCollapsed"
    :items="items"
    :config="toolbar.config.value"
    @action="handleAction"
  />
</template>
