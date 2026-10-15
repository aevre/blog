<script setup lang="ts">
import DesktopNav from "./DesktopNav.vue";
import MobileNav from "./MobileNav.vue";

const { isMobile } = useResponsive();
</script>

<template>
  <component :is="isMobile ? MobileNav : DesktopNav" />
</template>
