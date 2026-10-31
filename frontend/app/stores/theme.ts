import {
  applyBaseThemeMode,
  applyThemeColorsToCSSVars,
  generateThemeColors,
  getInitialMode,
  getInitialPrimaryColor,
  withViewTransition,
} from "@/utils/theme";
import type { ThemeColors, ThemeMode } from "~/types/theme";

export const useThemeStore = defineStore("theme", () => {
  const mode = ref<ThemeMode>(getInitialMode());
  const primaryColor = ref<string>(getInitialPrimaryColor());
  const themeColors = ref<ThemeColors>(generateThemeColors(primaryColor.value, mode.value)); // 先用 light 占位

  const setMode = (newMode: ThemeMode, x: number = 0, y: number = 0) => {
    if (newMode === mode.value) return;

    const newColors = generateThemeColors(primaryColor.value, newMode);
    themeColors.value = newColors;
    mode.value = newMode;

    withViewTransition(
      () => {
        applyBaseThemeMode(newMode);
        applyThemeColorsToCSSVars(newColors);
        useCookie("theme-mode", {
          expires: new Date(Date.now() + 356 * 24 * 60 * 60 * 1000),
        }).value = newMode;
      },
      {
        direction: newMode === "light",
        x,
        y,
      },
    );
  };

  const setPrimaryColor = (newColor: string) => {
    if (newColor === primaryColor.value) return;

    const newColors = generateThemeColors(newColor, mode.value ?? "light");
    themeColors.value = newColors;
    primaryColor.value = newColor;

    withViewTransition(() => {
      if (import.meta.client) {
        applyThemeColorsToCSSVars(newColors);
        useCookie("theme-primary-color", {
          expires: new Date(Date.now() + 356 * 24 * 60 * 60 * 1000),
        }).value = newColor;
      }
    });
  };

  return {
    mode,
    primaryColor,
    themeColors,
    setMode,
    setPrimaryColor,
  };
});
