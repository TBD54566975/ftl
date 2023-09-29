import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card } from '../../components/Card'
import { PageHeader } from '../../components/PageHeader'
import { CallEvent, Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { SidePanelContext } from '../../providers/side-panel-provider'
import { getCalls } from '../../services/console.service'
import { formatTimestampShort } from '../../utils/date.utils'
import { TimelineCallDetails } from '../timeline/details/TimelineCallDetails'
import { verbRefString } from '../verbs/verb.utils'

export const ModulePage = () => {
  const navigate = useNavigate()
  const { moduleName } = useParams()
  const { openPanel, closePanel } = React.useContext(SidePanelContext)
  const modules = React.useContext(modulesContext)
  const [module, setModule] = React.useState<Module | undefined>()
  const [calls, setCalls] = React.useState<CallEvent[] | undefined>()
  const [selectedCall, setSelectedCall] = React.useState<CallEvent | undefined>()

  React.useEffect(() => {
    if (modules) {
      const module = modules.modules.find((module) => module.name === moduleName?.toLocaleLowerCase())
      setModule(module)
    }
  }, [modules, moduleName])

  React.useEffect(() => {
    if (!module) return

    const fetchCalls = async () => {
      const calls = await getCalls(module.name)
      setCalls(calls)
    }
    fetchCalls()
  }, [module])

  const handleCallClicked = (call: CallEvent) => {
    if (selectedCall?.equals(call)) {
      setSelectedCall(undefined)
      closePanel()
      return
    }
    setSelectedCall(call)
    openPanel(<TimelineCallDetails timestamp={call.timeStamp!} call={call} />)
  }

  return (
    <>
      <div className='flex-1 flex flex-col h-full'>
        <div className='flex-1'>
          <PageHeader
            icon={<Square3Stack3DIcon />}
            title={module?.name || ''}
            breadcrumbs={[{ label: 'Modules', link: '/modules' }]}
          />
          <div className='flex-1 m-4'>
            <div className='grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4'>
              {module?.verbs.map((verb) => (
                <Card
                  key={verb.verb?.name}
                  topBarColor='bg-green-500'
                  onClick={() => navigate(`/modules/${module.name}/verbs/${verb.verb?.name}`)}
                >
                  {verb.verb?.name}
                  <p className='text-xs text-gray-400'>{verb.verb?.name}</p>
                </Card>
              ))}
            </div>
          </div>
        </div>
        <div className='flex-1 relative m-4'>
          <div className='border border-gray-100 dark:border-slate-700 rounded h-full absolute inset-0'>
            <table className={`w-full table-fixed text-gray-600 dark:text-gray-300`}>
              <thead>
                <tr className='flex text-xs'>
                  <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none sticky top-0 z-10'>
                    Date
                  </th>
                  <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none sticky top-0 z-10'>
                    Source
                  </th>
                  <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none sticky top-0 z-10'>
                    Destination
                  </th>
                  <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 flex-1 flex-grow sticky top-0 z-10'>
                    Request
                  </th>
                  <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 flex-1 flex-grow sticky top-0 z-10'>
                    Response
                  </th>
                  <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 flex-1 flex-grow sticky top-0 z-10'>
                    Error
                  </th>
                </tr>
              </thead>
            </table>
            <div className='overflow-y-auto h-[calc(100%-2rem)]'>
              <table className={`w-full table-fixed text-gray-600 dark:text-gray-300`}>
                <tbody className='text-xs'>
                  {calls?.map((call, index) => (
                    <tr
                      key={`${index}-${call.timeStamp?.toDate().toUTCString()}`}
                      className={`border-b border-gray-100 dark:border-slate-700 font-roboto-mono ${
                        selectedCall?.equals(call) ? 'bg-indigo-50 dark:bg-slate-700' : ''
                      } relative flex cursor-pointer hover:bg-indigo-50 dark:hover:bg-slate-700`}
                      onClick={() => handleCallClicked(call)}
                    >
                      <td className='p-1 w-40 items-center flex-none text-gray-400 dark:text-gray-400'>
                        {formatTimestampShort(call.timeStamp)}
                      </td>
                      <td className='p-1 w-40 flex-none text-indigo-500 dark:text-indigo-300'>
                        {call.sourceVerbRef && verbRefString(call.sourceVerbRef)}
                      </td>
                      <td className='p-1 w-40 flex-none text-indigo-500 dark:text-indigo-300'>
                        {call.destinationVerbRef && verbRefString(call.destinationVerbRef)}
                      </td>
                      <td className='p-1 flex-1 flex-grow truncate' title={call.request}>
                        {call.request}
                      </td>
                      <td className='p-1 flex-1 flex-grow truncate' title={call.response}>
                        {call.response}
                      </td>
                      <td className='p-1 flex-1 flex-grow truncate' title={call.error}>
                        {call.error}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
