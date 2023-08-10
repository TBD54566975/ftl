import { expect } from '@jest/globals'

expect.extend({
  noInterdependencies(received?: { file: string, text: string, absolutePath: string }) {
    // This regexp checks for formatting
    if (received) {
      return {
        pass: false,
        message: () => 'Expected no component interdependencies but found: \n' + `'${received.text}' in ${received.file}`,
      }
    }
    return {
      pass: true,
      message: () => 'Expected not to receive any interdependencies local to the components package.',
    }
  },
})