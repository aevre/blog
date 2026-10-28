<script lang="ts" setup>
import { useEditor, EditorContent } from "@tiptap/vue-3";
import StarterKit from "@tiptap/starter-kit";
import { Markdown } from "@tiptap/markdown";
import { watch } from "vue";

const content = defineModel<string>("content", { default: "" });

const editor = useEditor({
  content: content.value,
  contentType: "markdown",
  extensions: [StarterKit, Markdown],
  onUpdate: ({ editor }) => {
    content.value = editor.getMarkdown();
  },
  autofocus: true,
});

watch(content, (value) => {
  if (!editor.value) {
    return;
  }

  if (value === editor.value.getMarkdown()) {
    return;
  }

  editor.value.commands.setContent(value, {
    contentType: "markdown",
    emitUpdate: false,
  });
});
</script>

<template>
  <EditorContent :editor="editor" class="editor" />
</template>

<style lang="less" scoped>
.editor {
  width: 100%;
  height: 100%;
}

:deep(.tiptap) {
  outline: none;

  em {
    font-style: italic;
  }

  p,
  blockquote,
  ul,
  ol,
  dl,
  table {
    margin: 1.2rem 0;
  }

  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    scroll-margin-top: var(--header-height);
    font-weight: 700;
    color: var(--color-header);
    line-height: 1.4;
    margin-top: 2.4rem;
    margin-bottom: 1rem;
    position: relative;
  }

  h1 {
    font-size: 3.2rem;
    padding-bottom: 0.6rem;

    &::after {
      content: "";
      position: absolute;
      bottom: 0;
      left: 0;
      width: 80px;
      height: 4px;
      background: linear-gradient(90deg, var(--brand-primary), transparent);
      border-radius: 2px;
    }
  }

  h2 {
    font-size: 2.6rem;
    padding-bottom: 0.4rem;

    &::after {
      content: "";
      position: absolute;
      bottom: 0;
      left: 0;
      width: 60px;
      height: 3px;
      background: linear-gradient(90deg, var(--brand-primary), transparent);
      border-radius: 2px;
    }
  }

  h3 {
    font-size: 2.2rem;
  }

  h4 {
    font-size: 1.8rem;
    font-weight: 600;
  }

  h5 {
    font-size: 1.7rem;
    font-weight: 600;
  }

  h6 {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--text-secondary);
  }

  p {
    color: var(--text-primary);
    overflow-wrap: break-word;
    line-height: 1.8;
  }

  :not(pre.shiki) > code {
    display: inline;
    padding: 0.2em 0.45em;
    font-family: "Maple Mono", "Noto Sans SC", monospace;
    font-size: 0.86em;
    font-weight: 500;
    color: var(--inline-code-color);
    background-color: var(--inline-code-bg);
    border-radius: 4px;
    line-height: 1.45;
    white-space: break-spaces;
    overflow-wrap: break-word;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    border-radius: 8px;
    background: var(--bg-content);
    overflow: hidden;
    font-size: var(--font-size-table-row);
    color: var(--text-primary);
    box-shadow: var(--shadow-sm);
  }

  thead {
    background-color: var(--table-head-bg);
    color: var(--color-header);
    font-weight: bold;
    font-size: var(--font-size-table-header);
  }

  thead th {
    padding: 0.75em 1em;
    text-align: left;
    border-bottom: 2px solid var(--table-head-border);
  }

  tbody tr {
    border-bottom: 1px solid var(--border-table);
    font-size: var(--font-size-table-row);
    transition: background-color 0.2s ease;
  }

  tbody td {
    padding: 0.65em 1em;
  }

  tbody tr:nth-child(odd) {
    background-color: var(--table-row-even-bg);
  }

  tbody tr:hover {
    background-color: var(--table-row-hover-bg);
  }

  blockquote {
    border-radius: 8px;
    padding: 16px 20px;
    background: var(--bg-card-base);
    position: relative;
    border-left: none;
    margin: 1.5rem 0;
    box-shadow: var(--shadow-sm);

    &::before {
      display: block;
      position: absolute;
      content: "";
      width: 4px;
      left: 0;
      top: 0;
      height: 100%;
      background: linear-gradient(180deg, var(--brand-primary), var(--brand-hover));
      border-radius: 2px;
    }
  }

  ul,
  ol {
    margin: 1.2em 0;
    padding-left: 1.5em;
    color: var(--text-primary);
    font-size: 1.5rem;
    line-height: 1.75;
  }

  ul li,
  ol li {
    margin: 0.4rem 0;
    padding-left: 0.25rem;
    position: relative;
  }

  ul li::marker {
    color: var(--brand-primary);
  }

  ol li::marker {
    font-weight: bold;
    color: var(--brand-primary);
  }

  ul ul,
  ul ol,
  ol ul,
  ol ol {
    margin-top: 0.5em;
    margin-bottom: 0.5em;
    padding-left: 1.5em;
  }

  .contains-task-list {
    padding-left: 0;
    list-style: none;
  }

  mark {
    padding: 0 0.5rem;
    color: var(--text-on-brand);
    background: linear-gradient(135deg, var(--brand-primary), var(--brand-hover));
    border-radius: 4px;
  }

  u {
    text-decoration-line: underline;
    text-decoration-color: var(--brand-primary);
    text-underline-offset: 0.5rem;
    text-decoration-thickness: 2px;
  }

  a {
    color: var(--brand-primary);
    text-decoration: none;
    position: relative;
    transition: color 0.2s ease;

    &::after {
      content: "";
      position: absolute;
      bottom: -2px;
      left: 0;
      width: 0;
      height: 2px;
      background: var(--brand-primary);
      transition: width 0.2s ease;
    }

    &:hover::after {
      width: 100%;
    }
  }

  hr {
    border: none;
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--border-default), transparent);
    margin: 2rem 0;
  }
}
</style>
