import { computed, shallowReactive, shallowRef } from "vue";
import { globalItems } from "~/components/FloatToolbar/globalItems";
import type {
  ToolbarItem,
  ToolbarConfig,
  ToolbarManager,
  ToolbarSelection,
} from "~/components/FloatToolbar/types";

export type GlobalToolbarItemId = (typeof globalItems)[number]["id"];

type ToolbarOwner = symbol;

export default defineNuxtPlugin(() => {
  const defaultConfig: Required<ToolbarConfig> = {
    maxWidth: 480,
    iconSize: 19,
    gap: 8,
  };

  const registrations = shallowReactive(
    new Map<ToolbarItem["id"], { owner: symbol; item: ToolbarItem }>(),
  );

  const selectedIds = shallowRef<readonly ToolbarItem["id"][]>([]);
  let selectedOwner: ToolbarOwner | undefined;
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

    const ids = new Set<ToolbarItem["id"]>();

    for (const item of batch) {
      if (ids.has(item.id) || registrations.has(item.id)) {
        throw new Error(`[toolbar] Duplicate ID: "${String(item.id)}"`);
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

  function getSelection(): ToolbarSelection {
    return {
      ids: selectedIds.value,
      owner: selectedOwner,
    };
  }

  // 页面菜单在运行时注册，因此这里应使用 string[]。
  function setItems(
    ids: readonly (GlobalToolbarItemId | ToolbarItem["id"])[],
    owner?: ToolbarOwner,
  ) {
    if (new Set(ids).size !== ids.length) {
      throw new Error("[toolbar] Duplicate IDs in setItems()");
    }

    selectedIds.value = [...ids];
    selectedOwner = owner;
  }

  function restoreItems(selection: ToolbarSelection, owner: symbol) {
    if (selectedOwner !== owner) return;

    selectedIds.value = [...selection.ids];
    selectedOwner = selection.owner;
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
    getSelection,
    register,
    setItems,
    restoreItems,
    setConfig,
  };

  return {
    provide: {
      toolbar,
    },
  };
});
