<script setup lang="ts">
import { onClickOutside } from "@vueuse/core";
import { navLinks } from "./navLinks";
import type { MenuItem } from "~/types/ui/navbar";

const { $ts } = useI18n();

const items = computed<MenuItem[]>(() => {
  return navLinks.map((item) => {
    return {
      key: item.to,
      icon: item.icon,
      label: $ts(item.labelKey),
      to: item.to,
    };
  });
});

const dropdownOpen = ref(false);

onMounted(() => {
  watch(
    dropdownOpen,
    () => {
      if (dropdownOpen.value) {
        document.body.style.overflow = "hidden";
      } else {
        document.body.style.overflow = "";
      }
    },
    {
      immediate: true,
    },
  );
});

onUnmounted(() => {
  document.body.style.overflow = "";
});

const menuBtn = useTemplateRef("menuBtn");
const dropdown = useTemplateRef("dropdown");

onClickOutside(dropdown, (event) => {
  const target = event.target as Node | null;
  if (menuBtn.value?.contains(target)) return;

  dropdownOpen.value = false;
});

const { onBeforeEnter, onEnter, onAfterEnter, onBeforeLeave, onLeave, onAfterLeave } =
  useCollapseTransition({ getHeight: () => `calc(100vh - 60px)` });
</script>

<template>
  <nav id="navbar" class="navbar">
    <BaseLogo />
    <button ref="menuBtn" class="menu-btn" @click="dropdownOpen = !dropdownOpen">
      <Icon :name="dropdownOpen ? 'material-symbols:close' : 'hugeicons:menu-11'" size="28" />
    </button>
    <Teleport to="body">
      <Transition
        name="dropdown"
        @before-enter="onBeforeEnter"
        @enter="onEnter"
        @after-enter="onAfterEnter"
        @before-leave="onBeforeLeave"
        @leave="onLeave"
        @after-leave="onAfterLeave"
      >
        <div v-if="dropdownOpen" id="dropdown" ref="dropdown" class="dropdown">
          <ul :class="{ 'nav-menu-mobile': true }">
            <li v-for="item in items" :key="item.to" class="nav-item" @click="dropdownOpen = false">
              <NuxtLinkLocale :to="item.to">
                <div class="icon">
                  <Icon :name="item.icon" size="18" />
                </div>
                <span>
                  {{ item.label }}
                </span>
              </NuxtLinkLocale>
            </li>
          </ul>
        </div>
      </Transition>
    </Teleport>
  </nav>
</template>

<style lang="less" scoped>
.navbar {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.menu-btn {
  position: relative;
  right: -15px;
  background-color: transparent;
  border: none;
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  width: 50px;
  color: var(--text-primary);
}

.dropdown {
  position: fixed;
  top: 60px;
  right: 0;
  width: 100%;
  height: calc(100vh - 60px);
  background: var(--bg-elevated);
  backdrop-filter: blur(var(--blur-elevated));
  border-top: 2px solid var(--border-color-nav);
  box-shadow: var(--shadow-md);
  padding-top: 20px;
  z-index: 9999;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition:
    height 0.3s ease-in-out,
    opacity 0.2s ease-in-out;
  overflow: hidden;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
}

.dropdown-enter-to,
.dropdown-leave-from {
  opacity: 1;
}

.nav-menu-mobile {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 300px;
  gap: 10px;
  height: 100%;
  align-items: center;
  list-style: none;
  margin: 0 auto;

  .nav-item {
    width: 100%;
    height: 40px;

    a {
      width: 100%;
      height: 100%;
      display: flex;
      color: var(--text-primary);
      position: relative;
      align-items: center;
      padding: 0 8px;
      border-radius: 10px 10px 0 0;
      border-bottom: 1px solid var(--border-default);
      transition:
        color 0.2s ease-in-out,
        border-bottom 0.2s ease-in-out,
        background-color 0.3s ease-in-out;

      &:active {
        color: var(--brand-active);
      }

      &:hover {
        border-bottom: 1px solid var(--brand-primary);
        color: var(--brand-primary);
        background: var(--bg-interactive-hover);
      }

      :deep(.svg-icon) {
        transition: color 0.2s ease-in-out;
      }

      &:hover:deep(.svg-icon) {
        color: var(--brand-primary);
      }

      .icon {
        height: 100%;
        display: flex;
        justify-content: center;
        align-items: center;
        margin-right: 10px;
      }
    }
  }
}
</style>
