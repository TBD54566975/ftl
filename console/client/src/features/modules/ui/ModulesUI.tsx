import React from 'react'
import { ControlBar } from './ControlBar'
import { Sidebar } from './Sidebar'
import { classNames } from '../../../utils'

export const ModulesUI: React.FC<{
  withoutSidebarCls: string
  withSidebarCls: string
  controlBarCls: string
  className: string
  sidebarCls: string
}> = ({
  controlBarCls,
  withSidebarCls,
  withoutSidebarCls,
  sidebarCls,
  className
}) => {
  const sidebar = false
  return (
  <div className={classNames(className, sidebar && withSidebarCls, !sidebar && withoutSidebarCls)}>
    <ControlBar className={controlBarCls} />
    {sidebar && <Sidebar className={sidebarCls}/>}
  </div>
  )
}