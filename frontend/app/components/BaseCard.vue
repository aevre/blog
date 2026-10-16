<script setup lang="ts"></script>

<template>
  <div class="card">
    <slot />
  </div>
</template>

<style lang="less" scoped>
.card {
  position: relative;
  background: var(--bg-card-base);
  backdrop-filter: blur(var(--blur-surface));
  border: 1px solid var(--border-color-card);
  box-shadow: var(--shadow-card-base);
  border-radius: var(--radius-card);
  transition:
    box-shadow 0.3s ease,
    background 0.3s ease,
    border-color 0.3s ease,
    transform 0.3s ease;

  &:hover {
    box-shadow: var(--shadow-card-hover);
    background: var(--bg-card-hover);
    border-color: var(--border-color-card-hover);
    transform: translateY(-1px);
  }
}
</style>
