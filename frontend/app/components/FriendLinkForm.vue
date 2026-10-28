<script setup lang="ts">
import { onClickOutside } from "@vueuse/core";

/**
 * v-model:show 由父组件控制弹窗显隐
 */
const show = defineModel<boolean>("show", { default: false });

/**
 * 点击外部关闭
 */
const contentCard = useTemplateRef<HTMLDivElement>("content");
onMounted(() => {
  onClickOutside(contentCard, () => {
    show.value = false;
  });
});

watch(show, () => {
  if (show.value) {
    document.body.style.overflow = "hidden";
  } else {
    document.body.style.overflow = "";
  }
});

/**
 * 表单数据
 */
interface LinkFormData {
  name: string;
  url: string;
  description: string;
  avatar: string;
}

const form = ref<LinkFormData>({
  name: "",
  url: "",
  description: "",
  avatar: "",
});

const loading = ref(false);
const errorMsg = ref("");
const successMsg = ref("");

/**
 * 简单校验
 */
function validate() {
  if (!form.value.name.trim()) return "请输入站点名称";
  if (!form.value.url.trim()) return "请输入站点链接";
  try {
    new URL(form.value.url);
  } catch {
    return "站点链接格式错误";
  }
  try {
    new URL(form.value.avatar);
  } catch {
    return "站点LOGO链接格式错误";
  }
  return "";
}

const message = useMessage();
const { $ts } = useI18n();

async function handleSubmit() {
  errorMsg.value = "";
  successMsg.value = "";
  const err = validate();
  if (err) {
    errorMsg.value = err;
    message.error(err);
    return;
  }
  loading.value = true;
  try {
    await useApiFetch("/links/apply-link", {
      method: "post",
      body: form.value,
    });

    form.value = {
      name: "",
      avatar: "",
      description: "",
      url: "",
    };

    show.value = false;
    message.success($ts("submission.friendLink"));
  } catch (e) {
    console.warn(e);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="friend-link-form">
      <div v-if="show" class="friend-link-form">
        <div ref="content" class="content">
          <form @submit.prevent="handleSubmit">
            <h2 class="title">申请友链</h2>

            <label class="label">
              <span>站点名称</span>
              <input v-model="form.name" type="text" class="input" />
            </label>
            <label class="label">
              <span>站点链接</span>
              <input v-model="form.url" type="text" placeholder="https://" class="input" />
            </label>
            <label class="label">
              <span>站点描述</span>
              <input v-model="form.description" type="text" class="input" />
            </label>
            <label class="label">
              <span>站点LOGO</span>
              <input v-model="form.avatar" type="text" placeholder="https://" class="input" />
            </label>
            <button type="submit" class="submit-btn">提交</button>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style lang="less" scoped>
.friend-link-form-enter-active,
.friend-link-form-leave-active {
  transition:
    scale 0.3s ease-in-out,
    opacity 0.2s ease-in-out;
}

.friend-link-form-enter-from,
.friend-link-form-leave-to {
  scale: 0.3;
  opacity: 0;
}

.friend-link-form {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  background: var(--bg-overlay);
}

.content {
  max-width: 480px;
  width: 95%;
  background: var(--bg-elevated);
  backdrop-filter: blur(var(--blur-elevated));
  box-shadow:
    var(--shadow-xl),
    inset 0 0 0 1px var(--border-glass);
  border-radius: 10px;
  padding: 20px;
  color: var(--text-primary);
}

.title {
  margin-bottom: 20px;
}

.label {
  display: block;
  margin: 5px 0;
}

.input {
  width: 100%;
  padding: 8px 10px;
  background: var(--bg-content);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
  border-radius: 6px;
}

.input:focus {
  outline: 1px solid var(--brand-primary);
  border: 1px solid transparent;
}

.submit-btn {
  display: block;
  padding: 8px 12px;
  background: var(--brand-primary);
  color: var(--text-on-brand);
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  margin-top: 10px;
  margin-left: auto;

  &:hover {
    background-color: var(--brand-hover);
  }

  &:active {
    background-color: var(--brand-active);
  }
}

.submit-btn:disabled {
  background: var(--brand-disabled);
  cursor: not-allowed;
}
</style>
