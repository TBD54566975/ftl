import { ArrowRight02Icon } from 'hugeicons-react'
import type React from 'react'
import { classNames } from '../utils'

const Body: React.FC<{
  className?: string
  style?: React.CSSProperties
  children?: React.ReactNode
}> = ({ className, style, children }) => {
  return (
    <div className={classNames(className, 'flex-1')} style={{ height: 'calc(100% - 44px)', ...style }}>
      {children}
    </div>
  )
}

export const Page: React.FC<{
  className?: string
  style?: React.CSSProperties
  children?: React.ReactNode
}> & {
  Body: typeof Body
} = ({ className, style, children }) => {
  return (
    <div className={classNames(className, 'flex flex-col h-full')} style={style}>
      {children}
    </div>
  )
}

Page.Body = Body
