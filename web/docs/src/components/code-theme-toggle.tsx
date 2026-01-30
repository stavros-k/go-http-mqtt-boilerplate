"use client";
import { useEffect, useRef, useState } from "react";
import { IoChevronDown } from "react-icons/io5";
import { type CodeTheme, THEMES } from "@/utils/code-theme";
import type { Theme } from "./theme-toggle-client";

// Resolve code theme for current UI theme
function resolveCodeTheme(theme: Theme): CodeTheme {
    const storageKey = `codeTheme-${theme}`;
    const saved = localStorage.getItem(storageKey) as CodeTheme | null;
    const themeIDs = Object.keys(THEMES);

    // Validate saved theme matches the UI theme category
    if (saved && themeIDs.includes(saved) && THEMES[saved].category === theme) {
        document.documentElement.dataset.codeTheme = saved;
        return saved;
    }

    // Default based on UI theme
    const defaultTheme = (theme === "light" ? "light" : "dark") as CodeTheme;
    document.documentElement.dataset.codeTheme = defaultTheme;
    return defaultTheme;
}

export const CodeThemeToggle = () => {
    const [codeTheme, setCodeTheme] = useState<CodeTheme>("dark");
    const [uiTheme, setUiTheme] = useState<Theme>("dark");
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        // Track UI theme changes
        const updateUiTheme = () => {
            const theme = document.documentElement.dataset.theme as Theme | undefined;
            const newTheme = theme || "dark";
            setUiTheme(newTheme);
            // Resolve appropriate code theme when UI theme changes
            const resolvedTheme = resolveCodeTheme(newTheme);
            setCodeTheme(resolvedTheme);
        };

        // Initial check
        updateUiTheme();

        // Watch for theme changes
        const observer = new MutationObserver(updateUiTheme);
        observer.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ["data-theme"],
        });

        return () => observer.disconnect();
    }, []);

    // Close dropdown when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        if (isOpen) {
            document.addEventListener("mousedown", handleClickOutside);
            return () => document.removeEventListener("mousedown", handleClickOutside);
        }
    }, [isOpen]);

    const handleSelect = (newTheme: CodeTheme) => {
        setCodeTheme(newTheme);
        // Save to UI-theme-specific storage key
        const storageKey = `codeTheme-${uiTheme}`;
        localStorage.setItem(storageKey, newTheme);
        document.documentElement.dataset.codeTheme = newTheme;
        setIsOpen(false);
    };

    const handleHover = (themeId: CodeTheme) => {
        // Apply theme preview on hover
        document.documentElement.dataset.codeTheme = themeId;
    };

    const handleHoverEnd = () => {
        // Revert to selected theme when hover ends
        document.documentElement.dataset.codeTheme = codeTheme;
    };

    // Filter themes based on UI theme
    const filteredThemes = Object.entries(THEMES).filter(([_, theme]) => theme.category === uiTheme);

    return (
        <div
            ref={dropdownRef}
            className='relative w-full rounded-lg border border-border-primary bg-bg-secondary px-3 py-2'>
            <div className='flex flex-col gap-2'>
                <span className='text-center font-medium text-sm text-text-primary'>Code Theme</span>
                <button
                    type='button'
                    onClick={() => setIsOpen((prev) => !prev)}
                    className='flex w-full items-center justify-between gap-1 rounded border border-border-primary bg-bg-primary px-2 py-1 text-text-primary text-xs transition-colors hover:bg-bg-tertiary focus:outline-none focus:ring-1 focus:ring-accent-blue'>
                    <span className='truncate'>{THEMES[codeTheme].label}</span>
                    <IoChevronDown className={`h-3 w-3 shrink-0 transition-transform ${isOpen ? "rotate-180" : ""}`} />
                </button>
            </div>

            {isOpen && (
                <div
                    className='scrollbar-gutter-stable absolute top-full right-0 left-0 z-50 mt-1 max-h-[70vh] overflow-y-scroll rounded-lg border border-border-primary bg-bg-primary shadow-lg'
                    style={{
                        scrollbarWidth: "auto",
                        scrollbarColor: "rgb(100 116 139) transparent",
                    }}
                    onWheel={(e) => e.stopPropagation()}>
                    {filteredThemes.map(([id, theme]) => (
                        <button
                            key={id}
                            type='button'
                            onClick={() => handleSelect(id as CodeTheme)}
                            onMouseEnter={() => handleHover(id as CodeTheme)}
                            onMouseLeave={handleHoverEnd}
                            className={`w-full px-3 py-2 text-left text-xs transition-colors hover:bg-bg-tertiary ${
                                codeTheme === id
                                    ? "bg-accent-blue/20 font-semibold text-accent-blue"
                                    : "text-text-primary"
                            }`}>
                            {theme.label}
                        </button>
                    ))}
                </div>
            )}
        </div>
    );
};
