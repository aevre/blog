<script setup lang="ts">
import { ContactGroup, ThemeControlCard } from "#components";
import type { FloatActionItem } from "./FloatActionGroup/types";

const message = useMessage();

const { $ts } = useI18n();

const { $localePath } = useI18n();

const { floatActionItems, add } = useFloatActionStore();
const backTop = () => {
  window.scrollTo({
    top: 0,
    behavior: "smooth",
  });
};

const toUserSetting = () => {
  navigateTo($localePath("/admin"));
};

const copyRSSFeedUrl = async () => {
  await navigator.clipboard.writeText(window.location.origin + "/api/v1/rss");

  message.success($ts("message.rssCopied"), {
    keepAliveOnHover: true,
  });
};

const actions: FloatActionItem[] = [
  {
    id: "back-top",
    type: "button",
    icon: "iconamoon:arrow-up-2-fill",
    label: "tooltip.backToTop",
    onClick: backTop,
  },
  {
    id: "rss",
    type: "button",
    icon: "ion:logo-rss",
    label: "tooltip.rss",
    onClick: copyRSSFeedUrl,
  },
  {
    id: "contact",
    type: "panel",
    icon: "ion:mail",
    label: "tooltip.contact",
    component: markRaw(ContactGroup),
  },
  {
    id: "theme",
    type: "panel",
    icon: "fluent:apps-list-detail-24-filled",
    label: "tooltip.themeOption",
    component: markRaw(ThemeControlCard),
  },
  {
    id: "user-setting",
    type: "button",
    icon: "iconamoon:settings-fill",
    label: "tooltip.userSetting",
    onClick: toUserSetting,
  },
];

add(...actions);
</script>

<template>
  <FloatActionGroup :items="floatActionItems" />
</template>
