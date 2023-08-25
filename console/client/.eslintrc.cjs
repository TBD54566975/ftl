module.exports = {
  root: true,
  parserOptions: {
    ecmaVersgition: 2022,
    sourceType: 'module',
    allowImportExportEverywhere: true,
  },
  parser: '@typescript-eslint/parser',
  env: {
    browser: true,
    node: true,
    es6: true,
    mocha: true,
  },
  globals: {
    convertStoriesToJestCases: 'readonly',
  },
  plugins: [ 'react' ],
  extends: [
    'eslint:recommended',
    'plugin:compat/recommended',
    'plugin:@typescript-eslint/recommended',
    'prettier',
  ],
  ignorePatterns: [ '**/dist/*' ],
  rules: {
    strict: 0,
    /* Start: JS/TS quality rules */
    'react/jsx-key': [ 'error',
      { checkFragmentShorthand: true, checkKeyMustBeforeSpread: false, warnOnDuplicates: true },
    ], // https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-key.md
    'no-prototype-builtins': 0,
    'func-names': [ 'error', 'always', {
      generators: 'never',
    } ],
    'no-console': [ 'error', {
      allow: [ 'warn', 'error' ],
    } ],
    'no-return-assign': [ 'error' ],
    'class-methods-use-this': [ 'error', {
      exceptMethods: [ 'render' ],
    } ],
    'no-use-before-define': [ 'error' ],
    'no-unneeded-ternary': [ 'error' ],
    '@typescript-eslint/no-use-before-define': [ 'error' ],
    '@typescript-eslint/ban-ts-comment': [ 2, {
      'ts-ignore': 'allow-with-description',
    } ],
    '@typescript-eslint/no-unused-vars': [ 'warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' } ],
    /* End: JS/TS quality rules */
  },
  overrides: [
    {
      files: [ '**/*.spec.*' ],
      rules: {
        'func-names': 0,
        'no-console': 0,
        'no-shadow': 0,
        'no-return-assign': 0,
        'no-global-assign': 0,
        '@typescript-eslint/no-non-null-assertion': 0,
      },
    },
    {
      files: [ '**/*.svg.tsx' ],
      rules: {
        'max-len': 0,
      },
    },
    {
      files: [
        '**/*.cjs',
      ],
      rules: {
        '@typescript-eslint/no-var-requires': 0,
        'no-undef': 0,
        'func-names': 0,
      },
    },
  ],
}
