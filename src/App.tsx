import { useCallback, useEffect, useRef, useState } from 'react';
import { resume } from '@/data/resume';
import { useTheme } from '@/hooks/useTheme';
import { TopBar } from '@/components/TopBar';
import { Terminal, type TerminalHandle } from '@/components/Terminal';
import { CommandPalette } from '@/components/CommandPalette';
import { MatrixRain } from '@/components/MatrixRain';

const RESUME_URL = `${import.meta.env.BASE_URL}${resume.resumeFile}`;

export default function App() {
  const { theme, setTheme, cycleTheme, highContrast, toggleHighContrast } = useTheme();
  const [crt, setCrt] = useState(false);
  const [matrixActive, setMatrixActive] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const terminalRef = useRef<TerminalHandle>(null);

  const downloadResume = useCallback(() => {
    const a = document.createElement('a');
    a.href = RESUME_URL;
    a.download = resume.resumeFile;
    document.body.appendChild(a);
    a.click();
    a.remove();
  }, []);

  const printResume = useCallback(() => {
    const w = window.open(RESUME_URL, '_blank', 'noopener,noreferrer');
    if (!w) window.location.href = RESUME_URL;
  }, []);

  const exitMatrix = useCallback(() => {
    setMatrixActive(false);
    requestAnimationFrame(() => terminalRef.current?.focusInput());
  }, []);

  // Global command-palette shortcut: Ctrl/⌘ + K.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && (e.key === 'k' || e.key === 'K')) {
        e.preventDefault();
        setPaletteOpen((o) => !o);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  return (
    <div className={`flex h-full flex-col ${crt ? 'crt' : ''}`}>
      <TopBar
        theme={theme}
        setTheme={setTheme}
        highContrast={highContrast}
        toggleHighContrast={toggleHighContrast}
        onOpenPalette={() => setPaletteOpen(true)}
        onDownload={downloadResume}
      />

      <main className="min-h-0 flex-1">
        <Terminal
          ref={terminalRef}
          theme={theme}
          setTheme={setTheme}
          cycleTheme={cycleTheme}
          toggleHighContrast={toggleHighContrast}
          toggleCrt={() => setCrt((c) => !c)}
          onStartMatrix={() => setMatrixActive(true)}
          onDownload={downloadResume}
          onPrint={printResume}
        />
      </main>

      <CommandPalette
        open={paletteOpen}
        onClose={() => setPaletteOpen(false)}
        onRun={(cmd) => terminalRef.current?.runCommand(cmd)}
      />

      {matrixActive && <MatrixRain onExit={exitMatrix} />}
    </div>
  );
}
