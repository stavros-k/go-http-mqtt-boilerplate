import type { BundledLanguage, BundledTheme, CodeToHastOptions } from "shiki";

export type ThemeCategory = "light" | "dark";
type ThemeItem = { id: BundledTheme; label: string };

const LightThemes: Record<string, ThemeItem> = {
    "catppuccin-latte": { id: "catppuccin-latte", label: "Catppuccin Latte" },
    "everforest-light": { id: "everforest-light", label: "Everforest Light" },
    "github-light": { id: "github-light", label: "GitHub Light" },
    "github-light-default": {
        id: "github-light-default",
        label: "GitHub Light Default",
    },
    "github-light-high-contrast": {
        id: "github-light-high-contrast",
        label: "GitHub Light High Contrast",
    },
    "gruvbox-light-hard": {
        id: "gruvbox-light-hard",
        label: "Gruvbox Light Hard",
    },
    "gruvbox-light-medium": {
        id: "gruvbox-light-medium",
        label: "Gruvbox Light Medium",
    },
    "gruvbox-light-soft": {
        id: "gruvbox-light-soft",
        label: "Gruvbox Light Soft",
    },
    "kanagawa-lotus": { id: "kanagawa-lotus", label: "Kanagawa Lotus" },
    "light-plus": { id: "light-plus", label: "Light Plus" },
    "material-theme-lighter": {
        id: "material-theme-lighter",
        label: "Material Theme Lighter",
    },
    "min-light": { id: "min-light", label: "Min Light" },
    "rose-pine-dawn": { id: "rose-pine-dawn", label: "Rosé Pine Dawn" },
    "slack-ochin": { id: "slack-ochin", label: "Slack Ochin" },
    "snazzy-light": { id: "snazzy-light", label: "Snazzy Light" },
    "solarized-light": { id: "solarized-light", label: "Solarized Light" },
    "vitesse-light": { id: "vitesse-light", label: "Vitesse Light" },
};

const DarkThemes: Record<string, ThemeItem> = {
    andromeeda: { id: "andromeeda", label: "Andromeeda" },
    "aurora-x": { id: "aurora-x", label: "Aurora X" },
    "ayu-dark": { id: "ayu-dark", label: "Ayu Dark" },
    "catppuccin-frappe": {
        id: "catppuccin-frappe",
        label: "Catppuccin Frappé",
    },
    "catppuccin-macchiato": {
        id: "catppuccin-macchiato",
        label: "Catppuccin Macchiato",
    },
    "catppuccin-mocha": { id: "catppuccin-mocha", label: "Catppuccin Mocha" },
    "dark-plus": { id: "dark-plus", label: "Dark Plus" },
    dracula: { id: "dracula", label: "Dracula Theme" },
    "dracula-soft": { id: "dracula-soft", label: "Dracula Theme Soft" },
    "everforest-dark": { id: "everforest-dark", label: "Everforest Dark" },
    "github-dark": { id: "github-dark", label: "GitHub Dark" },
    "github-dark-default": {
        id: "github-dark-default",
        label: "GitHub Dark Default",
    },
    "github-dark-dimmed": {
        id: "github-dark-dimmed",
        label: "GitHub Dark Dimmed",
    },
    "github-dark-high-contrast": {
        id: "github-dark-high-contrast",
        label: "GitHub Dark High Contrast",
    },
    "gruvbox-dark-hard": {
        id: "gruvbox-dark-hard",
        label: "Gruvbox Dark Hard",
    },
    "gruvbox-dark-medium": {
        id: "gruvbox-dark-medium",
        label: "Gruvbox Dark Medium",
    },
    "gruvbox-dark-soft": {
        id: "gruvbox-dark-soft",
        label: "Gruvbox Dark Soft",
    },
    houston: { id: "houston", label: "Houston" },
    "kanagawa-dragon": { id: "kanagawa-dragon", label: "Kanagawa Dragon" },
    "kanagawa-wave": { id: "kanagawa-wave", label: "Kanagawa Wave" },
    laserwave: { id: "laserwave", label: "LaserWave" },
    "material-theme": { id: "material-theme", label: "Material Theme" },
    "material-theme-darker": {
        id: "material-theme-darker",
        label: "Material Theme Darker",
    },
    "material-theme-ocean": {
        id: "material-theme-ocean",
        label: "Material Theme Ocean",
    },
    "material-theme-palenight": {
        id: "material-theme-palenight",
        label: "Material Theme Palenight",
    },
    "min-dark": { id: "min-dark", label: "Min Dark" },
    monokai: { id: "monokai", label: "Monokai" },
    "night-owl": { id: "night-owl", label: "Night Owl" },
    nord: { id: "nord", label: "Nord" },
    plastic: { id: "plastic", label: "Plastic" },
    poimandres: { id: "poimandres", label: "Poimandres" },
    red: { id: "red", label: "Red" },
    "rose-pine": { id: "rose-pine", label: "Rosé Pine" },
    "rose-pine-moon": { id: "rose-pine-moon", label: "Rosé Pine Moon" },
    "slack-dark": { id: "slack-dark", label: "Slack Dark" },
    "solarized-dark": { id: "solarized-dark", label: "Solarized Dark" },
    "synthwave-84": { id: "synthwave-84", label: "Synthwave '84" },
    "tokyo-night": { id: "tokyo-night", label: "Tokyo Night" },
    vesper: { id: "vesper", label: "Vesper" },
    "vitesse-black": { id: "vitesse-black", label: "Vitesse Black" },
    "vitesse-dark": { id: "vitesse-dark", label: "Vitesse Dark" },
};

function getThemesAs(themes: Record<string, ThemeItem>, category: ThemeCategory) {
    const result: Record<string, ThemeItem & { category: ThemeCategory }> = {};
    for (const [id, theme] of Object.entries(themes)) {
        result[id] = {
            ...theme,
            category,
        };
    }
    return result;
}

export const THEMES: Record<string, ThemeItem & { category: ThemeCategory }> = {
    // Defaults
    light: { id: "one-light", label: "One Light (Default)", category: "light" },
    dark: {
        id: "one-dark-pro",
        label: "One Dark Pro (Default)",
        category: "dark",
    },
    ...getThemesAs(LightThemes, "light"),
    ...getThemesAs(DarkThemes, "dark"),
} as const;

export type CodeTheme = keyof typeof THEMES;

export function getShikiOptions(lang: BundledLanguage): CodeToHastOptions {
    const themes: Record<string, BundledTheme> = {};
    for (const [id, theme] of Object.entries(THEMES)) {
        themes[id] = theme.id;
    }
    return {
        lang: lang,
        themes: themes,
        cssVariablePrefix: "--shiki-",
        // This makes sure that no default styles are applied,
        // so we don't need to use !important
        defaultColor: false,
        transformers: [
            {
                pre: (node) => {
                    const existingClass = node.properties.class || "";
                    node.properties.class =
                        `${existingClass} p-4 rounded-xl border-2 border-border-primary shadow-sm`.trim();
                },
            },
        ],
    };
}
