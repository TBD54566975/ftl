export default {
  setupFilesAfterEnv: ['<rootDir>/.jest/setup.js'],
  testEnvironment: 'jest-environment-jsdom',
  transformIgnorePatterns: [],
  transform: {
    '^.+\\.(t|j)sx?$': ['@swc/jest'],
  },
  moduleNameMapper: {
    '^(\\.{1,2}/.*)\\.js$': '$1',
    '\\.css$': 'identity-obj-proxy',
  },
  injectGlobals: false,
}
