import { markRaw } from "vue";
import { ContectGroup, ThemeControlCard } from "#components";
import type { ToolbarItem } from "./types";

export const globalItems = [
  {
    id: "back-top",
    type: "action",
    icon: "iconamoon:arrow-up-2-fill",
    label: "tooltip.backToTop",
    handler: () => {
      window.scrollTo({
        top: 0,
        behavior: "smooth",
      });
    },
  },
  {
    id: "rss",
    type: "action",
    icon: "ion:logo-rss",
    label: "tooltip.rss",
    handler: async ({ message }) => {
      await navigator.clipboard.writeText(window.location.origin + "/api/v1/rss");
      message.info("copy");
    },
  },
  {
    id: "contact",
    type: "panel",
    icon: "ion:mail",
    label: "tooltip.contact",
    component: markRaw(ContectGroup),
  },
  {
    id: "theme",
    type: "panel",
    icon: "fluent:apps-list-detail-24-filled",
    label: "tooltip.themeOption",
    component: markRaw(ThemeControlCard),
  },
] as const satisfies readonly ToolbarItem[];
