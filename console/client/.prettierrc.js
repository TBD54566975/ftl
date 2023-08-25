export default {
  semi: true,
  singleQuote: true,
  jsxSingleQuote: true,
  trailingComma: 'es5',
  bracketSpacing: false,
  bracketSameLine: true,
  arrowParens: 'avoid',
  singleAttributePerLine: true,
  overrides: [
    {
      files: ['*.(t|j)sx?'],
      options: {
        parser: 'typescript',
      },
    },
  ],
};
