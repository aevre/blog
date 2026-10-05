export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) {
    return;
  }

  const auth = useAuthStore();

  if (to.meta.requiresInitialization && !auth.initialized) {
    return navigateTo("/setup");
  }

  if (!to.path.startsWith("/admin")) return;

  if (auth.accessToken) return;

  try {
    await auth.refresh();
  } catch {
    await auth.logout();

    return navigateTo("/auth/login");
  }
});
