<script setup lang="ts">
import type {
  MessageIconProps,
  MessageOptions,
  MessageSizeType,
  MessageType,
} from "~/types/ui/message";

interface Props {
  duration?: number;
  keepAliveOnHover?: boolean;
  max?: number;
  closable?: boolean;
  size?: MessageSizeType;
}

const {
  max = undefined,
  duration = 2000,
  keepAliveOnHover = false,
  closable = false,
  size = "medium",
} = defineProps<Props>();

interface Message {
  id: number;
  text: string;
  type: MessageType;
  keepAliveOnHover: boolean;
  icon: MessageIconProps | null;
  closable: boolean;
  timer?: ReturnType<typeof setTimeout>;
  remainTime?: number;
  startTime?: number;
  size: MessageSizeType;
}

const messageIconMap: Record<MessageType, MessageIconProps> = {
  default: { name: "" },
  success: { name: "icon-park-solid:success", color: "var(--state-success)" },
  error: { name: "icon-park-solid:error", color: "var(--state-error)" },
  info: { name: "fluent:info-sparkle-24-filled", color: "var(--state-info)" },
  warning: { name: "typcn:warning", color: "var(--state-warning)" },
};

const defaultOptions: Required<MessageOptions> = {
  duration,
  keepAliveOnHover,
  icon: null,
  closable,
  size,
};

let seed = 0;
const messages = ref<Message[]>([]);

// Create a new message
const create = (type: MessageType, text: string, options?: MessageOptions) => {
  const merged = { ...defaultOptions, ...options };
  const msg: Message = {
    id: ++seed,
    text,
    type,
    keepAliveOnHover: merged.keepAliveOnHover,
    icon: merged.icon,
    closable: merged.closable,
    size: merged.size,
  };

  if (max && messages.value.length >= max) removeOldestMessage();
  messages.value.push(msg);
  startTimer(msg, merged.duration);
};

// Remove the oldest message when max is reached
const removeOldestMessage = () => {
  const removed = messages.value.shift();
  if (!removed) return;
  if (removed.timer) clearTimeout(removed.timer);
};

// Remove a message by ID
const remove = (id: number) => {
  const index = messages.value.findIndex((m) => m.id === id);
  if (index !== -1) messages.value.splice(index, 1);
};

// Destroy a message and clear its timer
const destroy = (id: number) => {
  const index = messages.value.findIndex((m) => m.id === id);
  if (index === -1) return;
  const removed = messages.value.splice(index, 1);
  if (!removed[0]) return;
  if (removed[0].timer) clearTimeout(removed[0].timer);
};

// Start the auto-dismiss timer for a message
const startTimer = (msg: Message, duration: number) => {
  msg.startTime = Date.now();
  msg.remainTime = duration;
  msg.timer = setTimeout(() => {
    remove(msg.id);
  }, duration);
};

// Pause the timer on hover
const pauseTimer = (msg: Message) => {
  if (!msg.timer) return;
  clearTimeout(msg.timer);
  msg.remainTime = (msg.remainTime ?? 0) - (Date.now() - (msg.startTime ?? 0));
  msg.timer = undefined;
};

// Resume the timer on mouse leave
const resumeTimer = (msg: Message) => {
  if (msg.timer || (msg.remainTime ?? 0) <= 0) return;
  startTimer(msg, msg.remainTime!);
};

// Generate message API for a specific type
const createMessageApi =
  (type: Exclude<MessageType, "default">) => (text: string, options?: MessageOptions) =>
    create(type, text, options);

// Message API object
const messageApi = {
  create, // General create function for any type
  success: createMessageApi("success"),
  error: createMessageApi("error"),
  info: createMessageApi("info"),
  warning: createMessageApi("warning"),
};

provide("message", messageApi);

// TransitionGroup hook to fix leave animation jump
const beforeLeave = (el: Element) => {
  if (!(el instanceof HTMLElement)) return;
  const top = el.offsetTop;
  el.style.position = "absolute";
  el.style.top = `${top - 20}px`;
};
</script>

<template>
  <slot />
  <ClientOnly>
    <Teleport defer to="body">
      <div class="message-container">
        <TransitionGroup name="message-list" @before-leave="beforeLeave">
          <div
            v-for="msg in messages"
            :key="msg.id"
            :class="['message', `message--${msg.type}`, `message--${msg.size}`]"
            @mouseenter="msg.keepAliveOnHover && pauseTimer(msg)"
            @mouseleave="msg.keepAliveOnHover && resumeTimer(msg)"
          >
            <div class="icon">
              <Icon
                :name="msg.icon?.name || messageIconMap[msg.type].name"
                :style="{ color: msg.icon?.color || messageIconMap[msg.type].color }"
              />
            </div>
            <span class="content">
              {{ msg.text }}
            </span>
            <button v-if="msg.closable" class="close-btn" @click="destroy(msg.id)">
              <Icon name="material-symbols:close" size="20" />
            </button>
          </div>
        </TransitionGroup>
      </div>
    </Teleport>
  </ClientOnly>
</template>

<style lang="less" scoped>
.message-container {
  position: fixed;
  inset: 0;
  width: 100%;
  top: 20px;
  pointer-events: none;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.message {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-elevated);
  backdrop-filter: blur(var(--blur-elevated));
  border: 2px solid;
  position: relative;
  width: max-content;
  max-width: 600px;
  border-radius: 10px;
  padding: 8px 12px;
  margin: 5px 0;
  color: var(--text-primary);
  overflow-wrap: break-word;
  pointer-events: auto;
  box-shadow: var(--shadow-md);
  transition: all 0.3s ease;

  .icon,
  .close-btn {
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .close-btn {
    cursor: pointer;
    border: none;
    color: var(--text-primary);
    background-color: transparent;
    padding: 2px;
    border-radius: 50%;
    transition: background-color 0.2s ease;
  }

  .content {
    flex: 1;
    font-size: inherit;
  }

  &--default {
    border-color: var(--border-default);
  }

  &--success {
    border-color: var(--state-success);

    .icon {
      color: var(--state-success);
    }
  }

  &--info {
    border-color: var(--state-info);

    .icon {
      color: var(--state-info);
    }
  }

  &--error {
    border-color: var(--state-error);

    .icon {
      color: var(--state-error);
    }
  }

  &--warning {
    border-color: var(--state-warning);

    .icon {
      color: var(--state-warning);
    }
  }

  &--small {
    font-size: 1.4rem;
    padding: 6px 10px;
  }

  &--medium {
    font-size: 1.6rem;
    padding: 8px 12px;
  }

  &--large {
    font-size: 1.8rem;
    padding: 10px 14px;
  }
}

.message-list-move,
.message-list-enter-active,
.message-list-leave-active {
  transition: all 0.5s ease;
}

.message-list-enter-from,
.message-list-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
