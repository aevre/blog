<script setup lang="ts">
import { ImageViewer } from "#components";
import type { Column } from "~/components/BaseTable.vue";
import type { ApiPageResponse } from "~/types/api";
import type { PostAdminMeta } from "~/types/post";

// const message = useMessage();

const searchQuery = ref("");
const currentPage = ref(1);
const pageSize = ref(20);

const { get } = useClientApi();

const { data, pending, refresh } = useAsyncData(
  "admin-posts",
  () => {
    return get<ApiPageResponse<PostAdminMeta>>("/admin/posts", {
      query: {
        page: currentPage.value,
        pageSize: pageSize.value,
        search: searchQuery.value,
      },
    });
  },
  {
    default: () => ({
      code: 0,
      msg: "",
      data: {
        list: [],
        total: 0,
        page: 1,
        pageSize: 20,
      },
    }),
    server: false,
    watch: [currentPage, pageSize, searchQuery],
  },
);

watch(pending, () => {
  console.log(pending.value);
});

const posts = computed(() => data.value?.data?.list ?? []);
const total = computed(() => data.value?.data?.total ?? 0);
const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

function onSearch() {
  currentPage.value = 1;
  refresh();
}

function changePage(page: number) {
  currentPage.value = page;
  refresh();
}

// async function deletePost(id: number) {
//   try {
//     await del(`/admin/posts/${id}`);
//     message.success("文章已删除");
//     refresh();
//   } catch {
//     message.error("删除失败");
//   }
// }

const columns: Column<PostAdminMeta>[] = [
  {
    key: "id",
    title: "id",
    width: "35px",
  },
  {
    key: "cover",
    title: "封面",
    width: "80px",
    render: (row, _) =>
      h(ImageViewer, {
        src: row.cover,
        style: "width:56px; height: 40px; object-fit: cover; border-radius: 4px; display: block",
      }),
  },
  {
    key: "title",
    title: "标题",
  },
  {
    key: "status",
    title: "状态",
    render: (row) => {
      let status: string = "归档中";
      if (row.status === "published") {
        status = "已发布";
      }
      if (row.status === "draft") {
        status = "草稿";
      }
      return h(
        "span",
        {
          class: `status-tag status-tag-${row.status}`,
        },
        status,
      );
    },
  },
  {
    key: "actions",
    title: "操作",
    render: (row) => {
      return h(
        "button",
        {
          onClick: () => {
            navigateTo(`/admin/posts/edit/${row.id}`);
          },
        },
        "编辑",
      );
    },
  },
];
</script>

<template>
  <div class="posts-page">
    <div class="sticky-area">
      <header class="page-header">
        <div class="header-left">
          <h1 class="page-title">文章管理</h1>
          <span class="post-count">共 {{ total }} 篇文章</span>
        </div>
        <NuxtLinkLocale :to="'/admin/posts/edit'" class="create-btn">
          <Icon name="mdi:plus" size="18" />
          新建文章
        </NuxtLinkLocale>
      </header>

      <div class="search-bar">
        <div class="search-input-wrapper">
          <Icon name="material-symbols:search" size="18" class="search-icon" />
          <input
            v-model="searchQuery"
            class="search-input"
            placeholder="搜索文章标题..."
            @input="onSearch"
          />
        </div>
      </div>
    </div>

    <div class="table-wrapper">
      <BaseTable :columns="columns" :data="posts" />
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <button class="page-btn" :disabled="currentPage <= 1" @click="changePage(currentPage - 1)">
        <Icon name="mingcute:left-line" size="16" />
      </button>
      <template v-for="page in totalPages" :key="page">
        <button
          v-if="page === 1 || page === totalPages || Math.abs(page - currentPage) <= 2"
          class="page-btn"
          :class="{ active: page === currentPage }"
          @click="changePage(page)"
        >
          {{ page }}
        </button>
        <span v-else-if="page === currentPage - 3 || page === currentPage + 3" class="page-ellipsis"
          >...</span
        >
      </template>
      <button
        class="page-btn"
        :disabled="currentPage >= totalPages"
        @click="changePage(currentPage + 1)"
      >
        <Icon name="mingcute:right-line" size="16" />
      </button>
    </div>
  </div>
</template>

<style lang="less" scoped>
.posts-page {
  padding: 0 24px;
  color: var(--text-primary);
}

.sticky-area {
  position: sticky;
  top: 0;
  z-index: 10;
  background: var(--bg-nav-base);
  backdrop-filter: blur(var(--blur-elevated));
  padding-top: 24px;
  padding-bottom: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .page-title {
    font-size: 2.4rem;
    font-weight: 600;
    margin: 0;
  }

  .post-count {
    font-size: 1.3rem;
    color: var(--text-secondary);
    background: var(--bg-card-base);
    padding: 2px 10px;
    border-radius: 6px;
  }

  .create-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 36px;
    padding: 0 16px;
    background: var(--brand-primary);
    color: var(--text-on-brand);
    border-radius: 8px;
    font-size: 1.4rem;
    font-weight: 500;
    text-decoration: none;
    transition: opacity 0.2s;

    &:hover {
      opacity: 0.9;
    }
  }
}

.search-bar {
  margin-bottom: 0;

  .search-input-wrapper {
    display: flex;
    align-items: center;
    gap: 8px;
    max-width: 360px;
    padding: 8px 12px;
    border: 1px solid var(--border-default);
    border-radius: 8px;
    background: var(--bg-card-base);
    transition: border-color 0.2s;

    &:focus-within {
      border-color: var(--brand-primary);
    }
  }

  .search-icon {
    flex-shrink: 0;
    color: var(--text-secondary);
  }

  .search-input {
    flex: 1;
    border: none;
    outline: none;
    background: transparent;
    color: var(--text-primary);
    font-size: 1.4rem;
    font-family: inherit;

    &::placeholder {
      color: var(--text-tertiary);
    }
  }
}

.table-wrapper {
  overflow: auto;
  max-height: calc(100vh - 240px);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-card);
  background: var(--bg-card-base);
  backdrop-filter: blur(var(--blur-surface));

  :deep(.table-header) {
    position: sticky;
    top: 0;
    z-index: 1;
  }
}

.post-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 1.4rem;

  th {
    text-align: left;
    padding: 12px 16px;
    font-weight: 600;
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border-default);
    font-size: 1.2rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
  }

  td {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-default);
    vertical-align: middle;
  }

  tr:last-child td {
    border-bottom: none;
  }

  tr:hover {
    background: var(--bg-interactive-hover);
  }
}

.col-title {
  min-width: 200px;

  .title-text {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 1;
    line-clamp: 1;
    overflow: hidden;
    font-weight: 500;
  }
}

.col-status {
  width: 80px;
}

.status-badge {
  display: inline-flex;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 1.2rem;
  font-weight: 500;

  &.published {
    background: var(--green-100);
    color: var(--green-600);
  }

  &.draft {
    background: var(--yellow-100);
    color: var(--yellow-600);
  }
}

.col-views {
  width: 60px;
  text-align: center;
  color: var(--text-secondary);
}

.col-date {
  width: 100px;
  color: var(--text-secondary);
  font-size: 1.3rem;
  white-space: nowrap;
}

.col-actions {
  width: 80px;
  text-align: right;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-secondary);
  background: transparent;
  transition:
    background 0.15s,
    color 0.15s;

  &:hover {
    background: var(--bg-interactive-hover);
  }

  &.edit-btn:hover {
    color: var(--brand-primary);
  }

  &.delete-btn:hover {
    color: var(--state-error);
  }
}

:deep(.status-tag) {
  padding: 2px 8px;
  display: inline-block;
  border-radius: var(--radius-card);
  background: var(--gray-500);
  color: var(--text-primary);
  font-size: 14px;

  &-published {
    background: var(--green-500);
  }

  &-draft {
    background: var(--yellow-500);
  }
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  position: sticky;
  bottom: 0;
  z-index: 10;
  background: var(--bg-nav-base);
  backdrop-filter: blur(var(--blur-elevated));
  padding: 16px 0 24px;

  .page-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 32px;
    height: 32px;
    padding: 0 8px;
    border: 1px solid var(--border-default);
    border-radius: 6px;
    background: var(--bg-card-base);
    color: var(--text-primary);
    font-size: 1.3rem;
    cursor: pointer;
    transition:
      border-color 0.2s,
      background 0.2s;

    &:hover {
      border-color: var(--brand-primary);
    }

    &.active {
      background: var(--brand-primary);
      color: var(--text-on-brand);
      border-color: var(--brand-primary);
    }

    &:disabled {
      opacity: 0.4;
      cursor: not-allowed;
    }
  }

  .page-ellipsis {
    color: var(--text-tertiary);
    padding: 0 4px;
  }
}
</style>
