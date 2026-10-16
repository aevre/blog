<script setup lang="ts">
const { spacer = true } = defineProps<{ spacer?: boolean }>();
</script>

<template>
  <div class="container">
    <BaseSpacer v-if="spacer" height="80" />
    <div class="blur-layer-container" aria-hidden="true"></div>
    <slot />
  </div>
</template>

<style lang="less" scoped>
.container {
  position: relative;
  isolation: isolate;
  animation: opacity 0.5s ease;
  padding-bottom: 140px;
  min-height: 100vh;
}

.blur-layer-container {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: -1;
  border-radius: inherit;
  background: var(--bg-content);
  backdrop-filter: blur(var(--blur-surface));
}
</style>
