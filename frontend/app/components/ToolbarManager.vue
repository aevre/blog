<script setup lang="ts">
import type { ToolbarItem } from "./FloatToolbar/types";

const toolbar = useToolbar();
const message = useMessage();

const items = computed(() => toolbar.items.value.filter((item) => item.visible !== false));

function handleAction(item: Extract<ToolbarItem, { type: "action" }>) {
  void item.handler({ message });
}
</script>

<template>
  <FloatToolbar :items="items" :config="toolbar.config.value" @action="handleAction" />
</template>
