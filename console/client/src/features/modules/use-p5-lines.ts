import React from 'react'
import P5 from 'p5'
import { Item } from './create-layout-data-structure'

interface SketchCleanup {
  cleanup: () => void
}

type VerbID = `${string}.${string}`
type VerbMap = Map<VerbID, number>

type VerbRects = {
  top: number
  right: number
  bottom: number
  left: number
  width: number
  height: number
}[]

const visualization = ({
  width,
  height,
  container,
  verbs,
  verbMap,
  callers,
}: {
  width: number
  height: number
  container: HTMLDivElement
  verbs: VerbRects
  verbMap: VerbMap
  callers: Map<VerbID, VerbID[]>
}): SketchCleanup => {
  const sketch = (p5: P5) => {
    p5.setup = () => {
      p5.createCanvas(width, height)
      console.log('hit')
    }
    // p5.draw = () => {
    // // Constants
    // const LINE_COLOR = 'rgb(67 56 202)'
    // const ROUNDED_CORNER_RADIUS = 10
    // const OFFSET = 20

    // // Helper function to draw an arrow
    // const drawArrow = (x1: number, y1: number, x2: number, y2: number) => {
    //   const ARROW_SIZE = 5
    //   p5.line(x1, y1, x2, y2)
    //   const angle = Math.atan2(y1 - y2, x1 - x2)
    //   p5.push()
    //   p5.translate(x2, y2)
    //   p5.rotate(angle - Math.PI / 2)
    //   p5.triangle(-ARROW_SIZE * 0.5, ARROW_SIZE, ARROW_SIZE * 0.5, ARROW_SIZE, 0, 0)
    //   p5.pop()
    // }

    // callers.forEach((targets, caller) => {
    //   const callerIndex = verbMap.get(caller) as number
    //   const callerRect = verbs[callerIndex]
    //   targets.forEach((target) => {
    //     const targetIndex = verbMap.get(target) as number
    //     const targetRect = verbs[targetIndex]

    //     // Determine draw direction
    //     const drawFromLeft = callerRect.left < targetRect.left
    //     const callerX = drawFromLeft ? callerRect.left : callerRect.right
    //     const targetX = drawFromLeft ? targetRect.left : targetRect.right

    //     // Calculate y mid-points
    //     const callerMidY = callerRect.top + callerRect.height / 2
    //     const targetMidY = targetRect.top + targetRect.height / 2

    //     // Check for intersections and adjust the x-coordinates if necessary
    //     let maxOffsetX = 0
    //     for (let i = Math.min(callerIndex, targetIndex) + 1; i < Math.max(callerIndex, targetIndex); i++) {
    //       maxOffsetX = Math.max(maxOffsetX, drawFromLeft ? verbs[i].left : verbs[i].right)
    //     }
    //     const adjustedCallerX = drawFromLeft ? callerX - maxOffsetX - OFFSET : callerX + maxOffsetX + OFFSET

    //     // Set line properties
    //     p5.stroke(LINE_COLOR)
    //     p5.strokeWeight(2)
    //     p5.noFill()

    //     // Draw lines
    //     p5.beginShape()
    //     p5.vertex(callerX, callerMidY)
    //     p5.quadraticVertex(adjustedCallerX, callerMidY, adjustedCallerX, targetMidY)
    //     p5.endShape()

    //     // Draw arrow from the last point to the target
    //     drawArrow(adjustedCallerX, targetMidY, targetX, targetMidY)
    //   })
    // })
    // }
  }

  const p5 = new P5(sketch, container)

  return {
    cleanup: p5.remove,
  }
}

export const useP5Lines = ({
  width,
  height,
  data,
  container,
}: {
  width: number
  height: number
  data: Item[]
  container?: HTMLDivElement
}) => {
  React.useEffect(() => {
    if (container) {
      const verbMap = new Map<VerbMap>()
      const verbs: VerbRects = []
      const callers = new Map<VerbID, VerbID[]>()
      let i = 0
      for (const module of data) {
        module.verbs.forEach((verb) => {
          const id: VerbID = `${module.name}.${verb.name}`
          const verbEl = (document.querySelector(`[data-id="${id}"]`) ||
            document.querySelector(`[data-id="${module.name}"]`)) as HTMLLIElement | HTMLButtonElement
          const rect = verbEl.getBoundingClientRect()
          verbMap.set(id, i)
          verbs.push(rect)
          if (verb.calls.length) {
            const targets = verb.calls.map((call): VerbID => `${call.module}.${call.name}`)
            callers.set(id, targets)
          }
          i++
        })
      }
      const { cleanup } = visualization({ callers, verbMap, verbs, container, width, height })

      return cleanup // This removes the canvas when the component is rerendered.
    }
  }, [container, width, height, data])
}
