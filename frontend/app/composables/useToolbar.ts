import { onScopeDispose } from "vue";
import type { ToolbarConfig, ToolbarItem } from "~/components/FloatToolbar/types";
import type { GlobalToolbarItemId } from "~/plugins/toolbar";

type ItemId<T> = T extends { id: infer I extends string } ? I : never;

type ToolbarApi<Ids extends string> = {
  registry: ReturnType<typeof useNuxtApp>["$toolbar"]["registry"];
  items: ReturnType<typeof useNuxtApp>["$toolbar"]["items"];
  config: ReturnType<typeof useNuxtApp>["$toolbar"]["config"];
  setItems(ids: readonly Ids[]): void;
  setConfig(config: ToolbarConfig): void;
  register<const T extends readonly ToolbarItem[]>(
    definitions: T,
  ): ToolbarApi<Ids | ItemId<T[number]>>;
  register<const T extends ToolbarItem>(definition: T): ToolbarApi<Ids | ItemId<T>>;
};

export function useToolbar<const T extends readonly ToolbarItem[] = readonly []>(
  definitions?: T,
): ToolbarApi<GlobalToolbarItemId | ItemId<T[number]>> {
  const { $toolbar } = useNuxtApp();
  const unregisters = new Set<() => void>();

  const previousSelectedIds = [...$toolbar.items.value.map((i) => i.id)];

  function register<const R extends readonly ToolbarItem[]>(
    input: R | R[number],
  ): ToolbarApi<GlobalToolbarItemId | ItemId<T[number]> | ItemId<R[number]>> {
    const unregister = $toolbar.register(input as readonly ToolbarItem[]);
    let active = true;

    const dispose = () => {
      if (!active) return;
      active = false;
      unregister();
      unregisters.delete(dispose);
    };

    unregisters.add(dispose);

    return createToolbar<GlobalToolbarItemId | ItemId<T[number]> | ItemId<R[number]>>();
  }

  function createToolbar<
    Ids extends string = GlobalToolbarItemId | ItemId<T[number]>,
  >(): ToolbarApi<Ids> {
    return {
      registry: $toolbar.registry,
      items: $toolbar.items,
      config: $toolbar.config,
      setItems(ids) {
        $toolbar.setItems(ids);
      },
      setConfig(config) {
        $toolbar.setConfig(config);
      },
      register: register as ToolbarApi<Ids>["register"],
    };
  }

  if (definitions?.length) {
    const unregister = $toolbar.register(definitions);
    let active = true;

    const dispose = () => {
      if (!active) return;
      active = false;
      unregister();
      unregisters.delete(dispose);
    };

    unregisters.add(dispose);
  }

  onScopeDispose(() => {
    for (const unregister of [...unregisters]) {
      unregister();
    }

    $toolbar.setItems(previousSelectedIds);
  });

  return createToolbar();
}
