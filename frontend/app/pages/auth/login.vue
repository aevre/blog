<script setup lang="ts">
const { login } = useAuthStore();

const formData = ref({
  email: "",
  password: "",
});
</script>

<template>
  <div class="main">
    <div class="login-form-container">
      <h2>欢迎回来</h2>
      <form class="login-form" @submit.prevent="login(formData.email, formData.password)">
        <div class="input-box">
          <label for="email">邮箱</label>
          <BaseInput
            id="email"
            v-model="formData.email"
            class="input"
            type="text"
            placeholder="邮箱"
          />
        </div>
        <div class="input-box">
          <label for="password">密码</label>
          <BaseInput
            id="password"
            v-model="formData.password"
            class="input"
            type="password"
            placeholder="密码"
          />
        </div>
        <button class="login-button" type="submit">登录</button>
      </form>
    </div>
  </div>
</template>

<style lang="less" scoped>
.main {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

.login-form-container {
  min-width: 350px;
  max-width: 90vw;
  padding: 20px;
  background: var(--bg-card-base);
  backdrop-filter: blur(var(--blur-surface));
  box-shadow: var(--shadow-card-base);
  border-radius: 10px;
  color: var(--text-primary);
  text-align: center;
}

.login-form {
  display: flex;
  flex-direction: column;
}

.input-box {
  display: flex;
  margin: 10px 0;
  align-items: center;

  label {
    width: 6rem;
    font-size: 1.75rem;
  }
}

.input {
  width: 100%;
  height: 40px;
  color: var(--text-primary);
  margin-left: 1rem;
  font-weight: 700;
  border-radius: 8px;
  background: var(--bg-surface-muted);
}

.login-button {
  height: 40px;
  margin-top: 20px;
  border-radius: 10px;
  color: var(--text-on-brand);
  font-size: 1.75rem;
  background-color: var(--brand-primary);

  &:hover {
    background-color: var(--brand-hover);
  }

  &:active {
    background-color: var(--brand-active);
  }
}
</style>
