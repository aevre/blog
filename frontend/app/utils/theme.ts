import tinycolor from "tinycolor2";
import { nextTick } from "vue";
import type { ThemeColors, ThemeMode } from "~/types/theme";

interface BrandScale {
  brand50: string;
  brand100: string;
  brand200: string;
  brand300: string;
  brand400: string;
  brand500: string;
  brand600: string;
  brand700: string;
  brand800: string;
  brand900: string;
}

function generateBrandScale(baseColor: string): BrandScale {
  const base = tinycolor(baseColor);
  const { h, s, l } = base.toHsl();

  // Tight scale: narrow saturation, fixed lightness targets
  const sMin = Math.max(s * 0.7, 0.15);
  const sMax = Math.min(s * 1.15, 1);

  return {
    brand50: tinycolor({ h, s: sMin, l: 0.96 }).toHexString(),
    brand100: tinycolor({ h, s: sMin, l: 0.92 }).toHexString(),
    brand200: tinycolor({ h, s: Math.max(s * 0.8, 0.2), l: 0.86 }).toHexString(),
    brand300: tinycolor({ h, s: Math.max(s * 0.85, 0.25), l: 0.78 }).toHexString(),
    brand400: tinycolor({ h, s: Math.max(s * 0.9, 0.3), l: 0.72 }).toHexString(),
    brand500: base.toHexString(),
    brand600: tinycolor({
      h,
      s: Math.min(s * 1.05, sMax),
      l: Math.max(l - 0.1, 0.3),
    }).toHexString(),
    brand700: tinycolor({
      h,
      s: Math.min(s * 1.1, sMax),
      l: Math.max(l - 0.18, 0.22),
    }).toHexString(),
    brand800: tinycolor({
      h,
      s: Math.min(s * 1.12, sMax),
      l: Math.max(l - 0.28, 0.15),
    }).toHexString(),
    brand900: tinycolor({
      h,
      s: Math.min(s * 1.15, sMax),
      l: Math.max(l - 0.38, 0.08),
    }).toHexString(),
  };
}

export function generateThemeColors(baseColor: string, mode: ThemeMode): ThemeColors {
  const scale = generateBrandScale(baseColor);

  const hover =
    mode === "light" ? tinycolor(scale.brand500).darken(5) : tinycolor(scale.brand500).lighten(5);
  const active =
    mode === "light" ? tinycolor(scale.brand500).darken(10) : tinycolor(scale.brand500).lighten(10);
  const disabled = tinycolor.mix(scale.brand500, "#b0b0b0", 60);

  // On-brand text: check contrast
  const whiteContrast = tinycolor.readability(scale.brand500, "#ffffff");
  const onPrimary = whiteContrast >= 4.5 ? "#ffffff" : "#1a1a1e";

  const hoverContrast = tinycolor.readability(hover, "#ffffff");
  const onPrimaryHover = hoverContrast >= 4.5 ? "#ffffff" : "#1a1a1e";

  const activeContrast = tinycolor.readability(active, "#ffffff");
  const onPrimaryActive = activeContrast >= 4.5 ? "#ffffff" : "#1a1a1e";

  return {
    ...scale,
    primary: scale.brand500,
    hover: hover.toHexString(),
    active: active.toHexString(),
    disabled: disabled.toHexString(),
    bg: scale.brand50,
    onPrimary,
    onPrimaryHover,
    onPrimaryActive,
  };
}

export function applyThemeColorsToCSSVars(colors: ThemeColors) {
  const root = document.documentElement;

  // Brand scale
  root.style.setProperty("--brand-50", colors.brand50);
  root.style.setProperty("--brand-100", colors.brand100);
  root.style.setProperty("--brand-200", colors.brand200);
  root.style.setProperty("--brand-300", colors.brand300);
  root.style.setProperty("--brand-400", colors.brand400);
  root.style.setProperty("--brand-500", colors.brand500);
  root.style.setProperty("--brand-600", colors.brand600);
  root.style.setProperty("--brand-700", colors.brand700);
  root.style.setProperty("--brand-800", colors.brand800);
  root.style.setProperty("--brand-900", colors.brand900);

  // Semantic brand tokens
  root.style.setProperty("--brand-primary", colors.primary);
  root.style.setProperty("--brand-hover", colors.hover);
  root.style.setProperty("--brand-active", colors.active);
  root.style.setProperty("--brand-disabled", colors.disabled);

  // Unified state tokens (foreground)
  root.style.setProperty("--state-hover-brand", colors.hover);
  root.style.setProperty("--state-active-brand", colors.active);

  // Brand backgrounds - now reference unified state tokens in base.less
  // No override needed

  // Text on brand
  root.style.setProperty("--text-on-brand", colors.onPrimary);
  root.style.setProperty("--text-on-brand-hover", colors.onPrimaryHover);
  root.style.setProperty("--text-on-brand-active", colors.onPrimaryActive);

  // Derived tokens - all reference unified state tokens
  root.style.setProperty("--border-active", colors.primary);
  root.style.setProperty("--selection-bg", colors.primary);
  root.style.setProperty("--scrollbar-thumb-bg", colors.primary);
  root.style.setProperty("--scrollbar-thumb-hover", colors.hover);
  root.style.setProperty("--scrollbar-thumb-active", colors.active);
}

export function applyBaseThemeMode(mode: ThemeMode) {
  const root = document.documentElement;
  root.setAttribute("data-theme", mode);
}

export interface TransitionOption {
  direction?: boolean;
  x?: number;
  y?: number;
}

export function withViewTransition(
  applyFn: () => void,
  option: TransitionOption = { direction: true, x: 0, y: 0 },
) {
  if (typeof document !== "undefined" && document.startViewTransition) {
    const transition = document.startViewTransition(async () => {
      applyFn();
      await nextTick();
    });

    transition.ready
      .then(() => {
        const innerHeight = window.innerHeight;
        const innerWidth = window.innerWidth;

        const x = option.x ?? 0;
        const y = option.y ?? 0;

        const radius = Math.max(
          Math.hypot(x, y),
          Math.hypot(innerWidth - x, y),
          Math.hypot(x, innerHeight - y),
          Math.hypot(innerWidth - x, innerHeight - y),
        );

        const clipPath = [`circle(0 at ${x}px ${y}px)`, `circle(${radius}px at ${x}px ${y}px)`];

        const animation = document.documentElement.animate(
          {
            clipPath: option.direction ? clipPath : [...clipPath].reverse(),
          },
          {
            duration: 400,
            easing: "ease-in",
            fill: "both",
            pseudoElement: option.direction
              ? "::view-transition-new(root)"
              : "::view-transition-old(root)",
          },
        );

        const cleanup = () => animation.cancel();
        void transition.finished.then(cleanup, cleanup);
      })
      .catch(console.warn);
  } else {
    applyFn();
  }
}

export function getInitialMode(): ThemeMode {
  if (import.meta.server) return "light";

  const themeMode = useCookie<ThemeMode>("theme-mode", {
    default: () => "light",
  });

  return themeMode.value;
}

export function getInitialPrimaryColor(): string {
  // if (import.meta.server) return "#99a2ff";

  const color = useCookie<string>("theme-primary-color");

  return color.value ? decodeURIComponent(color.value) : "#99a2ff";
}
