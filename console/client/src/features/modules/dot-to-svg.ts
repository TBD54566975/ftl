import { instance } from '@viz-js/viz'

export const dotToSVG = async (dot: string) => {
  const viz = await instance()
  try {
    return viz.renderSVGElement(dot)
  } catch (e) {
    console.error(e)
  }
}
