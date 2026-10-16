<script setup lang="ts">
interface Props {
  toc: TocItem[];
  activeId: string;
}

const { toc } = defineProps<Props>();

const handleTocClick = (id: string) => {
  const target = document.getElementById(id);

  if (!target) return;

  target.scrollIntoView({
    behavior: "smooth",
    block: "start",
  });

  history.pushState(null, "", `#${id}`);
};
</script>

<template>
  <ul class="toc">
    <li v-for="item in toc" :key="item.id" :class="{ active: activeId === item.id }">
      <a :href="`#${item.id}`" @click.prevent="handleTocClick(item.id)">{{ item.text }}</a>
    </li>
  </ul>
</template>

<style lang="less" scoped>
.toc {
  max-height: min(500px, 60vh);
  width: 100%;
  list-style: none;
  padding: 0;
  margin: 0;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--brand-primary) transparent;

  &::-webkit-scrollbar {
    width: 4px;
  }

  &::-webkit-scrollbar-thumb {
    background-color: var(--brand-primary);
    border-radius: 2px;
  }

  &::-webkit-scrollbar-track {
    background-color: transparent;
  }

  li {
    position: relative;
    padding: 8px 12px;
    margin: 4px 0;
    border-radius: 6px;
    transition:
      background-color 0.2s ease,
      border-color 0.2s ease;
    border-left: 2px solid transparent;

    a {
      display: block;
      color: var(--text-primary);
      text-decoration: none;
      font-size: 1.4rem;
      line-height: 1.4;
      transition: color 0.2s ease;

      &:hover {
        color: var(--text-primary);
      }
    }

    &.active {
      background-color: var(--brand-bg);
      border-left-color: var(--brand-primary);

      a {
        color: var(--brand-primary);
        font-weight: 500;
      }
    }

    &:hover {
      background-color: var(--bg-interactive-hover);
    }

    &:active {
      background-color: var(--bg-interactive-active);
    }
  }
}
</style>
