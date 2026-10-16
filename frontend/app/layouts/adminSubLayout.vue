<script setup lang="ts">
const router = useRouter();

const handleBack = () => {
  router.back();
};
</script>

<template>
  <div class="admin-sub-layout">
    <header class="header">
      <button class="back-btn" @click="handleBack">
        <Icon name="eva:arrow-back-fill" size="28" />
      </button>
      <slot name="actions" />
    </header>
    <main class="main">
      <slot />
    </main>
  </div>
</template>

<style lang="less" scoped>
.admin-sub-layout {
  width: 100%;
  height: 100dvh;
  color: var(--text-primary);
}

.header {
  height: var(--header-height);
  width: 100%;
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-nav-base);
  backdrop-filter: blur(var(--blur-elevated));
  z-index: 10;
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
  padding: 0 10px;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
}

.main {
  height: 100%;
  padding: 65px 15px 15px;
}

.back-btn {
  background: none;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 10px;
  color: inherit;
}
</style>
