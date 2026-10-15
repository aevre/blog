import { computed, shallowReactive, shallowRef } from "vue";
import type { ComputedRef } from "vue";
import { globalItems } from "~/components/FloatToolbar/globalItems";
import type { ToolbarItem, ToolbarConfig } from "~/components/FloatToolbar/types";

export type GlobalToolbarItemId = (typeof globalItems)[number]["id"];

export interface ToolbarManager {
  registry: ComputedRef<ToolbarItem[]>;
  items: ComputedRef<ToolbarItem[]>;
  config: ComputedRef<Required<ToolbarConfig>>;

  register(input: ToolbarItem | readonly ToolbarItem[]): () => void;
  setItems(ids: readonly string[]): void;
  setConfig(config: ToolbarConfig): void;
}

export default defineNuxtPlugin(() => {
  const defaultConfig: Required<ToolbarConfig> = {
    maxWidth: 480,
    gap: 8,
  };

  const registrations = shallowReactive(new Map<string, { owner: symbol; item: ToolbarItem }>());

  const selectedIds = shallowRef<readonly string[]>([]);
  const configState = shallowRef({ ...defaultConfig });

  const globalOwner = Symbol("global-toolbar");

  for (const item of globalItems) {
    if (registrations.has(item.id)) {
      throw new Error(`[toolbar] Duplicate global ID: "${item.id}"`);
    }

    registrations.set(item.id, {
      owner: globalOwner,
      item,
    });
  }

  const registry = computed(() => Array.from(registrations.values(), ({ item }) => item));

  const items = computed(() =>
    selectedIds.value.flatMap((id) => {
      const item = registrations.get(id)?.item;
      return item && item.visible !== false ? [item] : [];
    }),
  );

  function register(input: ToolbarItem | readonly ToolbarItem[]): () => void {
    const batch: readonly ToolbarItem[] = Array.isArray(input) ? input : [input as ToolbarItem];

    const ids = new Set<string>();

    for (const item of batch) {
      if (ids.has(item.id) || registrations.has(item.id)) {
        throw new Error(`[toolbar] Duplicate ID: "${item.id}"`);
      }

      ids.add(item.id);
    }

    const owner = Symbol("toolbar-registration");

    for (const item of batch) {
      registrations.set(item.id, { owner, item });
    }

    let active = true;

    return () => {
      if (!active) return;
      active = false;

      for (const item of batch) {
        if (registrations.get(item.id)?.owner === owner) {
          registrations.delete(item.id);
        }
      }
    };
  }

  // 页面菜单在运行时注册，因此这里应使用 string[]。
  function setItems(ids: readonly (GlobalToolbarItemId | string)[]) {
    if (new Set(ids).size !== ids.length) {
      throw new Error("[toolbar] Duplicate IDs in setItems()");
    }

    selectedIds.value = [...ids];
  }

  function setConfig(config: ToolbarConfig) {
    configState.value = {
      ...defaultConfig,
      ...config,
    };
  }

  const toolbar: ToolbarManager = {
    registry,
    items,
    config: computed(() => configState.value),
    register,
    setItems,
    setConfig,
  };

  return {
    provide: {
      toolbar,
    },
  };
});
