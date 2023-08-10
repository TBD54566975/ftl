module.exports = {
  root: true,
  parserOptions: {
    ecmaVersion: 2022,
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
    'plugin:storybook/recommended',
  ],
  ignorePatterns: [ '**/dist/*' ],
  rules: {
    strict: 0,
    /* START: Prettier equivalent  rules */
    'max-len': [ 'error', {
      ignoreComments: true,
      tabWidth: 2, // Prettier:Tab Width
      ignoreTemplateLiterals: true,
      code: 120, // Prettier:Max length
    } ],
    'no-tabs': [ 'error' ], // Prettier: Tabs
    semi: [ 'error', 'never' ], // Prettier: Semicolons
    quotes: [ 'error', 'single', { // Prettier: Quotes
      avoidEscape: true,
      allowTemplateLiterals: true,
    } ],
    'quote-props': [ 'error', 'as-needed' ], // Prettier: Quote Props
    'comma-dangle': [ 'error', { // Prettier: Trailing Commas
      arrays: 'always-multiline',
      objects: 'always-multiline',
      imports: 'always-multiline',
      exports: 'always-multiline',
      functions: 'never',
    } ],
    'array-bracket-spacing': [ 'error', 'always' ], // Prettier: Bracket Spacing
    'object-curly-spacing': [ 'error', 'always' ],  // Prettier: Bracket Spacing
    'arrow-parens': [ 'error', 'as-needed' ], // Prettier: Arrow Function Parentheses
    'eol-last': [ 'error' ], // Prettier: End of Line
    // Prettier: JSX
    'jsx-quotes': [ 'error', 'prefer-single' ], // Prettier: JSX Quotes
    'react/jsx-tag-spacing': [ 'error' ], // https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-tag-spacing.md
    'react/jsx-closing-bracket-location': [ 'error', 'line-aligned' ], // https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-closing-bracket-location.md
    'react/jsx-indent-props': [ 2, 2 ], // https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-indent-props.md
    'react/jsx-max-props-per-line': [ 1 ], //https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-max-props-per-line.md
    /* END: Prettier equivalent  rules */

    /* Start: JS/TS quality rules */
    'react/jsx-key': [ 'error',
      { checkFragmentShorthand: true, checkKeyMustBeforeSpread: false, warnOnDuplicates: true },
    ], //https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-key.md
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
    /* End: JS/TS quality rules */

    /* START: Nested ternary specific rules*/
    indent: [ 'error', 2, {
      flatTernaryExpressions: true,
      ignoredNodes: [ 'ConditionalExpression' ],
      SwitchCase: 1,
    } ],
    'operator-linebreak': [ 'error', 'after', {
      overrides: {
        '?': 'before',
        ':': 'before',
      },
    } ],
    /* END: Nested ternary specific rules*/
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
