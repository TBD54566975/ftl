import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useStatus } from '../../api/status/use-status'
import { Tabs } from '../../components/Tabs'
import { ControllersList } from './ControllersList'
import { DeploymentsList } from './DeploymentsList'
import { RoutesList } from './RoutesList'
import { RunnersList } from './RunnersList'

export const InfrastructurePage = () => {
  const status = useStatus()
  const [searchParams, setSearchParams] = useSearchParams()

  const [tabs, setTabs] = useState([
    { name: 'Controllers', id: 'controllers' },
    { name: 'Runners', id: 'runners' },
    { name: 'Deployments', id: 'deployments' },
    { name: 'Routes', id: 'routes' },
  ])

  useEffect(() => {
    if (!status.data) {
      return
    }

    setTabs((prevTabs) =>
      prevTabs.map((tab) => {
        switch (tab.id) {
          case 'controllers':
            return { ...tab, count: status.data.controllers.length }
          case 'runners':
            return { ...tab, count: status.data.runners.length }
          case 'deployments':
            return { ...tab, count: status.data.deployments.length }
          case 'routes':
            return { ...tab, count: status.data.routes.length }
          default:
            return tab
        }
      }),
    )
  }, [status.data])

  const handleTabClick = (tabId: string) => {
    setSearchParams({ tab: tabId })
  }

  const currentTab = searchParams.get('tab') || tabs[0].id
  const renderTabContent = () => {
    switch (currentTab) {
      case 'controllers':
        return <ControllersList />
      case 'runners':
        return <RunnersList />
      case 'deployments':
        return <DeploymentsList />
      case 'routes':
        return <RoutesList />
      default:
        return <></>
    }
  }

  return (
    <div className='px-6'>
      <Tabs tabs={tabs} initialTabId={currentTab} onTabClick={handleTabClick} />
      <div className='mt-4'>{renderTabContent()}</div>
    </div>
  )
}
