import MarkdownIt, { type Token } from "markdown-it";
import { mark } from "@mdit/plugin-mark";
import { sub } from "@mdit/plugin-sub";
import { sup } from "@mdit/plugin-sup";
import type { Component, VNodeChild } from "vue";
import { NuxtImg } from "#components";
import CodeWrapper from "~/components/CodeWrapper";
import type { BundledLanguage, SpecialLanguage } from "shiki";
import { tasklist } from "@mdit/plugin-tasklist";

const md = MarkdownIt({ html: false, linkify: true }).use(mark).use(sup).use(sub).use(tasklist);

export type KeyedVNode = VNode & { key: string };

interface ParseResult {
  content: KeyedVNode[];
  toc?: TocItem[];
}

export interface TocItem {
  id: string;
  level: number;
  text: string;
}

interface Options {
  toc?: boolean;
}

function slugifyHeading(text: string): string {
  return text
    .trim()
    .toLowerCase()
    .replace(/[`~!@#$%^&*()+=[\]{};:'"\\|,.<>/?，。；：“”‘’、】【、？！…—]/g, " ")
    .replace(/[^\w\u4e00-\u9fa5\s-]/g, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
}

function uniqueSlug(text: string, slugCounters: Map<string, number>) {
  const n = slugCounters.get(text) ?? 0;
  slugCounters.set(text, n + 1);
  return n === 0 ? text : `${text}-${n}`;
}

function ensureHeadingID(text: string, level: number, slugCounters: Map<string, number>) {
  const slug = slugifyHeading(text);
  if (slug) return uniqueSlug(slug, slugCounters);
  return uniqueSlug(`heading-${level}`, slugCounters);
}

function createTokensParser(toc: boolean) {
  const keyCounters = new Map<string, number>();
  const slugCounters = new Map<string, number>();

  const tocItems: TocItem[] = [];

  const getNextKey = (type: string): string => {
    const currentCount = keyCounters.get(type) ?? 0;
    keyCounters.set(type, currentCount + 1);
    return `${type}-${currentCount}`;
  };

  function tokensToVNode(tokens: Token[]): VNodeChild[] {
    interface Stack {
      tag: string;
      children: VNodeChild[];
      attrs: { [key: string]: string | number };
      key: string;
    }
    const stack: Stack[] = [];
    const result: VNodeChild[] = [];

    const pushToParent = (node: VNodeChild) => {
      const parent = stack.at(-1);
      if (parent) {
        parent.children.push(node);
        return;
      }
      result.push(node);
    };

    type TokenHandler = (token: Token) => void;

    const leafTokenHandlers: Record<string, TokenHandler> = {
      inline: (token) => {
        if (!token.children) return;
        const inlineVNodes = tokensToVNode(token.children);
        for (const node of inlineVNodes) {
          pushToParent(node);
        }
      },
      text: (token) => {
        if (!token.content) return;
        pushToParent(token.content);
      },
      code_inline: (token) => {
        if (!token.content) return;
        const key = getNextKey("code");
        pushToParent(h("code", { key }, token.content));
      },
      checkbox_input: (token) => {
        if (!token.attrs) return;
        const attrs = Object.fromEntries(token.attrs);
        pushToParent(h("input", { ...attrs }));
      },
      hr: () => {
        const key = getNextKey("hr");
        pushToParent(h("hr", { key }));
      },
      image: (token) => {
        if (!token.attrs) return;
        const key = getNextKey("img");
        const attrs = Object.fromEntries(token.attrs);
        pushToParent(
          h(NuxtImg as Component, {
            ...attrs,
            key,
            src: attrs.src ?? "",
            class: "img",
            loading: "lazy",
          }),
        );
      },
      fence: (token) => {
        let lang = token.info.trim() as BundledLanguage | SpecialLanguage;
        if (!highlighter.getLoadedLanguages().includes(lang)) {
          lang = "plaintext";
        }
        const code = token.content;
        const key = getNextKey("pre");
        pushToParent(h(CodeWrapper, { lang, code, key }));
      },
      softbreak: () => {
        const key = getNextKey("br");
        pushToParent(h("br", { key }));
      },
    };

    function getInlinePlaintext(inline: Token | undefined): string {
      if (!inline || inline.type !== "inline") return "";
      const parts: string[] = [];
      for (const c of inline.children ?? []) {
        if (c.type === "text") parts.push(c.content);
      }
      return parts.join("");
    }

    for (const [i, token] of tokens.entries()) {
      if (toc && token.type === "heading_open") {
        const level = Number(token.tag.slice(1) ?? 0);
        const next = tokens[i + 1];
        const text = next?.type === "inline" ? getInlinePlaintext(next) : "";

        if (level >= 1 && level <= 6) {
          const id = ensureHeadingID(text, level, slugCounters);
          token.attrs = token.attrs ?? [];
          if (!token.attrs.some((attr) => attr[0] === "id")) token.attrs.push(["id", id]);
          tocItems.push({ id, level, text });
        }
      }

      if (token.tag && token.type.endsWith("_open")) {
        const attrs = Object.fromEntries(token.attrs ?? []);
        // 将开放标签推入栈中
        const key = getNextKey(token.tag);
        stack.push({ tag: token.tag, children: [], attrs, key });
        continue;
      }

      if (token.type.endsWith("_close")) {
        // 封闭标签，从栈中弹出并创建 VNode
        const top = stack.pop();
        if (!top) continue;
        const { tag, children, attrs, key } = top;
        const vnode = h(tag, { ...attrs, key }, children);
        pushToParent(vnode);
        continue;
      }

      const handler = leafTokenHandlers[token.type];
      if (handler) {
        handler(token);
        continue;
      }

      console.warn(`未处理的 token 类型: ${token.type}`);
    }

    return result;
  }

  return { tokensToVNode, tocItems };
}

const CHCHE_MAX_SIZE = 50;
const parseCache = new Map<string, Token[]>();

function hashString(str: string): string {
  let hash = 5381;
  for (let i = 0; i < str.length; i++) {
    hash = (hash * 33) ^ str.charCodeAt(i);
  }
  return (hash >>> 0).toString(36);
}

function getCacheKey(markdown: string, toc: boolean): string {
  return `${toc ? "1" : "0"}:${markdown.length}:${hashString(markdown)}`;
}

function setCache(key: string, value: Token[]) {
  if (parseCache.has(key)) parseCache.delete(key);
  parseCache.set(key, value);
  if (parseCache.size > CHCHE_MAX_SIZE) {
    const oldest = parseCache.keys().next().value;
    if (oldest !== undefined) parseCache.delete(oldest);
  }
}

export function parseMarkdownToVNode(markdown: string, options?: Options): ParseResult {
  const wantToc = options?.toc ?? false;

  const cacheKey = getCacheKey(markdown, wantToc);
  let tokens = parseCache.get(cacheKey);

  if (!tokens) {
    tokens = md.parse(markdown, {});
    setCache(cacheKey, tokens);
  }

  const { tokensToVNode, tocItems } = createTokensParser(wantToc);
  const content = tokensToVNode(tokens);

  if (import.meta.dev) {
    content.forEach((node, i) => {
      const hasKey =
        typeof node === "object" && node !== null && "key" in node && (node as VNode).key != null;
      if (!hasKey) {
        console.warn(`[parseMarkdownToVNode] 顶层第 ${i} 项缺少 key`);
      }
    });
  }

  const result: ParseResult = { content: content as KeyedVNode[], toc: tocItems };
  return result;
}
