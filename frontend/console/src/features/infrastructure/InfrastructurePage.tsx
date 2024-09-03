import { useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useStatus } from '../../api/status/use-status'
import { Tabs } from '../../components/Tabs'
import { ControllersList } from './ControllersList'
import { DeploymentsList } from './DeploymentsList'
import { RoutesList } from './RoutesList'
import { RunnersList } from './RunnersList'

export const InfrastructurePage = () => {
  const status = useStatus()
  const navigate = useNavigate()
  const location = useLocation()

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

  const currentTab = location.pathname.split('/').pop()

  const renderTabContent = () => {
    switch (currentTab) {
      case 'controllers':
        return <ControllersList controllers={status.data?.controllers || []} />
      case 'runners':
        return <RunnersList runners={status.data?.runners || []} />
      case 'deployments':
        return <DeploymentsList deployments={status.data?.deployments || []} />
      case 'routes':
        return <RoutesList routes={status.data?.routes || []} />
      default:
        return <></>
    }
  }

  const handleTabClick = (tabId: string) => {
    navigate(`/infrastructure/${tabId}`)
  }

  return (
    <div className='px-6'>
      <Tabs tabs={tabs} initialTabId={currentTab} onTabClick={handleTabClick} />
      <div className='mt-2'>{renderTabContent()}</div>
    </div>
  )
}
