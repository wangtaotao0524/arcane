<script lang="ts">
  import type { Icon as IconType } from "@lucide/svelte";

  let {
    variant = "default",
    text = "",
    bgColor = undefined,
    textColor = undefined,
    size = "default",
    rounded = true,
    className = "",
    icon = undefined,
    iconClass = "w-3.5 h-3.5 mr-1",
    statusVariant = undefined,
  }: {
    variant?: "default" | "secondary" | "destructive" | "outline" | "status";
    text: string;
    bgColor?: string;
    textColor?: string;
    size?: "default" | "sm" | "lg";
    rounded?: boolean;
    className?: string;
    icon?: typeof IconType;
    iconClass?: string;
    statusVariant?: "green" | "amber" | "red" | "blue" | "gray" | "purple";
  } = $props();

  // Map variant to style variables
  function getVariantStyles() {
    const variants = {
      default: {
        bg: "var(--primary)",
        color: "var(--primary-foreground)",
        border: "var(--primary)",
      },
      secondary: {
        bg: "var(--secondary)",
        color: "var(--secondary-foreground)",
        border: "var(--secondary)",
      },
      destructive: {
        bg: "var(--destructive)",
        color: "var(--destructive-foreground)",
        border: "var(--destructive)",
      },
      outline: {
        bg: "transparent",
        color: "var(--foreground)",
        border: "var(--border)",
      },
      status: {
        bg: "transparent",
        color: "var(--foreground)",
        border: "transparent",
      },
    };

    return variants[variant];
  }

  // Extended color map for all badge types
  function getCustomStyles() {
    const colorMap: Record<string, string> = {
      // Backgrounds
      "amber-100": "rgb(254 243 199)",
      "green-100": "rgb(220 252 231)",
      "red-100": "rgb(254 226 226)",
      "blue-100": "rgb(219 234 254)",
      "gray-100": "rgb(243 244 246)",
      "purple-100": "rgb(237 233 254)",
      "indigo-100": "rgb(224 231 255)",

      // Text colors (darker variants)
      "amber-900": "rgb(120 53 15)",
      "green-900": "rgb(20 83 45)",
      "red-900": "rgb(127 29 29)",
      "blue-900": "rgb(30 58 138)",
      "gray-900": "rgb(17 24 39)",
      "purple-900": "rgb(88 28 135)",
      "indigo-900": "rgb(49 46 129)",

      // Basic colors
      black: "rgb(0 0 0)",
      white: "rgb(255 255 255)",

      // Status colors
      running: "rgb(34 197 94)", // green-500
      partially: "rgb(245 158 11)", // amber-500
      stopped: "rgb(156 163 175)", // gray-400
      error: "rgb(239 68 68)", // red-500
    };

    const style = {
      bg: bgColor && bgColor in colorMap ? colorMap[bgColor] : undefined,
      color:
        textColor && textColor in colorMap ? colorMap[textColor] : undefined,
      border: bgColor && bgColor in colorMap ? colorMap[bgColor] : undefined,
    };

    return style;
  }

  // Add status presets for common status badges
  function applyStatusPreset(): {
    bg?: string;
    color?: string;
    border?: string;
  } {
    if (variant !== "status") return {};

    // Define color schemes for each status variant
    const variantStyles = {
      green: {
        bg: "rgba(34, 197, 94, 0.2)",
        color: "rgb(63, 185, 80)",
        border: "rgba(34, 197, 94, 0.2)",
      },
      amber: {
        bg: "rgba(234, 179, 8, 0.2)",
        color: "rgb(210, 153, 34)",
        border: "rgba(234, 179, 8, 0.2)",
      },
      red: {
        bg: "rgba(239, 68, 68, 0.15)",
        color: "rgb(185, 28, 28)",
        border: "rgba(239, 68, 68, 0.15)",
      },
      blue: {
        bg: "rgba(59, 130, 246, 0.15)",
        color: "rgb(37, 99, 235)",
        border: "rgba(59, 130, 246, 0.15)",
      },
      gray: {
        bg: "rgba(156, 163, 175, 0.15)",
        color: "rgb(75, 85, 99)",
        border: "rgba(156, 163, 175, 0.15)",
      },
      purple: {
        bg: "rgba(139, 92, 246, 0.15)",
        color: "rgb(107, 33, 168)",
        border: "rgba(139, 92, 246, 0.15)",
      },
    };

    // Map common status texts to variants for backward compatibility
    const statusToVariantMap: Record<string, keyof typeof variantStyles> = {
      running: "green",
      "in use": "green",
      managed: "green",
      unused: "amber",
      partially: "amber",
      "partially running": "amber",
      restarting: "amber",
      warning: "amber",
      error: "red",
      exited: "red",
      dead: "red",
      failed: "red",
      pending: "blue",
      creating: "blue",
      stopped: "gray",
      paused: "gray",
      external: "purple",
    };

    // If direct status variant is provided, use it
    if (statusVariant && statusVariant in variantStyles) {
      return variantStyles[statusVariant];
    }

    // Otherwise try to infer from text (backward compatibility)
    if (text) {
      const lowerText = text.toLowerCase();

      // Check for exact matches in our map
      if (lowerText in statusToVariantMap) {
        return variantStyles[statusToVariantMap[lowerText]];
      }

      // Check for partial matches
      for (const [key, variant] of Object.entries(statusToVariantMap)) {
        if (lowerText.includes(key)) {
          return variantStyles[variant];
        }
      }
    }

    // Default to blue if no match found
    return variantStyles.blue;
  }

  // Combine variant and custom styles
  const baseStyles = $derived(getVariantStyles());
  const customStyles = $derived(getCustomStyles());
  const statusStyles = $derived(applyStatusPreset());

  const style = $derived(
    [
      `background-color: ${customStyles.bg || statusStyles.bg || baseStyles.bg};`,
      `color: ${customStyles.color || statusStyles.color || baseStyles.color};`,
      `border: 1px solid ${customStyles.border || statusStyles.border || baseStyles.border};`,
    ].join(" ")
  );

  // Size classes
  const sizeClass = $derived(
    size === "sm"
      ? "text-xs py-0 px-2"
      : size === "lg"
        ? "text-sm py-1 px-3"
        : "text-xs py-0.5 px-2.5"
  );

  const roundedClass = $derived(rounded ? "rounded-full" : "rounded-md");
</script>

<span
  class={`inline-flex items-center font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 ${sizeClass} ${roundedClass} ${className}`}
  {style}
>
  {#if icon}
    {@const Icon = icon}
    <Icon class={iconClass} />
  {/if}
  {text}
</span>
