<script setup lang="ts">
interface Props {
  isSticky?: boolean;
}

const { isSticky = false } = defineProps<Props>();
</script>

<template>
  <header ref="header" :class="{ header: true }">
    <div class="header-container" :class="{ 'is-sticky': isSticky }">
      <NavbarWrapper />
    </div>
  </header>
</template>

<style lang="less" scoped>
.header {
  position: fixed;
  width: 100%;
  height: 100px;
  z-index: 1999;
}

.header-container {
  position: relative;
  max-width: 1000px;
  margin: auto;
  background: var(--bg-nav-base);
  backdrop-filter: blur(var(--blur-elevated));
  border: 1px solid var(--border-color-nav);
  overflow: hidden;
  width: calc(100% - 40px);
  height: 60px;
  top: 25px;
  border-radius: 15px;
  transition:
    top 0.5s ease,
    width 0.5s ease,
    max-width 0.5s ease,
    border-radius 0.5s ease,
    background 0.3s ease,
    border-color 0.3s ease,
    box-shadow 0.3s ease;
  animation: scale-in 0.5s ease;
  box-shadow: var(--shadow-nav);

  &:not(.is-sticky):hover {
    background: var(--bg-nav-hover);
  }

  &.is-sticky {
    border-radius: 0 0 15px 15px;
    top: 0;
    max-width: 100%;
    width: 100%;
  }
}

@media (width <= 768px) {
  .header-container {
    max-width: 100%;
    width: 100%;
    top: 0;
    border-radius: 0;

    &.is-sticky {
      border-radius: 0;
    }
  }
}
</style>
