import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useStreamModules } from '../../../api/modules/use-stream-modules'
import { classNames } from '../../../utils'
import { getTreeWidthFromLS } from '../module.utils'
import { Schema } from '../schema/Schema'
import { type DeclSchema, declSchemaFromModules } from '../schema/schema.utils'

const topbarHeight = 64

const SnippetContainer = ({
  decl,
  moduleName,
  visible,
  linkRect,
  containerRect,
}: { decl: DeclSchema; moduleName: string; visible: boolean; linkRect?: DOMRect; containerRect?: DOMRect }) => {
  const [containerX, setContainerX] = useState(containerRect?.x || getTreeWidthFromLS())
  useEffect(() => {
    if (containerRect !== undefined) {
      setContainerX(containerRect.x)
    }
  }, [containerRect])

  const ref = useRef<HTMLDivElement>(null)
  const snipRect = ref?.current?.getBoundingClientRect()

  const hasRects = !!snipRect && !!linkRect
  const fitsAbove = hasRects && linkRect.top - topbarHeight > snipRect.height
  const fitsToRight = hasRects && window.innerWidth - linkRect.left >= snipRect.width
  const fitsToLeft = hasRects && linkRect.left - containerX + linkRect.width >= snipRect.width
  const horizontalAlignmentClassNames = fitsToRight ? '-ml-1' : fitsToLeft ? '-translate-x-full left-full ml-0' : ''
  const style = {
    transform: !fitsToRight && !fitsToLeft ? `translateX(-${(linkRect?.left || 0) - containerX}px)` : undefined,
  }
  return (
    <div
      ref={ref}
      style={style}
      className={classNames(
        fitsAbove ? 'bottom-full' : '',
        visible ? '' : 'invisible',
        horizontalAlignmentClassNames,
        'absolute p-4 pl-0.5 rounded-md border-solid border border border-gray-400 bg-gray-200 dark:border-gray-800 dark:bg-gray-700 text-gray-700 dark:text-white text-xs font-normal z-10 drop-shadow-xl cursor-default',
      )}
    >
      <Schema schema={decl.schema} moduleName={moduleName} containerRect={containerRect} />
    </div>
  )
}

// When `slim` is true, print only the decl name, not the module name, and show nothing on hover.
export const DeclLink = ({
  moduleName,
  declName,
  slim,
  textColors = 'text-indigo-600 dark:text-indigo-400',
  containerRect,
}: { moduleName?: string; declName: string; slim?: boolean; textColors?: string; containerRect?: DOMRect }) => {
  if (!moduleName || !declName) {
    return
  }
  const navigate = useNavigate()
  const modules = useStreamModules()
  const decl = useMemo(
    () => (moduleName && !!modules?.data ? declSchemaFromModules(moduleName, declName, modules?.data) : undefined),
    [moduleName, declName, modules?.data],
  )
  const [isHovering, setIsHovering] = useState(false)
  const linkRef = useRef<HTMLSpanElement>(null)

  const str = moduleName && slim !== true ? `${moduleName}.${declName}` : declName

  if (!decl) {
    return str
  }

  return (
    <span
      className='inline-block rounded-md cursor-pointer hover:bg-gray-400/30 hover:dark:bg-gray-900/30 p-1 -m-1 relative'
      onMouseEnter={() => setIsHovering(true)}
      onMouseLeave={() => setIsHovering(false)}
    >
      <span ref={linkRef} className={textColors} onClick={() => navigate(`/modules/${moduleName}/${decl.declType}/${declName}`)}>
        {str}
      </span>
      {!slim && (
        <SnippetContainer
          decl={decl}
          moduleName={moduleName}
          visible={isHovering}
          linkRect={linkRef?.current?.getBoundingClientRect()}
          containerRect={containerRect}
        />
      )}
    </span>
  )
}
