<script setup lang="ts">
const { menuItems } = useAdminMenu();

const { open, collapsed } = useSidebar();
const { isMobile } = useResponsive();

const handleToggle = () => {
  if (isMobile.value) {
    collapsed.value = false;
    open.value = !open.value;
  } else {
    collapsed.value = !collapsed.value;
  }
};
</script>

<template>
  <div class="admin-layout">
    <BaseSidebar ref="sidebar" v-model:open="open" v-model:collapsed="collapsed" :items="menuItems">
      <template #header>
        <div class="logo">
          <img src="/images/logo.png" class="logo-image" />
        </div>
      </template>
    </BaseSidebar>
    <div class="right">
      <header class="header">
        <button class="menu-btn" @click="handleToggle">
          <Icon v-if="!isMobile && !collapsed" name="icon-park-outline:menu-fold" size="28" />
          <Icon v-else-if="!isMobile && collapsed" name="icon-park-outline:menu-unfold" size="28" />
          <Icon v-else name="hugeicons:menu-11" size="24" class="icon" />
        </button>
      </header>
      <main class="main">
        <slot />
      </main>
    </div>
  </div>
</template>

<style lang="less" scoped>
.admin-layout {
  width: 100%;
  height: 100vh;
  display: flex;
  color: var(--text-primary);
}

.right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.logo {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  color: var(--text-on-brand);

  .logo-image {
    width: 100%;
    max-width: 140px;
  }
}

.header {
  position: relative;
  height: var(--header-height);
  flex-shrink: 0;
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-nav-base);
  backdrop-filter: blur(var(--blur-elevated));
  display: flex;
  align-items: center;
  padding: 0 10px;

  .menu-btn {
    position: relative;
    background: none;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 10px;
    color: var(--text-primary);
    border-radius: 8px;
    transition: background 0.2s ease;

    &:hover {
      background: var(--bg-interactive-hover);
    }
  }
}

.main {
  flex: 1;
  min-height: 0;
  padding: 10px;
  overflow-y: auto;
}
</style>
