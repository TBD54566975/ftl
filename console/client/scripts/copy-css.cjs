const fg = require('fast-glob')
const path = require('path')
const fs = require('fs/promises')

;(async () => {
  const src = path.resolve(__dirname, '../src')
  const files = await fg([`${src}/**/*.css`, `${src}/**/*.css.d.ts`], {
    dot: true,
  })
  const dist = path.resolve(__dirname, '../dist')
  await fs.mkdir(dist, {recursive: true})
  await Promise.all(
    files.map(async file => {
      try {
        await fs.copyFile(file, file.replace(src, dist))
        // eslint-disable-next-line no-console
        console.log(`${file.replace(src, '')} was copied to dist`)
      } catch {
        console.error(` ${file.replace(src, '')}could not be copied`)
      }
    })
  )
})()
