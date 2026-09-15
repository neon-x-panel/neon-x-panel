import { useMemo } from 'react';
import { useTheme as useNeonTheme } from '@/hooks/useTheme';

export type Theme = 'light' | 'dark' | 'system';
export type ResolvedTheme = 'light' | 'dark';
export type ColorTheme = 'default';

export function useTheme() {
  const { isDark, toggleTheme } = useNeonTheme();
  return useMemo(() => ({
    theme: (isDark ? 'dark' : 'light') as Theme,
    setTheme: () => toggleTheme(),
    resolvedTheme: (isDark ? 'dark' : 'light') as ResolvedTheme,
    colorTheme: 'default' as ColorTheme,
    setColorTheme: () => {},
    radius: 0.5,
    setRadius: () => {},
  }), [isDark, toggleTheme]);
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
