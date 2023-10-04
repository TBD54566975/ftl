import { instance } from '@viz-js/viz'

export const dotToSVG = async (dot: string): Promise<SVGSVGElement | undefined> => {
  const viz = await instance()
  try {
    return viz.renderSVGElement(dot)
  } catch (e) {
    console.error(e)
  }
}
