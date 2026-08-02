import { Command as CommandIcon, Contrast, Download } from 'lucide-react';
import { THEMES, type ThemeName } from '@/hooks/useTheme';
import { promptText } from '@/components/Prompt';

const SWATCH: Record<ThemeName, string> = {
  dark: 'bg-[#5ef2b0]',
  green: 'bg-[#7aff9e]',
  amber: 'bg-[#ffd166]',
};

interface TopBarProps {
  theme: ThemeName;
  setTheme: (t: ThemeName) => void;
  highContrast: boolean;
  toggleHighContrast: () => void;
  onOpenPalette: () => void;
  onDownload: () => void;
}

/** macOS-style window chrome + quick controls (theme, contrast, palette). */
export function TopBar({
  theme,
  setTheme,
  highContrast,
  toggleHighContrast,
  onOpenPalette,
  onDownload,
}: TopBarProps) {
  return (
    <header className="no-print flex items-center justify-between gap-2 border-b border-term-dim/25 bg-term-bg/80 px-3 py-2 backdrop-blur">
      <div className="flex items-center gap-3">
        <div className="flex gap-1.5" aria-hidden="true">
          <span className="h-3 w-3 rounded-full bg-[#ff5f56]" />
          <span className="h-3 w-3 rounded-full bg-[#ffbd2e]" />
          <span className="h-3 w-3 rounded-full bg-[#27c93f]" />
        </div>
        <span className="hidden text-xs text-term-dim sm:inline">{promptText} — /bin/zsh</span>
      </div>

      <div className="flex items-center gap-2">
        <div
          className="flex items-center gap-1 rounded-md border border-term-dim/30 p-0.5"
          role="group"
          aria-label="Theme"
        >
          {THEMES.map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTheme(t)}
              title={`${t} theme`}
              aria-label={`${t} theme`}
              aria-pressed={theme === t}
              className={`h-4 w-4 rounded-full ${SWATCH[t]} ring-offset-1 ring-offset-term-bg transition ${
                theme === t ? 'ring-2 ring-term-fg/70' : 'opacity-70 hover:opacity-100'
              }`}
            />
          ))}
        </div>

        <button
          type="button"
          onClick={toggleHighContrast}
          title="Toggle high contrast"
          aria-label="Toggle high contrast"
          aria-pressed={highContrast}
          className={`rounded-md border border-term-dim/30 p-1.5 transition-colors hover:text-term-accent ${
            highContrast ? 'text-term-accent' : 'text-term-dim'
          }`}
        >
          <Contrast size={15} />
        </button>

        <button
          type="button"
          onClick={onDownload}
          title="Download résumé PDF"
          aria-label="Download résumé PDF"
          className="rounded-md border border-term-dim/30 p-1.5 text-term-dim transition-colors hover:text-term-accent"
        >
          <Download size={15} />
        </button>

        <button
          type="button"
          onClick={onOpenPalette}
          title="Command palette (Ctrl/⌘+K)"
          aria-label="Open command palette"
          className="flex items-center gap-1 rounded-md border border-term-dim/30 px-2 py-1.5 text-xs text-term-dim transition-colors hover:text-term-accent"
        >
          <CommandIcon size={14} />
          <span className="hidden sm:inline">K</span>
        </button>
      </div>
    </header>
  );
}
