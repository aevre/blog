<script setup lang="ts">
import type { Post } from "~/types/post";
import type { ApiResponse } from "~/types/api";
import { ArticleToc } from "#components";

const { isMobile } = useResponsive();

const route = useRoute();

const params = computed(() => route.params);

const { data } = await useFetch<ApiResponse<Post>>(() => `/api/v1/posts/${params.value.id}`, {
  method: "get",
});

const post = computed<Post>(() => {
  return (
    data.value?.data ?? {
      id: -1,
      title: "",
      content: "",
      summary: "",
      publishedAt: "",
      updatedAt: "",
      viewCount: 0,
      readTimeMinutes: 0,
      cover: "",
      tags: [],
      categories: [],
    }
  );
});

useHead(() => ({ title: post.value.title }));

const rendered = computed(() => {
  if (!post.value.content) {
    return {
      content: [],
      toc: [],
    };
  }
  return parseMarkdownToVNode(post.value.content, { toc: true });
});

const content = computed(() => rendered.value.content);
const toc = computed(() => rendered.value.toc ?? []);
const { activeId } = useActiveToc(toc);

const toolbar = useToolbar().register({
  id: "toc",
  type: "panel",
  label: "tooltip.toc",
  component: ArticleToc,
  icon: "mdi:table-of-contents",
  width: 400,
  props: { toc, activeId },
});

watch(
  isMobile,
  () => {
    toolbar.setItems(isMobile.value ? ["back-top", "toc", "theme"] : ["back-top", "theme"]);
  },
  { immediate: true },
);

const tocShow = ref(true);

const { onBeforeEnter, onEnter, onAfterEnter, onBeforeLeave, onLeave, onAfterLeave } =
  useCollapseTransition();
</script>

<template>
  <ContentPanel :spacer="false">
    <ArticleCover :src="post.cover" :title="post.title" />
    <article class="article">
      <main class="article-content">
        <MarkdownRenderer :content="content" />
      </main>
      <div class="article-toc">
        <div class="article-toc-header" @click="tocShow = !tocShow">
          <span class="title">目录</span>
          <Icon
            class="icon"
            :name="`${tocShow ? 'mingcute:up-line' : 'mingcute:down-line'}`"
            size="24"
          />
        </div>
        <Transition
          name="toc-collapse"
          @before-enter="onBeforeEnter"
          @enter="onEnter"
          @after-enter="onAfterEnter"
          @before-leave="onBeforeLeave"
          @leave="onLeave"
          @after-leave="onAfterLeave"
        >
          <div v-if="toc.length !== 0 && tocShow">
            <ArticleToc :toc="toc" :active-id="activeId" />
          </div>
        </Transition>
      </div>
    </article>
  </ContentPanel>
</template>

<style lang="less" scoped>
.article {
  position: relative;
  width: 100%;
  display: flex;
  justify-content: center;
  align-self: flex-start;
  gap: 24px;
  padding: 0 20px;
  padding-top: 30px;
}

.article-content {
  width: 100%;
  max-width: 800px;
  background-color: var(--bg-surface);
  backdrop-filter: blur(var(--blur-surface));
  animation: article-show 1s ease-in-out;
  box-shadow: var(--shadow-lg);
  padding: 40px;
  margin-top: 8px;
  border: 1px solid var(--border-color-card);
  border-radius: 12px;
  position: relative;
  overflow: hidden;
}

.article-toc {
  position: sticky;
  top: 80px;
  width: 240px;
  height: max-content;
  background-color: var(--bg-elevated);
  backdrop-filter: blur(var(--blur-elevated));
  box-shadow: var(--shadow-md);
  padding: 16px;
  margin-top: 8px;
  border: 1px solid var(--border-color-card);
  border-radius: 12px;
  animation: article-show 1s ease-in-out;
  transition:
    box-shadow 0.3s ease,
    transform 0.3s ease;

  &:hover {
    box-shadow: var(--shadow-lg);
    transform: translateY(-2px);
  }

  .article-toc-header {
    display: flex;
    cursor: pointer;

    .title {
      font-size: 2rem;
    }

    .icon {
      margin-left: auto;
    }
  }
}

.toc-collapse-enter-active,
.toc-collapse-leave-active {
  overflow: hidden;
  transition:
    height 0.3s ease-in-out,
    opacity 0.3s ease-in-out;
}

.toc-collapse-enter-from,
.toc-collapse-leave-to {
  opacity: 0;
}

.toc-collapse-enter-to,
.toc-collapse-leave-from {
  opacity: 1;
}

@media (width <= 1200px) {
  .article-toc {
    display: none;
  }

  .article {
    padding: 0;
  }

  .article-content {
    padding: 0 16px;
    margin-top: 0;
  }
}
</style>
