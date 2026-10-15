<script setup lang="ts">
const { isVisible } = useLoadingBar();
</script>

<template>
  <div>
    <NuxtRouteAnnouncer />
    <LoadingBar :is-visible="isVisible" />
    <MessageProvider :max="5">
      <NuxtLayout>
        <NuxtPage />
        <ToolbarManager />
      </NuxtLayout>
    </MessageProvider>
  </div>
</template>

<style lang="less">
.page-enter-active {
  transition:
    opacity 0.25s ease-out,
    transform 0.25s ease-out;
}

.page-leave-active {
  transition: opacity 0.2s ease-in;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(6px);
}

.page-leave-to {
  opacity: 0;
}

.layout-enter-active {
  transition: opacity 0.2s ease-out;
}

.layout-leave-active {
  transition: opacity 0.15s ease-in;
}

.layout-enter-from {
  opacity: 0;
}

.layout-leave-to {
  opacity: 0;
}
</style>
