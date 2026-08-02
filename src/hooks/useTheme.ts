import { useCallback, useEffect, useState } from 'react';

export const THEMES = ['dark', 'green', 'amber'] as const;
export type ThemeName = (typeof THEMES)[number];

const THEME_KEY = 'tr:theme';
const HC_KEY = 'tr:hc';

function readTheme(): ThemeName {
  const stored = localStorage.getItem(THEME_KEY);
  return (THEMES as readonly string[]).includes(stored ?? '')
    ? (stored as ThemeName)
    : 'dark';
}

function readHc(): boolean {
  const stored = localStorage.getItem(HC_KEY);
  if (stored === 'true') return true;
  if (stored === 'false') return false;
  // Fall back to the OS preference the first time.
  return window.matchMedia?.('(prefers-contrast: more)').matches ?? false;
}

/**
 * Central theme controller. Applies the theme + high-contrast classes to
 * <html> and persists the choice so it survives reloads.
 */
export function useTheme() {
  const [theme, setThemeState] = useState<ThemeName>(readTheme);
  const [highContrast, setHighContrast] = useState<boolean>(readHc);

  useEffect(() => {
    const root = document.documentElement;
    root.classList.remove('theme-dark', 'theme-green', 'theme-amber');
    root.classList.add(`theme-${theme}`);
    root.classList.toggle('hc', highContrast);
    localStorage.setItem(THEME_KEY, theme);
    localStorage.setItem(HC_KEY, String(highContrast));
  }, [theme, highContrast]);

  const setTheme = useCallback((t: ThemeName) => setThemeState(t), []);

  const cycleTheme = useCallback(() => {
    setThemeState((cur) => THEMES[(THEMES.indexOf(cur) + 1) % THEMES.length]);
  }, []);

  const toggleHighContrast = useCallback(() => setHighContrast((v) => !v), []);

  return { theme, setTheme, cycleTheme, highContrast, toggleHighContrast };
}
