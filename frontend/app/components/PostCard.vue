<script setup lang="ts">
interface Props {
  title: string;
  summary?: string;
  cover?: string;
  date?: string;
  url: string;
}

const { title, summary = "", cover = "", date = "1970-01-01", url } = defineProps<Props>();

const updatedAt = computed(() => {
  return `上次更新于 ${formatDate(date, "YYYY-MM-DD")}`;
});

const postCardRef = useTemplateRef<HTMLElement>("post-card");
const isVisible = ref(false);

useAddClassOnIntersect(postCardRef, "show");
</script>

<template>
  <div ref="post-card" class="post-card" :class="{ show: isVisible }">
    <NuxtLinkLocale :to="url" class="post-card-link">
      <div class="cover">
        <NuxtImg :src="cover" />
      </div>

      <div class="date">
        <Icon name="fluent:clock-12-regular" />
        <span>{{ updatedAt }}</span>
      </div>
      <div class="title" :title="title">{{ title }}</div>
      <div class="summary" :title="summary">{{ summary }}</div>
    </NuxtLinkLocale>
  </div>
</template>

<style lang="less" scoped>
.post-card {
  height: 260px;
  width: 100%;
  margin: 10px 0;
  opacity: 0;
  color: var(--text-primary);
  transform: translateY(50px);
  transition:
    opacity 0.5s ease-out,
    transform 0.5s ease-out;

  &.show {
    opacity: 1;
    transform: translateY(0);
  }

  a {
    color: inherit;
  }
}

.post-card-link {
  display: block;
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: 12px;
  overflow: hidden;
  background: var(--bg-card-base);
  backdrop-filter: blur(var(--blur-surface));
  box-shadow: var(--shadow-card-base);
  border: 1px solid var(--border-color-card);
  transition:
    color 0.5s ease,
    box-shadow 0.3s ease,
    border-color 0.3s ease;

  .cover {
    height: calc(100% - 65px);
    width: 100%;
    border-radius: 12px;
    overflow: hidden;
    background: var(--bg-surface-muted);
    transition:
      transform 0.5s ease,
      height 0.5s ease;

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      transition: transform 0.5s ease;
    }
  }

  .date,
  .title {
    position: absolute;
    padding: 4px 10px;
    background: var(--bg-card-title);
    backdrop-filter: blur(var(--blur-surface));
    border: 1px solid var(--border-glass);
    border-radius: 6px;
  }

  .date {
    top: 20px;
    left: 20px;
    font-size: 1.2rem;
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .title {
    bottom: 85px;
    left: 20px;
    width: max-content;
    max-width: 90%;
    font-size: 1.8rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    transition: bottom 0.5s ease;
  }

  .summary {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    line-clamp: 2;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100%;
    height: 55px;
    font-size: 1.5rem;
    line-height: 27.5px;
    letter-spacing: 2px;
    padding: 5px 10px;
  }

  &:hover {
    box-shadow: var(--shadow-card-hover);
    border-color: var(--border-color-card-hover);

    .cover {
      transform: scale(1.1);
      height: 100%;

      img {
        transform: scale(1.05);
      }
    }

    .title {
      bottom: 20px;
    }
  }
}

@media (width <= 768px) {
  .post-card-link {
    .title {
      font-size: 1.5rem;
      left: 10px;
      bottom: 75px;
      padding: 5px 8px;
    }

    .date {
      top: 10px;
      left: 10px;
      font-size: 1rem;
      padding: 3px 8px;
    }
  }
}
</style>
