<script setup lang="ts">
interface ContactItem {
  url: string;
  icon: string;
  title: string;
}

const show = defineModel<boolean>("show", { default: false, required: true });

const { $ts } = useI18n();

const contactItems = computed<ContactItem[]>(() =>
  show.value
    ? [
        {
          url: "https://github.com/aevre",
          icon: "codicon:github",
          title: $ts("tooltip.github"),
        },
        {
          url: "https://qm.qq.com/cgi-bin/qm/qr?k=82gkG9i3QYouTHUsD9h4mk1UGFYZdlxT",
          icon: "streamline-logos:qq-logo-solid",
          title: $ts("tooltip.addQQFriend"),
        },
      ]
    : [],
);
</script>

<template>
  <TransitionGroup name="contact-list" tag="ul" class="contact-list">
    <li
      v-for="(item, index) in contactItems"
      :key="item.url"
      :title="item.title"
      class="contact-item"
      :style="{
        transitionDelay: `${index * 50}ms`,
      }"
    >
      <a :href="item.url" target="_blank">
        <Icon :name="item.icon" size="24" class="icon" />
      </a>
    </li>
  </TransitionGroup>
</template>

<style lang="less" scoped>
.contact-list {
  display: flex;
  max-width: 210px;
  list-style: none;
  width: max-content;
  min-height: 40px;
  margin-left: auto;
  align-items: center;
  gap: 8px;
  justify-content: end;
  flex-wrap: wrap;
}

.contact-list-enter-active,
.contact-list-leave-active {
  transition: all 0.3s ease;
}

.contact-list-enter-from,
.contact-list-leave-to {
  opacity: 0;
  transform-origin: right center;
  transform: translateX(10px) scale(0.9);
}

.contact-item {
  position: relative;
  width: 35px;
  height: 35px;
  flex-shrink: 0;
  background: var(--bg-nav-base);
  backdrop-filter: blur(var(--blur-elevated));
  border-radius: var(--radius-nav);
  border: 1px solid var(--border-color-nav);
  box-shadow: var(--shadow-nav);
  overflow: hidden;

  &:hover {
    background: var(--bg-nav-hover);
  }

  &:active {
    background: var(--bg-nav-active);
  }

  a {
    position: relative;
    width: 100%;
    height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
  }
}

.icon {
  color: var(--text-primary);
}
</style>
