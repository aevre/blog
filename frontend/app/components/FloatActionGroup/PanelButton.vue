<script setup lang="ts">
import { onClickOutside } from "@vueuse/core";
import type { Component } from "vue";

const { icon, component } = defineProps<{
  icon: string;
  component: Component;
}>();

const open = ref(false);

const panelContent = useTemplateRef<HTMLDivElement>("panelContent");
const actionButton = useTemplateRef<HTMLButtonElement>("actionButton");

onClickOutside(panelContent, (event) => {
  // 判断点击事件触发位置是否是 action button
  // 如果是则不关闭，防止 panel content 闪烁
  if (actionButton.value?.contains(event.target as Node)) return;
  open.value = false;
});
</script>

<template>
  <div class="panel-button-wrapper">
    <button ref="actionButton" class="action-button" @click="open = !open">
      <Icon :name="icon" size="24" class="icon" />
    </button>

    <div ref="panelContent" class="panel-content">
      <Component :is="component" v-model:show="open" />
    </div>
  </div>
</template>

<style lang="less" scoped>
.panel-button-wrapper {
  position: relative;
}

.icon {
  color: var(--text-primary);
}

.action-button {
  background: var(--bg-nav-base);
  backdrop-filter: blur(var(--blur-elevated));
  position: relative;
  border-radius: var(--radius-nav);
  width: 100%;
  height: 40px;
  display: flex;
  justify-content: center;
  align-items: center;
  border: 1px solid var(--border-color-nav);
  cursor: pointer;
  box-shadow: var(--shadow-nav);
  transition: var(--transition-nav);
  overflow: hidden;
}

.panel-content {
  position: absolute;
  max-width: 300px;
  right: 120%;
  bottom: 0;
  z-index: 10;
}
</style>
