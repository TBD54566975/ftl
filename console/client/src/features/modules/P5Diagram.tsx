import React from 'react'
import P5 from 'p5'

interface SketchCleanup {
  cleanup: () => void
}

const visualisation = ({ width, height, id }: { width: number; height: number; id: string }): SketchCleanup => {
  const sketch = (p5: P5) => {
    p5.setup = () => {
      p5.createCanvas(width, height)
    }
    p5.draw = () => {
      p5.line(0, 0, width, height)
    }
  }

  const p5 = new P5(sketch, id as unknown as HTMLElement)

  return {
    cleanup: p5.remove,
  }
}

export const P5Diagram = ({ width, height }: { width: number; height: number }) => {
  const id = React.useId()

  React.useEffect(() => {
    const { cleanup } = visualisation({
      id,
      width,
      height,
    })

    return cleanup // This removes the canvas when the component is rerendered.
  }, [])

  return <div id={id} />
}
