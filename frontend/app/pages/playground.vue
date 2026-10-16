<script setup lang="ts">
definePageMeta({
  layout: false,
});

const createDummyPanel = (title: string) => ({
  setup() {
    return () =>
      h("div", { style: "padding: 10px;" }, [
        h("h3", title),
        h("p", `这是 ${title} 的面板内容区域。`),
      ]);
  },
});

const toolbar = useToolbar([
  {
    id: "item-1",
    label: "返回顶部",
    icon: "mingcute:up-line",
    type: "action",
    priority: 100,
    handler: () => window.scrollTo({ top: 0, behavior: "smooth" }),
  },
  {
    id: "item-3",
    label: "大纲目录",
    icon: "mingcute:toc-line",
    type: "panel",
    priority: 85,
    component: createDummyPanel("文章目录"),
  },
  {
    id: "item-4",
    label: "全屏阅读",
    icon: "mingcute:fullscreen-line",
    type: "action",
    priority: 80,
    handler: () => document.documentElement.requestFullscreen?.(),
  },
  {
    id: "item-5",
    label: "夜间模式",
    icon: "mingcute:moon-line",
    type: "action",
    priority: 75,
    handler: () => {
      document.documentElement.classList.toggle("dark");
    },
  },
  {
    id: "item-6",
    label: "文章检索",
    icon: "mingcute:search-line",
    type: "panel",
    priority: 70,
    component: createDummyPanel("快捷搜索"),
  },
  {
    id: "item-7",
    label: "分享页面",
    icon: "mingcute:share-line",
    type: "panel",
    priority: 65,
    component: createDummyPanel("分享面板"),
    width: 300,
  },
  {
    id: "item-8",
    label: "字体调整",
    icon: "mingcute:font-size-line",
    type: "panel",
    priority: 60,
    component: createDummyPanel("字体大小设置"),
  },
  {
    id: "item-9",
    label: "背景主题",
    icon: "mingcute:palette-line",
    type: "panel",
    priority: 55,
    component: createDummyPanel("主题配色"),
  },
  {
    id: "item-12",
    label: "评论留言",
    icon: "mingcute:message-3-line",
    type: "panel",
    priority: 40,
    component: createDummyPanel("快捷评论区"),
  },
  {
    id: "item-13",
    label: "历史足迹",
    icon: "mingcute:history-line",
    type: "panel",
    priority: 35,
    component: createDummyPanel("浏览历史"),
  },
  {
    id: "item-14",
    label: "数据统计",
    icon: "mingcute:bar-chart-line",
    type: "panel",
    priority: 30,
    component: createDummyPanel("阅读数据统计"),
  },
  {
    id: "item-16",
    label: "开发者模式",
    icon: "mingcute:code-line",
    type: "action",
    priority: 20,
    handler: () => console.log("Dev Mode"),
  },
  {
    id: "item-17",
    label: "意见反馈",
    icon: "mingcute:bug-line",
    type: "panel",
    priority: 15,
    component: createDummyPanel("反馈表单"),
  },
  {
    id: "item-18",
    label: "关于本站",
    icon: "mingcute:information-line",
    type: "panel",
    priority: 10,
    component: createDummyPanel("关于页面"),
  },
  {
    id: "item-19",
    label: "帮助中心",
    icon: "mingcute:question-line",
    type: "action",
    priority: 5,
    handler: () => alert("帮助中心"),
  },
  {
    id: "item-20",
    label: "系统设置",
    icon: "mingcute:settings-1-line",
    type: "panel",
    priority: 1,
    component: createDummyPanel("全局系统设置"),
    width: 360,
  },
] as const);

toolbar.setItems([
  "item-1",
  "item-3",
  "item-4",
  "item-5",
  "item-6",
  "item-7",
  "item-8",
  "item-9",
  "item-12",
  "item-13",
  "item-14",
  "item-16",
  "item-17",
  "item-18",
  "item-19",
  "item-20",
]);

const content = ref("");
const summary = ref("");
const loading = ref(false);

let eventSource: EventSource | null = null;

const closeSSE = () => {
  if (eventSource) {
    eventSource.close();
    eventSource = null;
  }
};

const summarize = async () => {
  if (!content.value.trim()) {
    return;
  }

  closeSSE();

  summary.value = "";
  loading.value = true;

  try {
    const data = await $fetch<{
      sessionId: string;
    }>("/api/v1/model/summarize", {
      method: "POST",
      body: {
        content: content.value,
      },
    });

    console.log("session:", data.sessionId);

    eventSource = new EventSource(`/api/v1/model/summarize/${data.sessionId}`);

    eventSource.onopen = () => {
      console.log("SSE connected");
    };

    eventSource.onmessage = (event) => {
      console.log("message:", event.data);
      summary.value += event.data;
    };

    eventSource.addEventListener("done", () => {
      console.log("SSE done");
      loading.value = false;
      closeSSE();
    });

    eventSource.onerror = (err) => {
      console.error("SSE error:", err);
      loading.value = false;
      closeSSE();
    };
  } catch (err) {
    console.error("summarize error:", err);
    loading.value = false;
  }
};

onBeforeUnmount(() => {
  closeSSE();
});
</script>

<template>
  <div class="container">
    <textarea v-model="content" class="content-input" placeholder="请输入需要摘要的内容" />

    <button class="submit-btn" :disabled="loading" @click="summarize">
      {{ loading ? "生成中..." : "生成摘要" }}
    </button>

    <div class="summary">
      {{ summary }}
    </div>
  </div>
</template>

<style scoped lang="less">
.container {
  width: 100%;
  min-height: 100vh;
  padding: 40px;
  color: var(--text-primary);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.content-input {
  width: 100%;
  min-height: 240px;
  padding: 12px;
  resize: vertical;
}

.submit-btn {
  width: 120px;
  height: 40px;
}

.summary {
  white-space: pre-wrap;
  text-align: left;
  padding: 16px;
  border: 1px solid var(--border-default);
  min-height: 120px;
}
</style>
