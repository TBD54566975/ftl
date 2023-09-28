import { SVG } from '@svgdotjs/svg.js'
import '@svgdotjs/svg.panzoom.js/dist/svg.panzoom.esm.js'
import { vizID } from './constants'

export const svgZoom = () => {
  // enables panZoom
  const canvas = SVG(`#${vizID}`)
    //@ts-ignore: lib types bad
    ?.panZoom()
  const box = canvas.bbox()
  return {
    to(id: string) {
      const module = canvas.findOne(`#${id}`)
      //@ts-ignore: lib types bad
      const bbox = module?.bbox()
      if (bbox) {
        canvas.zoom(2, { x: bbox.x, y: bbox.y })
      }
    },
    in() {
      const zoomLevel = canvas.zoom()
      canvas.zoom(zoomLevel + 0.1) // Increase the zoom level by 0.1
    },
    out() {
      const zoomLevel = canvas.zoom()
      canvas.zoom(zoomLevel - 0.1) // Decrease the zoom level by 0.1
    },
    reset() {
      canvas.viewbox(box).zoom(1, { x: 0, y: 0 }) // Reset zoom level to 1 and pan to origin
    },
  }
}
