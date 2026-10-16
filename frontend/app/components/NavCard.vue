<script setup lang="ts">
import type { MenuItem } from "~/types/ui/navbar";

interface Props {
  menuItems: MenuItem[];
}

const { menuItems } = defineProps<Props>();

const route = useRoute();
</script>

<template>
  <aside class="aside">
    <ul>
      <li v-for="item in menuItems" :key="item.label" :class="{ active: item.to === route.path }">
        {{ item.label }}
      </li>
    </ul>
  </aside>
</template>

<style lang="less" scoped>
.aside {
  height: max-content;
  width: 200px;
  border-radius: 12px;
  box-shadow: var(--shadow-md);
  position: absolute;
  top: 20px;
  left: 10px;
  background: var(--bg-nav-base);
  backdrop-filter: blur(var(--blur-elevated));
  border: 1px solid var(--border-color-nav);
  color: var(--text-primary);
  overflow: hidden;

  ul {
    list-style: none;
    padding: 6px;

    li {
      padding: 10px 14px;
      border-radius: 8px;
      cursor: pointer;
      transition: background 0.2s ease;
      font-size: 1.5rem;

      &:hover {
        background: var(--bg-interactive-hover);
      }
    }

    .active {
      background: var(--brand-primary);
      color: var(--text-on-brand);
      font-weight: 600;

      &:hover {
        background: var(--brand-hover);
      }
    }
  }
}
</style>
