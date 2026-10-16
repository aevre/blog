<script setup lang="ts">
interface Overview {
  total: number;
  normal: number;
  abnormal: number;
  pending: number;
}

const { data } = await useFetch<{
  data: Overview;
}>("/api/v1/links/overview", {
  method: "get",
});

const overview = computed(() => data.value?.data);

const { $ts } = useI18n();
</script>

<template>
  <div class="links">
    <h2>总览</h2>
    <div class="overview">
      <div v-for="(value, key) of overview" :key="key" class="card">
        <div class="data">{{ value }}</div>
        <div class="label">{{ $ts("link.overview." + key) }}</div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.links {
  color: var(--text-primary);
}

.overview {
  display: grid;
  margin-top: 20px;
  padding: 0 20px;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.card {
  display: flex;
  flex-direction: column;
  justify-content: space-evenly;
  background: var(--bg-card-base);
  backdrop-filter: blur(var(--blur-surface));
  height: 140px;
  padding: 20px;
  border-radius: var(--radius-card);
  box-shadow: var(--shadow-card-base);
  transition:
    box-shadow 0.2s ease-in-out,
    scale 0.2s ease-in-out;

  .data {
    font-size: 3rem;
    font-weight: 700;
  }

  &:hover {
    scale: 1.03;
    box-shadow: var(--shadow-card-hover);
  }
}
</style>
