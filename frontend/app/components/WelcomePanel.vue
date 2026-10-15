<script setup lang="ts">
const handleDownArrowClick = () => {
  const innerHeight = window.innerHeight;
  const scrollableHeight = document.documentElement.scrollHeight - innerHeight;

  if (scrollableHeight > innerHeight) {
    window.scrollTo({
      top: innerHeight,
      behavior: "smooth",
    });
  } else {
    window.scrollTo({
      top: scrollableHeight,
      behavior: "smooth",
    });
  }
};
</script>

<template>
  <div class="welcome">
    <div class="welcome-content">
      <div class="introduction">
        <div class="avatar">
          <NuxtImg
            src="https://q1.qlogo.cn/g?b=qq&nk=188191770&s=640"
            alt="avatar"
            class="avatar-image"
          />
        </div>
        <p class="greetings">Cherish every moment.</p>
      </div>
    </div>
    <button class="down-arrow" @click="handleDownArrowClick">
      <Icon name="mingcute:down-line" size="50" />
    </button>
  </div>
</template>

<style lang="less" scoped>
.welcome {
  width: 100%;
  height: 100vh;
  position: relative;
}

.introduction {
  height: calc(100vh - 60px);
  max-width: 600px;
  width: 95%;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  margin-bottom: 20px;
  text-align: center;
}

.avatar {
  height: 130px;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: fade-in-down 1s ease-in-out;

  .avatar-image {
    width: 100px;
    height: 100px;
    border: 3px solid var(--brand-disabled);
    border-radius: 50%;
    object-fit: cover;
    animation: breathing-light 5s ease-in-out infinite;
    box-shadow: var(--shadow-md);
  }
}

.greetings {
  position: relative;
  height: 60px;
  width: 100%;
  background: var(--bg-nav-base);
  color: var(--text-primary);
  border: 1px solid var(--border-color-nav);
  border-radius: 12px;
  line-height: 60px;
  margin: 25px 0;
  font-family: var(--font-family-caveat);
  font-size: 2.5rem;
  font-weight: 800;
  animation: fade-in-down 1s ease;
  box-shadow: var(--shadow-md);
  overflow: hidden;
}

.contact-list {
  display: flex;
  max-width: 210px;
  list-style: none;
  width: max-content;
  gap: 8px;
  justify-content: start;
  animation: fade-in-down 1s ease-in-out;
  flex-wrap: wrap;
}

.down-arrow {
  position: absolute;
  bottom: 20px;
  left: 50%;
  cursor: pointer;
  transform: translateX(-50%);
  color: var(--text-secondary);
  animation: bounce 2.5s infinite ease-in-out;
  background-color: transparent;
}

@keyframes bounce {
  0%,
  100% {
    transform: translate(-50%, 0);
  }

  50% {
    transform: translate(-50%, 10px);
  }
}

@keyframes breathing-light {
  0%,
  100% {
    transform: scale(1);
  }

  50% {
    transform: scale(1.05);
  }
}
</style>
