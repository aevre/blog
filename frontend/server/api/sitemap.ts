import { defineEventHandler } from "h3";

export default defineEventHandler(async () => {
  try {
    // 调用后端 API 获取 posts 的 meta
    const { data } = await $fetch<{ data: { id: number; updatedAt: Date }[] }>(
      "/api/v1/posts/meta",
    );
    const metas = data ?? [];

    const routeList = metas.map((meta) => ({
      loc: `/blog/${meta.id}`,
      _i18nTransform: true,
      lastmod: new Date(meta.updatedAt).toISOString(),
    }));

    return routeList;
  } catch (error) {
    console.error("获取sitemap失败", error);
    return [];
  }
});
