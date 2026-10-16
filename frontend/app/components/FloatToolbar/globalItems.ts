import { ContactGroup, ThemeControlCard } from "#components";
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
    handler: async ({ notify }) => {
      await navigator.clipboard.writeText(window.location.origin + "/api/v1/rss");
      notify("message.rssCopied", { type: "info" });
    },
  },
  {
    id: "contact",
    type: "panel",
    icon: "ion:mail",
    label: "tooltip.contact",
    component: ContactGroup,
  },
  {
    id: "theme",
    type: "panel",
    icon: "fluent:apps-list-detail-24-filled",
    label: "tooltip.themeOption",
    component: ThemeControlCard,
    props: {
      show: true,
    },
  },
] as const satisfies readonly ToolbarItem[];
