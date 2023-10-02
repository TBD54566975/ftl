import { ControlBar } from './ControlBar'
import React from 'react'
import { classNames } from '../utils'

interface Breadcrumb {
  label: string
  link?: string
}

interface Props {
  icon?: React.ReactNode
  title: string
  children?: React.ReactNode
  breadcrumbs?: Breadcrumb[]
  className?: string
}

export const PageHeader = ({ icon, title, children, breadcrumbs, className }: Props) => {
  return (
    <ControlBar className={classNames(className, `sticky top-0 z-10 justify-between`)}>
        <ControlBar.Icon>{icon}</ControlBar.Icon>
        {breadcrumbs && breadcrumbs.length > 0 && (
          <ControlBar.Breadcrumb data={breadcrumbs} />
        )}
        <ControlBar.Title>{title}</ControlBar.Title>
      {children}
    </ControlBar>
  )
}
