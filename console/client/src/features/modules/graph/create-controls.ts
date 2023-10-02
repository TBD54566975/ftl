import { svgZoom } from './svg-zoom'
import { controlIcons } from '../modules.constants'

export const createControls = (
  zoom: ReturnType<typeof svgZoom>,
): [Map<'in' | 'out' | 'reset', HTMLButtonElement>, () => void] => {
  const actions = ['in', 'out', 'reset'] as const
  const buttons: Map<(typeof actions)[number], HTMLButtonElement> = new Map()
  for (const action of actions) {
    const btn = document.createElement('button')
    btn.classList.add(
      ...'relative inline-flex items-center bg-white dark:hover:bg-indigo-700 dark:bg-gray-700/40 px-2 py-2 text-gray-500 dark:text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-10'.split(
        ' ',
      ),
    )
    const scr = document.createElement('span')
    scr.classList.add('sr-only')
    scr.innerText = action
    const parser = new DOMParser()
    const doc = parser.parseFromString(controlIcons[action], 'image/svg+xml')
    const svg = doc.documentElement
    btn.replaceChildren(scr, svg)
    btn.addEventListener('click', zoom[action])
    buttons.set(action, btn)
  }

  const removeEventListeners = () => {
    for (const action of actions) {
      buttons.get(action)?.removeEventListener('click', zoom[action])
    }
  }
  return [buttons, removeEventListeners]
}
