<script setup lang="ts">
import type { ApiPageResponse } from "~/types/api";
import type { Post } from "~/types/post";

const page = ref<number>(1);
const pageSize = ref<number>(10);

const { data } = await useFetch<ApiPageResponse<Post>>("/api/v1/posts", {
  method: "get",
  query: {
    page,
    pageSize,
  },
});

const posts = computed(() => data.value?.data.list ?? []);
const total = computed(() => data.value?.data.total ?? posts.value.length);

watch(page, () => {
  scrollTo({
    left: 0,
    top: window.innerHeight - 40,
    behavior: "smooth",
  } as ScrollToOptions);
});
</script>

<template>
  <div>
    <WelcomePanel />
    <ContentPanel :spacer="false" class="content-panel">
      <div class="post-list">
        <h1 v-if="posts.length > 0" class="title">Article</h1>
        <PostCard
          v-for="(post, index) in posts"
          :key="post.id"
          :index="index"
          :cover="post.cover"
          :url="'/blog/' + post.id"
          :title="post.title"
          :date="post.updatedAt"
          :summary="post.summary"
        />
      </div>
      <BasePagination v-model:page="page" :page-size="pageSize" :total="total">
        <template #default="{ items, select }">
          <nav class="pagination">
            <button
              v-for="item in items"
              :key="item.key"
              :disabled="item.disabled"
              :class="{
                active: item.active,
                'pagination-item': true,
              }"
              @click="select(item)"
            >
              {{ item.label }}
            </button>
          </nav>
        </template>
      </BasePagination>
    </ContentPanel>
  </div>
</template>

<style lang="less" scoped>
.content-panel {
  border-radius: 10px 10px 0 0;
  box-shadow: var(--shadow-md);
}

.title {
  position: relative;
  color: var(--text-primary);
  font-weight: 500;

  &::after {
    position: absolute;
    content: "";
    width: 80px;
    height: 6px;
    left: 0;
    bottom: 0;
    border-radius: 5px;
    background-color: var(--brand-primary);
  }
}

.post-list {
  max-width: 800px;
  margin: 0 auto;
  display: grid;
  gap: 10px;
  padding: 10px 20px;
  grid-template-columns: 1fr;
}

.pagination {
  margin: 0 auto;
  width: max-content;
}

.pagination-item {
  margin: 0 8px;
  background: var(--bg-card-base);
  min-width: 32px;
  min-height: 32px;
  border-radius: var(--radius-card);
  border: 1px solid var(--border-color-card);
  padding: 8px;
  color: var(--text-primary);
  box-shadow: var(--shadow-card-base);
  font-family: MapleMono, monospace;
  font-weight: bold;

  &:hover {
    background-color: var(--bg-card-hover);
  }

  &:active {
    background-color: var(--bg-card-active);
  }

  &.active {
    background-color: var(--brand-primary);
    color: var(--text-primary);
  }
}
</style>
