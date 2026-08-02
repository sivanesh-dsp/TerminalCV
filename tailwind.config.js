/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        mono: [
          'JetBrains Mono',
          'Fira Code',
          'SFMono-Regular',
          'Menlo',
          'Consolas',
          'Liberation Mono',
          'monospace',
        ],
      },
      colors: {
        // Theme-driven colors resolve from CSS custom properties so the
        // theme switcher (dark / green / amber / high-contrast) can swap
        // palettes at runtime without re-rendering the tree.
        term: {
          bg: 'rgb(var(--term-bg) / <alpha-value>)',
          fg: 'rgb(var(--term-fg) / <alpha-value>)',
          dim: 'rgb(var(--term-dim) / <alpha-value>)',
          accent: 'rgb(var(--term-accent) / <alpha-value>)',
          accent2: 'rgb(var(--term-accent2) / <alpha-value>)',
          prompt: 'rgb(var(--term-prompt) / <alpha-value>)',
          selection: 'rgb(var(--term-selection) / <alpha-value>)',
          error: 'rgb(var(--term-error) / <alpha-value>)',
          success: 'rgb(var(--term-success) / <alpha-value>)',
          warn: 'rgb(var(--term-warn) / <alpha-value>)',
          link: 'rgb(var(--term-link) / <alpha-value>)',
        },
      },
      keyframes: {
        blink: {
          '0%, 49%': { opacity: '1' },
          '50%, 100%': { opacity: '0' },
        },
        flicker: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.85' },
        },
        'fade-in': {
          from: { opacity: '0', transform: 'translateY(2px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
      },
      animation: {
        blink: 'blink 1s step-end infinite',
        flicker: 'flicker 3s ease-in-out infinite',
        'fade-in': 'fade-in 0.15s ease-out',
      },
    },
  },
  plugins: [],
};
