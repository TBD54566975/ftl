module.exports = {
  env: {
    browser: true,
    es2021: true,
  },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended',
    'plugin:storybook/recommended',
  ],
  overrides: [
    {
      env: {
        node: true,
      },
      files: ['.eslintrc.{js,cjs}'],
      parserOptions: {
        sourceType: 'script',
      },
    },
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  plugins: ['@typescript-eslint', 'react'],
  rules: {
    'semi': ['error', 'never'],
    'quotes': ['error', 'single', { 'avoidEscape': true, 'allowTemplateLiterals': true }],
    'jsx-quotes': ['error', 'prefer-single'],
    'no-trailing-spaces': 'error',
    'no-multiple-empty-lines': ['error', { 'max': 1, 'maxEOF': 0, 'maxBOF': 0 }],
    'eol-last': ['error', 'always'],
    'max-len': ['error', { 'code': 120, 'tabWidth': 4, 'ignoreUrls': true, 'ignoreComments': false, 'ignoreRegExpLiterals': true, 'ignoreStrings': true, 'ignoreTemplateLiterals': true }],
    'react/react-in-jsx-scope': 'off',
    'react/jsx-uses-react': 'off',
    'func-style': ['error', 'expression'],
    'react/prop-types': 'off',
    '@typescript-eslint/consistent-type-definitions': ['error', 'interface'],
    '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
    '@typescript-eslint/ban-ts-comment': [
      'error',
      {
        'ts-ignore': 'allow-with-description',
      },
    ],
    'indent': ['error', 2, {
      'SwitchCase': 1,
      'VariableDeclarator': 1,
      'outerIIFEBody': 1,
      'MemberExpression': 1,
      'FunctionDeclaration': {
        'parameters': 1,
        'body': 1,
      },
      'FunctionExpression': {
        'parameters': 1,
        'body': 1,
      },
      'CallExpression': {
        'arguments': 1,
      },
      'ArrayExpression': 1,
      'ObjectExpression': 1,
      'ImportDeclaration': 1,
      'flatTernaryExpressions': false,
      'ignoreComments': false,
    }],
    'react/jsx-indent': ['error', 2],
    'react/jsx-indent-props': ['error', 2],
    'no-multi-spaces': ['error'],
  },
  settings: {
    react: {
      version: 'detect',
    },
  },
}
