<script setup lang="ts">
interface Props {
  toc: TocItem[];
  activeId: string;
}

const { toc } = defineProps<Props>();

const handleTocClick = (id: string) => {
  const target = document.getElementById(id);
  if (!target) return;

  target.scrollIntoView({ behavior: "smooth", block: "start" });
  history.pushState(null, "", `#${id}`);
};
</script>

<template>
  <nav class="toc-nav" aria-label="文章目录">
    <ul class="toc-list">
      <li
        v-for="item in toc"
        :key="item.id"
        class="toc-item"
        :class="{ active: activeId === item.id, 'toc-item-root': item.level <= 2 }"
        :style="{ '--toc-depth': Math.min(3, Math.max(0, item.level - 2)) }"
      >
        <a
          class="toc-link"
          :href="`#${item.id}`"
          :aria-current="activeId === item.id ? 'location' : undefined"
          @click.prevent="handleTocClick(item.id)"
        >
          {{ item.text }}
        </a>
      </li>
    </ul>
  </nav>
</template>

<style lang="less" scoped>
.toc-nav {
  width: 100%;
  max-height: min(500px, 55vh);
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--border-glass) transparent;

  &::-webkit-scrollbar {
    width: 4px;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--border-glass);
    border-radius: 4px;
  }
}

.toc-list {
  display: grid;
  gap: 2px;
  list-style: none;
  margin: 0;
  padding: 0;
}

.toc-item {
  position: relative;
  min-width: 0;
  padding-left: calc(10px + var(--toc-depth) * 12px);
  border-radius: 5px;
  transition: background-color 120ms ease;

  &::before {
    position: absolute;
    top: 7px;
    bottom: 7px;
    left: 0;
    width: 2px;
    border-radius: 2px;
    background: transparent;
    content: "";
    transition: background-color 120ms ease;
  }

  &:hover {
    background: var(--bg-interactive-hover);
  }

  &.active {
    background: var(--brand-bg-muted);

    &::before {
      background: var(--brand-primary);
    }
  }
}

.toc-link {
  display: block;
  overflow: hidden;
  padding: 7px 9px 7px 0;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 400;
  line-height: 1.45;
  text-decoration: none;
  text-overflow: ellipsis;
  overflow-wrap: anywhere;
  transition: color 120ms ease;

  &:hover {
    color: var(--text-primary);
  }

  .toc-item-root & {
    font-weight: 500;
  }

  .toc-item.active &,
  .toc-item.active &:hover {
    color: var(--brand-primary);
    font-weight: 600;
  }
}
</style>
