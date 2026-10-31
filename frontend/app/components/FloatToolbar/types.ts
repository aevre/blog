import type { Component } from "vue";

export interface ToolbarItemBase {
  id: string | symbol;
  label: string;
  icon: string;
  priority?: number;
  visible?: boolean;
}

export interface ToolbarContext {
  notify: (message: string, options?: { type?: string }) => void;
}

export interface ToolbarAction extends ToolbarItemBase {
  type: "action";
  handler: (context: ToolbarContext) => void | Promise<void>;
}

export interface ToolbarPanel extends ToolbarItemBase {
  type: "panel";
  component: Component;
  props?: Record<string, unknown>;
  width?: number;
}

export type ToolbarItem = ToolbarAction | ToolbarPanel;

export interface ToolbarSelection {
  ids: readonly ToolbarItem["id"][];
  owner?: symbol;
}

export interface ToolbarConfig {
  /** 工具栏最大宽度 */
  maxWidth?: number;
  iconSize?: number;

  /** 菜单项之间的间距 */
  gap?: number;
}

export interface ToolbarManager {
  /** 所有已注册的菜单项 */
  registry: ComputedRef<ToolbarItem[]>;

  /** 当前页面实际显示的菜单项 */
  items: ComputedRef<ToolbarItem[]>;

  /** 当前布局配置 */
  config: ComputedRef<ToolbarConfig>;

  getSelection(): ToolbarSelection;

  /** 注册菜单项，返回注销函数 */
  register(items: ToolbarItem | readonly ToolbarItem[]): () => void;

  /** 设置当前页面显示的菜单 ID 列表 */
  setItems(ids: readonly ToolbarItem["id"][], owner?: symbol): void;

  restoreItems(selection: ToolbarSelection, owner: symbol): void;

  /** 覆盖当前布局配置 */
  setConfig(config: ToolbarConfig): void;
}
