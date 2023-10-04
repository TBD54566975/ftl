import { SVG, Element, Svg } from '@svgdotjs/svg.js'
import '@svgdotjs/svg.panzoom.js/dist/svg.panzoom.esm.js'

const centerGroup = ({
  canvas,
  selector,
  padding,
  width,
  height,
}: {
  canvas: Svg
  selector: `.${string}` | `#${string}`
  padding: number
  width: number
  height: number
}): [number, number, number, number] => {
  // Find the modules group with the class the id
  const group = canvas.findOne(selector) as Element

  // Get the group's bounding box in the global SVG coordinate system
  const BBox = group.rbox(canvas)

  // Calculate the scale factor to fit the group within the desired width and height
  const scaleX = width / BBox.width
  const scaleY = height / BBox.height
  const scale = Math.min(scaleX, scaleY)

  // Calculate dynamic padding factor based on overflow
  const overflowX = (BBox.width * scale) / width
  const overflowY = (BBox.height * scale) / height
  const paddingFactor = Math.max(overflowX, overflowY) + padding // Base padding of 10% + dynamic adjustment

  const newWidth = (width / scale) * paddingFactor
  const newHeight = (height / scale) * paddingFactor

  // Calculate the new viewbox coordinates to center the .graph group with padding
  const newX = BBox.cx - newWidth / 2
  const newY = BBox.cy - newHeight / 2

  // new viewbox
  return [newX, newY, newWidth, newHeight]
}

export const svgZoom = (svg: SVGSVGElement, width: number, height: number) => {
  // Create an SVG.js instance from the provided SVG element
  const canvas = SVG(svg)

  // Center Graph
  const viewbox = centerGroup({
    canvas,
    height,
    width,
    selector: '.graph',
    padding: 0.3,
  })

  canvas.viewbox(...viewbox)

  //@ts-ignore: lib types bad
  canvas.panZoom()

  return {
    to(id: string) {
      const viewbox = centerGroup({
        canvas,
        height,
        width,
        selector: `#${id}`,
        padding: 5.5,
      })
      canvas.animate().viewbox(...viewbox)
    },
    in() {
      const zoomLevel = canvas.zoom()
      canvas.animate().zoom(zoomLevel + 0.1) // Increase the zoom level by 0.1
    },
    out() {
      const zoomLevel = canvas.zoom()
      canvas.animate().zoom(zoomLevel - 0.1) // Decrease the zoom level by 0.1
    },
    reset() {
      const viewbox = centerGroup({
        canvas,
        height,
        width,
        selector: '.graph',
        padding: 0.3,
      })
      canvas.animate().viewbox(...viewbox)
    },
  }
}
