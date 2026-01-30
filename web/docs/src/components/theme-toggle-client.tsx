"use client";
import { useEffect, useState } from "react";

export type Theme = "light" | "dark";

export const ThemeToggle = () => {
    const [theme, setTheme] = useState<Theme>("dark");

    useEffect(() => {
        // Check for saved preference first
        const saved = localStorage.getItem("theme") as Theme | null;
        if (saved) {
            setTheme(saved);
            document.documentElement.dataset.theme = saved;
        } else {
            // Only use system preference if no saved preference exists
            const isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
            const systemTheme = isDark ? "dark" : "light";
            setTheme(systemTheme);
            document.documentElement.dataset.theme = systemTheme;
        }
    }, []);

    const toggleTheme = () => {
        const newTheme = theme === "dark" ? "light" : "dark";
        setTheme(newTheme);
        localStorage.setItem("theme", newTheme);
        document.documentElement.dataset.theme = newTheme;
    };

    return (
        <button
            type='button'
            onClick={toggleTheme}
            className='flex h-12 w-12 items-center justify-center rounded-xl border-2 border-border-primary bg-bg-secondary text-xl shadow-md transition-all duration-200 hover:scale-110 hover:border-border-secondary hover:shadow-lg'
            aria-label='Toggle theme'>
            {theme === "dark" ? "☀️" : "🌙"}
        </button>
    );
};
