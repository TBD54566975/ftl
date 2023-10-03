import { instance } from '@viz-js/viz'

export const dotToSVG = async (dot: string): Promise<[SVGSVGElement, number] | undefined> => {
  const viz = await instance()
  try {
    const svg =  viz.renderSVGElement(dot)
    const { width, height } = svg.viewBox.baseVal
    return [svg,  width/height]
  } catch (e) {
    console.error(e)
  }
}
