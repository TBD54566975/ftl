import { instance } from '@viz-js/viz'

export const dot2Svg = async (dot: string) => {
  console.log(dot)
  const viz = await instance()
  try {
    return viz.renderSVGElement(dot)
  } catch (e) {
    console.error(e)
  }
}
