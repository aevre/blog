<script setup lang="ts">
import type { ApiResponse } from "~/types/api";

definePageMeta({ layout: "auth-layout" });

interface InitializeReq {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  siteName: string;
  logo: string;
  greeting: string;
  description: string;
}

interface InitializeData {
  initialized: boolean;
}

const form = reactive<InitializeReq>({
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
  siteName: "",
  logo: "",
  greeting: "",
  description: "",
});

const loading = ref(false);
const error = ref("");

const initialize = async () => {
  error.value = "";

  if (form.password !== form.confirmPassword) {
    error.value = "两次输入的密码不一致";
    return;
  }

  loading.value = true;

  try {
    const { data } = await useApiFetch<ApiResponse<InitializeData>>("/setup", {
      method: "post",
      body: form,
    });

    if (!data.value?.data.initialized) {
      error.value = "初始化未完成";
      return;
    }

    await navigateTo("/auth/login");
  } catch (err) {
    console.error(err);
    error.value = "初始化失败，请检查输入后重试";
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <main class="setup">
    <div class="container">
      <header class="header">
        <div class="title">
          <h1>初始化站点</h1>
          <p>创建管理员账户并完成站点的基本配置。</p>
        </div>
      </header>

      <form class="form" @submit.prevent="initialize">
        <section class="form-card">
          <header class="form-card-header">
            <div>
              <h2>管理员账户</h2>
              <p>使用此账户登录管理后台。</p>
            </div>
          </header>

          <div class="form-grid">
            <div class="field">
              <label for="username">用户名</label>

              <input
                id="username"
                v-model="form.username"
                name="username"
                type="text"
                autocomplete="username"
                minlength="3"
                maxlength="32"
                required
              />

              <span class="field-hint">3–32 个字符</span>
            </div>

            <div class="field">
              <label for="email">邮箱</label>

              <input
                id="email"
                v-model="form.email"
                name="email"
                type="email"
                autocomplete="email"
                maxlength="255"
                required
              />
            </div>

            <div class="field">
              <label for="password">密码</label>

              <input
                id="password"
                v-model="form.password"
                name="password"
                type="password"
                autocomplete="new-password"
                minlength="8"
                maxlength="128"
                required
              />

              <span class="field-hint">至少 8 个字符</span>
            </div>

            <div class="field">
              <label for="confirm-password">确认密码</label>

              <input
                id="confirm-password"
                v-model="form.confirmPassword"
                name="confirmPassword"
                type="password"
                autocomplete="new-password"
                minlength="8"
                maxlength="128"
                required
              />
            </div>
          </div>
        </section>

        <section class="form-card">
          <header class="form-card-header">
            <div>
              <h2>站点信息</h2>
              <p>设置站点的名称和公开信息。</p>
            </div>
          </header>

          <div class="form-grid">
            <div class="field field-full">
              <label for="site-name">站点名称</label>

              <input
                id="site-name"
                v-model="form.siteName"
                name="siteName"
                type="text"
                maxlength="100"
                required
              />
            </div>

            <div class="field field-full">
              <label for="logo">
                Logo
                <span class="field-optional">可选</span>
              </label>

              <input
                id="logo"
                v-model="form.logo"
                name="logo"
                type="url"
                maxlength="2048"
                placeholder="https://example.com/logo.png"
              />
            </div>

            <div class="field field-full">
              <label for="greeting">
                欢迎语
                <span class="field-optional">可选</span>
              </label>

              <input
                id="greeting"
                v-model="form.greeting"
                name="greeting"
                type="text"
                maxlength="255"
              />
            </div>

            <div class="field field-full">
              <label for="description">
                站点描述
                <span class="field-optional">可选</span>
              </label>

              <textarea
                id="description"
                v-model="form.description"
                name="description"
                maxlength="1000"
                rows="4"
              />
            </div>
          </div>
        </section>

        <div v-if="error" class="form-error">
          {{ error }}
        </div>

        <div class="form-actions">
          <button class="submit" type="submit" :disabled="loading">
            {{ loading ? "正在初始化…" : "完成初始化" }}
          </button>
        </div>
      </form>

      <footer class="footer">初始化完成后，你可以使用管理员账户登录。</footer>
    </div>
  </main>
</template>

<style lang="less" scoped>
.setup {
  min-height: 100vh;
  box-sizing: border-box;
  padding: 8rem 2.4rem;
  color: var(--text-primary);
  font-family: var(--font-family-base);
  font-weight: var(--font-weight-base);
  line-height: var(--line-height-base);
  -webkit-font-smoothing: var(--font-smoothing-webkit);
  -moz-osx-font-smoothing: var(--font-smoothing-moz);
  display: flex;
  justify-content: center;
}

.container {
  width: 100%;
  max-width: 64rem;
}

.header {
  margin-bottom: 3.2rem;
}

.brand-mark {
  width: 3.2rem;
  height: 3.2rem;
  display: grid;
  place-items: center;
  color: var(--text-on-brand);
  background: var(--brand-primary);
  border-radius: 0.8rem;
  font-size: 1.5rem;
  font-weight: 600;
}

.title {
  h1 {
    margin: 0 0 0.8rem;
    color: var(--text-primary);
    font-size: 3.2rem;
    line-height: 1.2;
    font-weight: 600;
    letter-spacing: -0.02em;
  }

  p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 1.5rem;
  }
}

.form {
  display: flex;
  flex-direction: column;
  gap: 1.6rem;
}

.form-card {
  padding: 2.4rem;
  background: var(--bg-card-base);
  backdrop-filter: blur(var(--blur-surface));
  border: 1px solid var(--border-color-card);
  border-radius: var(--radius-card);
  box-shadow: var(--shadow-card-base);
  transition: var(--transition-base);

  &:focus-within {
    border-color: var(--border-hover);
  }
}

.form-card-header {
  margin-bottom: 2.4rem;

  h2 {
    margin: 0 0 0.4rem;
    color: var(--text-primary);
    font-size: 1.8rem;
    line-height: 1.4;
    font-weight: 600;
  }

  p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 1.4rem;
  }
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 2rem 1.6rem;
}

.field {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.7rem;

  label {
    color: var(--text-primary);
    font-size: 1.4rem;
    font-weight: 500;
  }

  input,
  textarea {
    width: 100%;
    box-sizing: border-box;
    padding: 1rem 1.2rem;
    color: var(--text-primary);
    background: var(--bg-surface);
    border: 1px solid var(--border-default);
    border-radius: 0.8rem;
    outline: none;
    font: inherit;
    font-size: 1.4rem;
    transition:
      border-color 0.2s ease,
      box-shadow 0.2s ease,
      background 0.2s ease;

    &::placeholder {
      color: var(--text-tertiary);
    }

    &:hover {
      border-color: var(--border-hover);
    }

    &:focus {
      border-color: var(--border-active);
      box-shadow: var(--focus-ring);
    }

    &:disabled {
      color: var(--text-disabled);
      background: var(--bg-disabled);
      border-color: var(--border-disabled);
      cursor: not-allowed;
    }
  }

  textarea {
    min-height: 10rem;
    resize: vertical;
  }
}

.field-full {
  grid-column: 1 / -1;
}

.field-hint,
.field-optional {
  color: var(--text-tertiary);
  font-size: 1.2rem;
  font-weight: 400;
}

.field-optional {
  margin-left: 0.4rem;
}

.form-error {
  margin: 0;
  padding: 1rem 1.2rem;
  color: var(--state-error);
  background: var(--state-error-bg);
  border: 1px solid color-mix(in srgb, var(--state-error) 20%, transparent);
  border-radius: 0.8rem;
  font-size: 1.4rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.submit {
  min-width: 12rem;
  padding: 1rem 1.8rem;
  color: var(--text-on-brand);
  background: var(--brand-primary);
  border: 1px solid var(--brand-primary);
  border-radius: 0.8rem;
  font: inherit;
  font-size: 1.4rem;
  font-weight: 500;
  cursor: pointer;
  transition:
    background 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.1s ease;

  &:focus-visible {
    box-shadow: var(--focus-ring);
  }

  &:disabled {
    color: var(--text-on-brand);
    background: var(--brand-disabled);
    border-color: var(--brand-disabled);
    cursor: not-allowed;
  }

  &:hover:not(:disabled) {
    background: var(--brand-hover);
    border-color: var(--brand-hover);
  }

  &:active:not(:disabled) {
    background: var(--brand-active);
    border-color: var(--brand-active);
    transform: translateY(1px);
  }
}

.footer {
  margin-top: 2.4rem;
  color: var(--text-tertiary);
  font-size: 1.2rem;
  text-align: center;
}

@media (width < 640px) {
  .setup {
    padding: 4rem 1.6rem;
  }

  .title {
    h1 {
      font-size: 2.8rem;
    }
  }

  .form-card {
    padding: 2rem;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .field-full {
    grid-column: auto;
  }

  .form-actions {
    .submit {
      width: 100%;
    }
  }
}
</style>
