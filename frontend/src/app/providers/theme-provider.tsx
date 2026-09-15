import { useMemo } from 'react';
import { useTheme as useNeonTheme } from '@/hooks/useTheme';

export type Theme = 'light' | 'dark' | 'system';
export type ResolvedTheme = 'light' | 'dark';
export type ColorTheme =
  | 'default'
  | 'blue'
  | 'green'
  | 'purple'
  | 'violet'
  | 'rose'
  | 'orange'
  | 'red'
  | 'yellow'
  | 'zinc'
  | 'neutral'
  | 'slate'
  | 'stone';

export function useTheme() {
  const { isDark, toggleTheme } = useNeonTheme();
  const setTheme = (_theme?: Theme) => toggleTheme();
  return useMemo(
    () => ({
      theme: (isDark ? 'dark' : 'light') as Theme,
      setTheme,
      resolvedTheme: (isDark ? 'dark' : 'light') as ResolvedTheme,
      colorTheme: 'default' as ColorTheme,
      setColorTheme: () => {},
      radius: 0.5,
      setRadius: () => {},
    }),
    [isDark, setTheme],
  );
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
