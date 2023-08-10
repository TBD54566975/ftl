const cp = require('child_process')
const chokidar = require('chokidar')

const run = scriptName => cp.spawn('npm', [ 'run', scriptName ], { stdio: 'inherit' })

run('build:css-types')
run('copy:css')
run('storybook')
run('test:watch')
run('build:playwright-test')


chokidar.watch('src/**/*.css').on('change', () => {
  run('build:css-types')
  run('copy:css')
})

chokidar.watch('src/**/*.stories.tsx').on('change', () => {
  run('build:playwright-test')
})
