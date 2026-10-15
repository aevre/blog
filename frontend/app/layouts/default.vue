<script setup lang="ts">
useHead({
  bodyAttrs: {
    class: "with-bg",
  },
});

const isSticky = ref(false);
useScrollHandler(100, (over) => (isSticky.value = over));

const toolbar = useToolbar([]).register({
  id: "test",
  type: "action",
  label: "123",
  icon: "mingcute:right-line",
  handler: () => {},
});

toolbar.setItems(["back-top", "test"]);
</script>

<template>
  <div class="default-layout">
    <ReadProgress />
    <SiteUtilityBar />
    <BaseHeader :is-sticky="isSticky" />
    <main class="main">
      <slot />
    </main>
    <BaseFooter />
  </div>
</template>

<style lang="less" scoped>
.default-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.main {
  margin: 0;
  flex: 1;
}
</style>
