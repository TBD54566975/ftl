import React from "react";
import { classNames } from "../../../utils";

const Header: React.FC<{
  className?: string,
  children?: React.ReactNode
  style?: React.CSSProperties
}> = ({
  className,
  children,
  style
}) => {
  return (
    <div
      style={style}
      className={classNames(
      'p-1 text-xs border-b border-gray-300 dark:border-slate-700 font-bold',
      className
    )}>
      {children}
    </div>
  )
}

const Body: React.FC<{
  className?: string,
  children?: React.ReactNode
  style?: React.CSSProperties
}> = ({
  className,
  children,
  style
}) => {
  return (
    <div
      style={style}
      className={classNames(className, 'flex-1 overflow-auto')}
    >
        {children}
    </div>
  )
}

export const Panel: React.FC<{
  className?: string,
  children?: React.ReactNode
  style?: React.CSSProperties
}> & {
  Header: typeof Header
  Body: typeof Body
} = ({
  className,
  children,
  style
}) =>  {
  return (
    <div
      style={style}
      className={classNames(className, 'flex flex-col p-2 rounded')}
    >
      {children}
    </div>
  )
}

Panel.Body = Body
Panel.Header = Header