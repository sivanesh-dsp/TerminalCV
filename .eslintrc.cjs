module.exports = {
  root: true,
  env: { browser: true, es2020: true },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react-hooks/recommended',
  ],
  ignorePatterns: ['dist', '.eslintrc.cjs', 'vite.config.ts', 'postcss.config.js', 'tailwind.config.js'],
  parser: '@typescript-eslint/parser',
  plugins: ['react-refresh'],
  rules: {
    'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
  },
  overrides: [
    {
      // Command modules deliberately co-locate small render helpers with the
      // exported Command definitions, which is fine (these files aren't HMR
      // component boundaries). Silence the fast-refresh-only heuristic here.
      files: ['src/commands/**/*.tsx'],
      rules: { 'react-refresh/only-export-components': 'off' },
    },
  ],
};
