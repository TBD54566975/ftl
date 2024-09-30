import { useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useModules } from '../../../api/modules/use-modules'
import { classNames } from '../../../utils'
import { Schema } from '../schema/Schema'
import { type DeclSchema, declFromModules } from '../schema/schema.utils'

const SnippetContainer = ({ decl, linkRect, containerRect }: { decl: DeclSchema; linkRect?: DOMRect; containerRect?: DOMRect }) => {
  const ref = useRef<HTMLDivElement>(null)
  const snipRect = ref?.current?.getBoundingClientRect()

  const hasRects = !!snipRect && !!linkRect
  const toTop = hasRects && window.innerHeight - linkRect.top - linkRect.height < snipRect.height + linkRect.height
  const fitsToRight = hasRects && window.innerWidth - linkRect.left >= snipRect.width
  const fitsToLeft = hasRects && !!containerRect && linkRect.left - containerRect.x + linkRect.width >= snipRect.width
  const horizontalAlignmentClassNames = fitsToRight ? '-ml-1' : fitsToLeft ? '-translate-x-full left-full ml-0' : ''
  const style = {
    transform: !fitsToRight && !fitsToLeft ? `translateX(-${(linkRect?.left || 0) - (containerRect?.left || 0)}px)` : undefined,
  }
  return (
    <div
      ref={ref}
      style={style}
      className={classNames(
        toTop ? 'bottom-full' : '',
        horizontalAlignmentClassNames,
        'absolute p-4 rounded-md border-solid border border border-gray-400 bg-gray-200 dark:border-gray-800 dark:bg-gray-700 text-gray-700 dark:text-white text-xs font-normal z-10 drop-shadow-xl cursor-default',
      )}
    >
      <Schema schema={decl.schema} containerRect={containerRect} />
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
  const navigate = useNavigate()
  const modules = useModules()
  const decl = useMemo(() => (moduleName ? declFromModules(moduleName, declName, modules) : undefined), [moduleName, declName, modules?.data])
  const [isHovering, setIsHovering] = useState(false)

  const str = moduleName && slim !== true ? `${moduleName}.${declName}` : declName

  if (!decl) {
    return str
  }

  const linkRef = useRef<HTMLSpanElement>(null)
  return (
    <span
      className='inline-block rounded-md cursor-pointer hover:bg-gray-400/30 hover:dark:bg-gray-900/30 p-1 -m-1 relative'
      onMouseEnter={() => setIsHovering(true)}
      onMouseLeave={() => setIsHovering(false)}
    >
      <span ref={linkRef} className={textColors} onClick={() => navigate(`/modules/${moduleName}/${decl.declType}/${declName}`)}>
        {str}
      </span>
      {!slim && isHovering && <SnippetContainer decl={decl} linkRect={linkRef?.current?.getBoundingClientRect()} containerRect={containerRect} />}
    </span>
  )
}
