<script setup lang="ts">
import { useDebounceFn } from "@vueuse/core";
import type { FloatActionItem } from "./types";
import ActionButton from "./ActionButton.vue";
import PanelButton from "./PanelButton.vue";

const { width = 40 } = defineProps<{
  width?: number | string;
  items: FloatActionItem[];
}>();

const show = ref(false);

const SCROLL_SHOW_THRESHOLD = 80;
const SCROLL_HIDE_THRESHOLD = 20;

/**
 * 滚动事件处理函数，控制 `show` 状态
 */
const handleScroll = () => {
  const scrollTop = document.documentElement.scrollTop;

  if (scrollTop > SCROLL_SHOW_THRESHOLD) {
    show.value = true;
  } else if (scrollTop < SCROLL_HIDE_THRESHOLD) {
    show.value = false;
  }
};

const debouncedHandleScroll = useDebounceFn(handleScroll, 50); // 50ms 防抖间隔

onMounted(() => {
  nextTick(() => {
    handleScroll();
    window.addEventListener("scroll", debouncedHandleScroll);
  });
});

onUnmounted(() => {
  window.removeEventListener("scroll", debouncedHandleScroll);
});
</script>

<template>
  <Transition name="float-action-bar">
    <div v-if="show" class="float-action-bar" :style="{ width: `${width}px` }">
      <template v-for="item in items" :key="item.id">
        <ActionButton
          v-if="item.type === 'button'"
          :icon="item.icon"
          :title="$ts(item.label)"
          @click="item.onClick"
        />
        <PanelButton
          v-else-if="item.type === 'panel'"
          :icon="item.icon"
          :title="$ts(item.label)"
          :component="item.component"
        />
      </template>
    </div>
  </Transition>
</template>

<style lang="less" scoped>
.float-action-bar {
  position: fixed;
  bottom: 24px;
  right: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  z-index: 999;
}

.float-action-bar-enter-active,
.float-action-bar-leave-active {
  transition:
    opacity 0.5s ease-in-out,
    transform 0.5s ease-in-out;
}

.float-action-bar-enter-from,
.float-action-bar-leave-to {
  opacity: 0;
  transform: scale(0.3) translateX(100%);
}
</style>
