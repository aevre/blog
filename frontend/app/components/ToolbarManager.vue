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

function handleAction(item: Extract<ToolbarItem, { type: "action" }>) {
  const context: ToolbarContext = {
    notify: (msg, options) => {
      const type = options?.type ?? "info";
      message.create(type as MessageType, $ts(msg));
    },
  };
  void item.handler(context);
}
</script>

<template>
  <FloatToolbar :items="items" :config="toolbar.config.value" @action="handleAction" />
</template>
