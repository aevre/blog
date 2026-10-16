<script setup lang="ts">
const markdownRaw = defineModel<string>("content", {
  default: "",
});

const editorRef = useTemplateRef<HTMLDivElement>("editor");

const { isMobile } = useResponsive();

type Mode = "preview" | "edit" | "both";
const mode = ref<Mode>("both");
const userTouchedMode = ref(false);

const rendered = computed(() => {
  console.log("render markdown", markdownRaw.value);
  if (!markdownRaw.value) {
    return {
      content: [],
      toc: [],
    };
  }
  return parseMarkdownToVNode(markdownRaw.value, { toc: true });
});

const content = computed(() => rendered.value.content);

watch(
  isMobile,
  (m) => {
    if (!userTouchedMode.value) {
      mode.value = m ? "edit" : "both";
    }
  },
  { immediate: true },
);

const isFullscreen = ref(false);

function syncFullscreenState() {
  isFullscreen.value = !!document.fullscreenElement;
}

onMounted(() => {
  document.addEventListener("fullscreenchange", syncFullscreenState);
  syncFullscreenState();
});

onBeforeUnmount(() => {
  document.removeEventListener("fullscreenchange", syncFullscreenState);
});

async function toggleFullscreen() {
  try {
    if (document.fullscreenElement) {
      await document.exitFullscreen();
    } else {
      await editorRef.value?.requestFullscreen();
    }
  } catch (e) {
    console.warn(e);
  }
}

const editRatio = ref(0);
const previewRatio = ref(0);

function updateEditRatio() {
  const edit = textareaRef.value;
  if (!edit) return;

  editRatio.value = edit.scrollTop / Math.max(edit.scrollHeight - edit.clientHeight, 1);
  console.log(editRatio.value);
}

function updatePreviewRatio() {
  const preview = previewRef.value;
  if (!preview) return;

  previewRatio.value = preview.scrollTop / Math.max(preview.scrollHeight - preview.clientHeight, 1);
}

function toggleMode() {
  userTouchedMode.value = true;

  if (isMobile.value) {
    mode.value = mode.value === "edit" ? "preview" : "edit";
    return;
  }

  if (mode.value === "both") mode.value = "edit";
  else if (mode.value === "edit") mode.value = "preview";
  else mode.value = "both";
}

const textareaRef = useTemplateRef<HTMLTextAreaElement>("textarea");
const previewRef = useTemplateRef<HTMLDivElement>("preview");

let syncing = false;

function syncToPreview() {
  if (syncing) return;
  syncing = true;

  const edit = textareaRef.value;
  const preview = previewRef.value;

  if (!edit || !preview) return;

  updateEditRatio();

  requestAnimationFrame(() => {
    preview.scrollTop = editRatio.value * (preview.scrollHeight - preview.clientHeight);

    syncing = false;
  });
}

function syncToEdit() {
  if (syncing) return;
  syncing = true;

  const edit = textareaRef.value;
  const preview = previewRef.value;

  if (!edit || !preview) return;

  updatePreviewRatio();

  requestAnimationFrame(() => {
    edit.scrollTop = previewRatio.value * (edit.scrollHeight - edit.clientHeight);

    syncing = false;
  });
}

watch(mode, async () => {
  await nextTick();

  const edit = textareaRef.value;
  const preview = textareaRef.value;

  if (edit) {
    edit.scrollTop = previewRatio.value * (edit.scrollHeight - edit.clientHeight);
  }

  if (preview) {
    preview.scrollTop = previewRatio.value * (preview.scrollHeight - preview.clientHeight);
  }
});
</script>

<template>
  <div ref="editor" class="article-edit">
    <div class="toolbar">
      <div class="tools-left"></div>
      <div class="tools-right">
        <button class="btn" @click="toggleMode">
          <Icon v-if="mode === 'edit'" name="mingcute:eye-line" size="18" />
          <Icon v-else-if="mode === 'preview'" name="mingcute:layout-grid-line" size="18" />
          <Icon v-else name="mingcute:edit-2-line" size="18" />
        </button>
        <button class="btn" @click="toggleFullscreen">
          <Icon v-if="!isFullscreen" name="mingcute:fullscreen-fill" size="18" />
          <Icon v-else name="mingcute:fullscreen-exit-fill" size="18" />
        </button>
      </div>
    </div>
    <div class="main">
      <div v-if="mode === 'edit' || mode === 'both'" class="edit">
        <textarea
          ref="textarea"
          v-model="markdownRaw"
          :class="{ 'split-border': mode === 'both' }"
          placeholder="输入文章内容"
          @scroll="syncToPreview"
        ></textarea>
      </div>
      <div
        v-if="mode === 'preview' || mode === 'both'"
        ref="preview"
        class="preview"
        @scroll="syncToEdit"
      >
        <MarkdownRenderer :content="content" />
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.article-edit {
  border: 1.5px solid var(--border-default);
  border-radius: 8px;
  box-shadow: var(--shadow-md);
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-elevated);
  backdrop-filter: blur(var(--blur-surface));
}

.toolbar {
  height: 40px;
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  padding: 5px 15px;
  border-bottom: 1px solid var(--border-default);

  .tools-left {
    display: flex;
    margin-right: auto;
    min-width: 0;
  }

  .tools-right {
    display: flex;
    margin-left: auto;
    min-width: 0;
  }

  .btn {
    display: inline-flex;
    justify-content: center;
    align-items: center;
    color: var(--text-primary);
    background-color: transparent;
    padding: 6px;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    user-select: none;

    &:hover {
      background-color: var(--bg-interactive-hover);
    }

    &:active {
      background-color: var(--bg-interactive-active);
      transform: scale(0.97);
    }
  }
}

.main {
  width: 100%;
  flex: 1;
  min-height: 0;
  display: flex;
  padding: 8px 0;
  gap: 0;
}

.edit,
.preview {
  flex: 1;
  min-width: 0;
  height: 100%;
}

.edit {
  display: flex;
}

.edit textarea {
  flex: 1;
  width: 100%;
  height: 100%;
  color: var(--text-primary);
  padding: 10px 20px;
  resize: none;
  font-size: 1.6rem;
  line-height: 1.5;
  outline: none;
  border: none;
  border-radius: 8px;
  background-color: transparent;
  min-width: 0;
  overflow: auto;
  scrollbar-width: none;
}

.edit textarea.split-border {
  border-right: 1px solid var(--border-default);
  border-radius: 8px 0 0 8px;
}

.preview {
  padding: 10px 14px;
  overflow: auto;
}

.preview.split {
  border-radius: 0 8px 8px 0;
}
</style>
