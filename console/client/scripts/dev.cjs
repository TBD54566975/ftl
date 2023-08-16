const cp = require('child_process')
const chokidar = require('chokidar')

const run = scriptName => cp.spawn('npm', [ 'run', scriptName ], { stdio: 'inherit' })

run('build:css-types')
run('copy:css')
run('test:watch')


chokidar.watch('src/**/*.css').on('change', () => {
  run('build:css-types')
  run('copy:css')
})


