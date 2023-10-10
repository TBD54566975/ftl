import React from 'react'
import { classNames } from '../../../utils'
import { backgrounds, borders } from './components.constants'

const Header: React.FC<{
  className?: string
  children?: React.ReactNode
  style?: React.CSSProperties
}> = ({ className, children, style }) => {
  return (
    <div style={style} className={classNames('h-8 text-xs border-gray-300 dark:border-slate-700 font-bold', className)}>
      {children}
    </div>
  )
}

const Body: React.FC<{
  className?: string
  children?: React.ReactNode
  style?: React.CSSProperties
}> = ({ className, children, style }) => {
  return (
    <div style={{ height: 'calc(100% - 2rem)', ...style }} className={classNames(className, 'overflow-auto')}>
      {children}
    </div>
  )
}

export const Panel: React.FC & {
  Header: typeof Header
  Body: typeof Body
} = ({
  className,
  children,
  style,
}: {
  className?: string
  children?: React.ReactNode
  style?: React.CSSProperties
}) => {
  return (
    <div style={style} className={classNames(className, `p-2 overflow-hidden ${borders.level1} ${backgrounds.level1}`)}>
      {children}
    </div>
  )
}

Panel.Body = Body
Panel.Header = Header
