export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore();

  // if (to.meta.requiresInitialization && !auth.initialized) {
  //   return navigateTo("/setup");
  // }

  if (!to.path.startsWith("/admin")) return;

  await auth.fetchUser();

  if (!auth.isAuthenticated) return navigateTo("/auth/login");
});
