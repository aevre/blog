<script setup lang="ts">
import { onClickOutside } from "@vueuse/core";

interface Option {
  label: string;
  value: string | number;
}

const { options = [] } = defineProps<{
  options?: Option[];
}>();

let observer: ResizeObserver | null = null;

function updateDropDirection() {
  const trigger = selectRef.value;
  if (!trigger) return;

  const rect = trigger.getBoundingClientRect();
  const estimatedHeight = 150;
  dropUp.value = rect.bottom + estimatedHeight > window.innerHeight && rect.top > estimatedHeight;
}

const modelValue = defineModel<string | number>("value", { required: true });

const open = ref(false);
const highlightedIndex = ref(0);
const selectRef = useTemplateRef("selectRef");

const selectedLabel = computed(() => {
  const selected = options.find((opt) => opt.value === modelValue.value);
  return selected ? selected.label : "Select";
});

const dropUp = ref(false);
onMounted(() => {
  updateDropDirection();

  const el = selectRef.value;
  if (!el) return;

  // ResizeObserver 检测 trigger 本身尺寸变化
  observer = new ResizeObserver(updateDropDirection);
  observer.observe(el);

  // 监听 scroll 和 resize 事件
  window.addEventListener("scroll", updateDropDirection, true);
  window.addEventListener("resize", updateDropDirection);
});

onBeforeUnmount(() => {
  observer?.disconnect();
  window.removeEventListener("scroll", updateDropDirection, true);
  window.removeEventListener("resize", updateDropDirection);
});

function toggle() {
  open.value = !open.value;

  if (open.value) {
    highlightedIndex.value = options.findIndex((opt) => opt.value === modelValue.value);
    selectRef.value?.focus();
  }
}

function close() {
  open.value = false;
}

const handleSelect = (value: string | number) => {
  modelValue.value = value;
  close();
};

function onKeydown(e: KeyboardEvent) {
  if (!open.value) return;
  if (e.key === "ArrowDown") {
    e.preventDefault();
    highlightedIndex.value = (highlightedIndex.value + 1) % options.length;
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    highlightedIndex.value = (highlightedIndex.value - 1 + options.length) % options.length;
  } else if (e.key === "Enter") {
    e.preventDefault();
    handleSelect(options[highlightedIndex.value]!.value);
  } else if (e.key === "Escape") {
    e.preventDefault();
    close();
  }
}

onClickOutside(selectRef, close);
</script>

<template>
  <div ref="selectRef" tabindex="0" class="select" @keydown="onKeydown">
    <button class="select-toggle" @click="toggle">
      {{ selectedLabel }}
      <Icon :name="open ? 'mingcute:up-fill' : 'mingcute:down-fill'" size="18" />
    </button>
    <transition :name="dropUp ? 'fade-up' : 'fade-down'">
      <ul
        v-if="open"
        ref="selectMenuRef"
        class="select-menu"
        :class="{ 'position-up': dropUp }"
        role="listbox"
        :aria-activedescendant="`option-${highlightedIndex}`"
      >
        <li
          v-for="(option, index) in options"
          :id="`option-${index}`"
          :key="option.value"
          class="select-item"
          :class="{ highlighted: index === highlightedIndex }"
          role="option"
          tabindex="-1"
          @click="handleSelect(option.value)"
          @mouseenter="highlightedIndex = index"
        >
          {{ option.label }}
        </li>
      </ul>
    </transition>
  </div>
</template>

<style lang="less" scoped>
.select {
  position: relative;
  display: inline-block;
  user-select: none;
}

.select-toggle {
  cursor: pointer;
  padding: 6px 12px;
  border: 1px solid var(--border-color-select);
  background: var(--bg-select);
  color: var(--text-primary);
  border-radius: 4px;
  width: 100%;
  height: 32px;
  min-width: 100px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
  box-shadow: var(--shadow-md);
}

.select-menu {
  position: absolute;
  left: 0;
  top: 100%;
  margin-top: 4px;
  border: 1px solid var(--border-glass);
  background: var(--bg-elevated);
  backdrop-filter: blur(var(--blur-elevated));
  border-radius: 4px;
  min-width: 100%;
  max-height: 150px;
  overflow-y: auto;
  box-shadow: var(--shadow-md);
  z-index: 1000;
  padding: 0;
  list-style: none;
}

.position-up {
  top: auto;
  bottom: calc(100% + 10px); // 向上展开
}

.fade-down-enter-active,
.fade-down-leave-active {
  transition:
    opacity 0.3s ease-in-out,
    top 0.3s ease-in-out;
}

.fade-down-enter-from,
.fade-down-leave-to {
  opacity: 0;
  top: 50%;
}

.fade-up-enter-active,
.fade-up-leave-active {
  transition:
    opacity 0.3s ease-in-out,
    bottom 0.3s ease-in-out;
}

.fade-up-enter-from,
.fade-up-leave-to {
  opacity: 0;
  bottom: 50%;
}

.select-item {
  padding: 6px 12px;
  cursor: pointer;
  font-weight: 500;
  font-size: 1em;
  color: var(--text-primary);
  transition:
    background-color 0.2s ease-in-out,
    color 0.2s ease-in-out;
}

.select-item.highlighted,
.select-item:hover {
  background-color: var(--brand-primary);
  color: var(--text-on-brand);
}
</style>
