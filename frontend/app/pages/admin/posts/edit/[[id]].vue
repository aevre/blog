<script setup lang="ts">
import { watch } from "vue";
import type { ApiResponse } from "~/types/api";
import type { PostAdminDetail } from "~/types/post";

definePageMeta({
  layout: false,
});

const route = useRoute();
const id = computed<number>(() => Number(route.params.id) || -1);

const { data } = await useApiFetch<ApiResponse<PostAdminDetail>>(`/admin/posts/${id.value}`);

// const { get } = useClientApi();

const content = ref<string>("");
const title = ref<string>("");

watch(
  () => data.value?.data,
  (post) => {
    title.value = post?.title ?? "";
    content.value = post?.content ?? "";
  },
  { immediate: true },
);

// onMounted(async () => {
//   const { data } = await api<ApiResponse<PostAdminDetail>>(`/admin/posts/${id.value}`);
//   content.value = data.content;
//   title.value = data.title;
// });

const charCount = computed(() => content.value.length);

const lineCount = computed(() => {
  const text = content.value ?? "";
  if (!text.trim()) return 0;
  const trimmed = text.replace(/^\n+|\n+$/g, "");
  const mergedEmptyline = trimmed.replace(/\n{3,}/g, "\n\n");
  return mergedEmptyline.split("\n").length;
});
</script>

<template>
  <div>
    <NuxtLayout name="admin-sub-layout">
      <template #actions>
        <div class="editor-header">
          <BaseInput v-model="title" class="title-input" placeholder="输入文章标题..." />
        </div>
      </template>
      <div class="editor-body">
        <Tiptap v-model:content="content" />
      </div>
      <div class="editor-footer">
        <div class="detail">
          <div>{{ `${charCount} 字 - ${lineCount} 行` }}</div>
        </div>
        <div class="actions">
          <button class="btn trash">存入草稿箱</button>
          <button class="btn publish">发布</button>
        </div>
      </div>
    </NuxtLayout>
  </div>
</template>

<style lang="less" scoped>
.editor-body {
  display: flex;
  width: 100%;
  height: calc(100% - 60px);
  min-height: 0;
  min-width: 0;
  background-color: var(--bg-elevated);
  backdrop-filter: blur(var(--blur-surface));
  border-radius: var(--radius-card);
  overflow-y: auto;
  padding: 20px;
}

.editor {
  width: 100%;
  height: 100%;
}

.editor-header {
  display: flex;
  width: 100%;
  height: 100%;
  justify-content: center;
  margin: 0 auto;
  max-width: 1200px;
}

.title-input {
  height: 100%;
  border: none;
  font-size: 2rem;
  width: 100%;

  &:focus-within {
    border: none;
    box-shadow: none;
  }
}

.editor-footer {
  width: 100%;
  height: 50px;
  margin-top: 10px;
  background-color: var(--bg-card-base);
  backdrop-filter: blur(var(--blur-surface));
  border: 1px solid var(--border-default);
  border-radius: 8px;
  box-shadow: var(--shadow-md);
  display: flex;
  padding: 0 8px;

  .detail {
    flex-grow: 1;
    line-height: 50px;
    font-size: 1.4rem;
  }

  .actions {
    flex-shrink: 0;
    height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 8px;
    gap: 5px;

    .btn {
      height: 100%;
      padding: 8px;
      line-height: 1;
      border-radius: 8px;
      color: var(--text-primary);
      background-color: transparent;
      box-shadow: var(--shadow-sm);
    }

    .trash {
      border: 1px solid var(--border-default);
      background-color: var(--bg-card-base);

      &:hover {
        background-color: var(--bg-card-hover);
      }

      &:active {
        background-color: var(--bg-card-active);
      }
    }

    .publish {
      background-color: var(--brand-primary);
      box-shadow: var(--shadow-sm);

      &:hover {
        background-color: var(--brand-hover);
      }

      &:active {
        background-color: var(--brand-active);
      }
    }
  }
}
</style>
