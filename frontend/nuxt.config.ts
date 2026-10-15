// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",

  devServer: {
    host: "0.0.0.0",
    port: 3000,
  },
  // 启用开发工具和时间线调试
  devtools: {
    enabled: true,
    timeline: {
      enabled: true,
    },
  },

  // 应用信息配置
  app: {
    head: {
      title: "Aevre's Blog",
      htmlAttrs: { lang: "en" },
      charset: "utf-8",
      viewport: "width=device-width, initial-scale=1, maximum-scale=1",
      link: [{ rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
      script: [{}],
    },
    pageTransition: { name: "page", mode: "out-in" },
    layoutTransition: { name: "layout", mode: "out-in" },
  },

  // 站点基本信息
  site: {
    url: "https://blog.immortel.top",
    name: "Aevre's Blog",
  },

  // Sitemap 配置
  sitemap: {
    sources: ["/api/sitemap"],
    exclude: ["/admin/**", "/playground"],
    // cacheMaxAgeSeconds: 6 * 60 * 60, // 6小时缓存，可按需开启
    autoLastmod: true, // 自动生成最后修改时间，方便爬虫
  },

  experimental: {
    typescriptPlugin: true,
  },

  // TypeScript 配置
  typescript: {
    typeCheck: true,
  },

  routeRules: {
    "/admin/**": { appLayout: "admin-layout" },
    "/auth/**": { appLayout: "auth-layout" },
  },

  // Nitro 配置
  nitro: {
    esbuild: {
      options: {
        target: "es2022",
      },
    },
  },

  fonts: {
    provider: "bunny",
    families: [
      { name: "Open Sans", provider: "bunny" },
      {
        name: "Noto Sans SC",
        provider: "bunny",
      },
      { name: "Caveat", provider: "bunny" },
    ],
  },

  // Vite 配置
  vite: {
    css: {
      devSourcemap: true, // 开启 CSS Source Map，方便调试
    },
    optimizeDeps: {
      include: ["tinycolor2", "@vueuse/core", "lodash-es"],
    },
  },

  // 只将 Vue 单文件组件纳入自动导入，保留组件目录内的类型和实现文件。
  components: [{ path: "~/components", extensions: [".vue"] }],

  // 全局样式
  css: ["~/assets/styles/main.less"],

  i18n: {
    strategy: "prefix_except_default",
    defaultLocale: "zh",
    locales: [
      { code: "en", name: "English", file: "en.json", iso: "en-US", dir: "ltr" },
      { code: "zh", name: "简体中文", file: "zh.json", iso: "zh-CN", dir: "ltr" },
      { code: "ja", name: "日本語", file: "ja.json", iso: "ja-JP", dir: "ltr" },
    ],
    disablePageLocales: true,
    redirects: true,
    localeCookie: "user-locale",
    meta: true,
    metaBaseUrl: "https://blog.immortel.top/",
    autoDetectLanguage: true,
    autoDetectPath: "/",
  },

  icon: {
    clientBundle: {
      scan: true,
    },
    serverBundle: {
      collections: [
        "codicon",
        "fluent",
        "hugeicons",
        "icon-park-solid",
        "iconamoon",
        "ion",
        "mingcute",
        "ri",
        "streamline-logos",
        "typcn",
        "weui",
      ],
    },
  },

  image: {},

  modules: [
    "@nuxt/fonts",
    "@nuxt/eslint",
    "@nuxt/icon",
    "@nuxt/scripts",
    "@pinia/nuxt",
    "@nuxtjs/sitemap",
    "@nuxt/image",
    "motion-v/nuxt",
    "nuxt-i18n-micro",
  ],
});
