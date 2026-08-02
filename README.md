# Sivanesh B — Terminal Résumé

An interactive, terminal-style portfolio inspired by [terminal.shop](https://terminal.shop).
Explore my résumé by **typing commands** instead of scrolling a webpage — complete with
command history, tab-completion, live suggestions, theming, a command palette, and a few
easter eggs. 🥚

> **Live prompt:** `sivanesh@portfolio:~$`

All content is generated from a single typed data file (`src/data/resume.ts`) that was
extracted from the original PDF résumé — the source of truth ships in `public/` and is
downloadable from within the terminal.

---

## ✨ Features

- **Real terminal feel** — monospace UI, blinking block cursor, fake shell prompt, CRT mode.
- **Command history** with <kbd>↑</kbd> / <kbd>↓</kbd> recall (persisted across reloads).
- **Tab autocomplete** for commands _and_ arguments (`theme `, `cat `, `man `, `sudo `).
- **Live command suggestions** as you type.
- **Command palette** — <kbd>Ctrl</kbd>/<kbd>⌘</kbd>+<kbd>K</kbd> fuzzy launcher.
- **Keyboard shortcuts** — <kbd>Ctrl</kbd>+<kbd>L</kbd> clear, <kbd>Ctrl</kbd>+<kbd>C</kbd>
  cancel, <kbd>Ctrl</kbd>+<kbd>U</kbd> kill line, <kbd>Ctrl</kbd>+<kbd>A</kbd>/<kbd>E</kbd> home/end.
- **Three themes + high-contrast** — `dark`, `green`, `amber`, plus an accessible HC toggle.
- **Animated welcome** — ASCII wordmark + typewriter intro (respects reduced-motion).
- **Search** the whole résumé (`search kubernetes`) with match highlighting.
- **Copy output** button on every command block; **download** & **print** the PDF.
- **Easter eggs** — `sudo hire-me`, `coffee`, `fortune`, `neofetch`, `matrix`, `hack`.
- **Accessible & responsive** — screen-reader landmarks, focus rings, keyboard-first, mobile-friendly.

---

## 🧰 Tech stack

| Concern      | Choice                                             |
| ------------ | -------------------------------------------------- |
| Framework    | React 18 + TypeScript (strict)                     |
| Build tool   | Vite 5                                             |
| Styling      | Tailwind CSS 3 (CSS-variable driven theming)       |
| Animation    | Framer Motion                                      |
| Icons        | lucide-react                                       |
| Terminal     | Custom emulator (no `xterm.js`) — see note below   |
| Data         | Static, strongly-typed `resume.ts` (no backend)    |

> **Why a custom terminal instead of `xterm.js`?** `xterm.js` renders to a canvas/grid, which
> makes clickable links, copy buttons, responsive wrapping, theming and screen-reader support
> awkward. A lightweight React emulator delivers the same UX with better accessibility, smaller
> JS, and full control over rendering — exactly the "simpler solution" the brief allows.

---

## ⌨️ Command reference

Type `help` in the app for the live list. Grouped overview:

| Category | Commands |
| -------- | -------- |
| **Info** | `about`, `whoami` |
| **Résumé** | `resume`, `experience`, `projects`, `skills`, `techstack`, `certifications`, `education`, `achievements`, `timeline`, `stats` |
| **Connect** | `contact`, `email`, `linkedin`, `github`, `blog` |
| **System** | `help`, `clear`, `history`, `search`, `theme`, `download`, `print`, `ls`, `cat`, `pwd`, `echo`, `date`, `banner`, `man`, `exit` |
| **Fun** | `coffee`, `fortune`, `neofetch`, `matrix`, `hack`, `sudo hire-me` |

Many commands have aliases (`cv`, `certs`, `edu`, `stack`, `cls`, `grep`, …). Try `man <command>`.

> ℹ️ The résumé PDF does not list a GitHub or blog, so `github` and `blog` honestly report that
> they're unavailable rather than inventing links. Add a GitHub handle any time in
> `src/data/resume.ts` under `contact.github`.

---

## 🚀 Getting started

**Prerequisites:** Node.js ≥ 18 and npm.

```bash
# install (uses a project-local cache — see "Isolated environment" below)
npm install

# start the dev server
npm run dev            # http://localhost:5173

# type-check + production build
npm run build

# preview the production build locally
npm run preview
```

### Isolated environment

This project is configured to be fully self-contained, virtualenv-style. `.npmrc` points npm's
download cache at `./.npm-cache` (instead of the shared `~/.npm`), and packages install to
`./node_modules` as usual. **Deleting the project folder removes everything** — nothing is
written system-wide.

---

## 🛠️ Customization

Everything visitors see comes from **one file**: [`src/data/resume.ts`](src/data/resume.ts).

- Edit your name, title, summary, skills, experience, projects, certs, education, timeline.
- `username` / `host` control the shell prompt (`<username>@<host>:~$`).
- Add optional links under `contact` (`github`, `portfolio`) — the relevant commands light up
  automatically.
- Replace `public/Sivanesh_B_Platform_DevOps_Engineer_Resume.pdf` and update `resumeFile`.

Themes live in [`src/index.css`](src/index.css) as CSS-variable palettes (`theme-dark`,
`theme-green`, `theme-amber`, `.hc`). Add another palette + list it in
[`src/hooks/useTheme.ts`](src/hooks/useTheme.ts) to create a new theme.

Add a command by exporting a `Command` object from a module in `src/commands/` and registering
it in [`src/commands/registry.ts`](src/commands/registry.ts).

---

## 📁 Project structure

```
src/
├─ commands/           # Command definitions + registry + autocomplete data
│  ├─ info.tsx         #   about, experience, skills, projects, stats, contact…
│  ├─ system.tsx       #   help, ls, cat, theme, search, download, print…
│  ├─ fun.tsx          #   sudo hire-me, coffee, fortune, neofetch, matrix, hack
│  ├─ registry.ts      #   aggregates commands + argument completions
│  └─ types.ts         #   Command / CommandContext / TerminalActions types
├─ components/
│  ├─ Terminal.tsx     #   orchestrator: execution, history, keys, sticky-scroll
│  ├─ InputLine.tsx    #   transparent input + block cursor + suggestions
│  ├─ HistoryBlock.tsx #   a rendered command + its output (+ copy button)
│  ├─ CommandPalette.tsx, MatrixRain.tsx, TopBar.tsx, Welcome.tsx …
│  └─ output/ui.tsx    #   reusable output primitives (Heading, KV, Bullet…)
├─ data/
│  ├─ resume.ts        #   ← single source of truth
│  └─ ascii/           #   ASCII banners (raw text imports)
├─ hooks/              # useTheme, useCommandHistory
├─ utils/              # autocomplete, search, formatting
└─ types/resume.ts     # ResumeData model
```

---

## ♿ Accessibility & ⚡ performance

- Semantic `<h1>` landmark, `role="log"` + `aria-live` output region, labelled input.
- Visible focus rings, full keyboard operation, `prefers-reduced-motion` + `prefers-contrast`
  honoured, high-contrast toggle, `<noscript>` fallback linking the PDF.
- Fonts are `preconnect`-ed with `display=swap`; heavy animation code is split into its own
  chunk; production JS is ~100 KB gzipped. Targets a Lighthouse score of 95+.

---

## ☁️ Deployment

### Vercel (zero-config)

1. Import the repo at [vercel.com/new](https://vercel.com/new).
2. Framework preset **Vite** is auto-detected — Build `npm run build`, Output `dist`.
3. Deploy. The default base (`/`) is correct for Vercel domains — no env vars needed.

### GitHub Pages (automated)

A workflow is included at [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml):

1. Push to `main`.
2. In **Settings → Pages**, set **Source = GitHub Actions**.
3. The workflow builds with `VITE_BASE=/<repo-name>/` and publishes `dist`.

Manual build for a sub-path (any static host):

```bash
VITE_BASE=/terminal-resume/ npm run build
```

---

## 📄 License

MIT — see the header of `package.json`. The résumé content and PDF belong to Sivanesh B.
