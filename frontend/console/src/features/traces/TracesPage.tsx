import { ArrowLeft02Icon } from 'hugeicons-react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { useRequestTraceEvents } from '../../api/timeline/use-request-trace-events'
import { Loader } from '../../components/Loader'
import { TraceDetails } from './TraceDetails'
import { TraceDetailsCall } from './details/TraceDetailsCall'
import { TraceDetailsIngress } from './details/TraceDetailsIngress'

export const TracesPage = () => {
  const navigate = useNavigate()

  const { requestKey } = useParams<{ requestKey: string }>()
  const requestEvents = useRequestTraceEvents(requestKey)
  const events = requestEvents.data?.reverse() ?? []

  const [searchParams] = useSearchParams()
  const eventIdParam = searchParams.get('event_id')
  const selectedEventId = eventIdParam ? BigInt(eventIdParam) : undefined

  if (events.length === 0) {
    return
  }

  if (requestKey === undefined) {
    return
  }

  const handleBack = () => {
    if (window.history.length > 1) {
      navigate(-1)
    } else {
      navigate('/modules')
    }
  }

  if (requestEvents.isLoading) {
    return (
      <div className='flex justify-center items-center min-h-screen'>
        <Loader />
      </div>
    )
  }

  const selectedEvent = events.find((event) => event.id === selectedEventId)
  let eventDetailsComponent: React.ReactNode
  switch (selectedEvent?.entry.case) {
    case 'call':
      eventDetailsComponent = <TraceDetailsCall event={selectedEvent} />
      break
    case 'ingress':
      eventDetailsComponent = <TraceDetailsIngress event={selectedEvent} />
      break
    default:
      eventDetailsComponent = <p>No details available for this event type.</p>
      break
  }

  return (
    <div className='flex h-full'>
      <div className='w-1/2 p-4 h-full overflow-y-auto'>
        <div className='flex items-center mb-2'>
          <button type='button' onClick={handleBack} className='flex items-center p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-600 cursor-pointer'>
            <ArrowLeft02Icon className='w-6 h-6' />
          </button>
          <span className='text-xl font-semibold ml-2'>Trace Details</span>
        </div>
        <TraceDetails requestKey={requestKey} events={events} selectedEventId={selectedEventId} />
      </div>

      <div className='my-4 border-l border-gray-100 dark:border-gray-700' />

      <div className='w-1/2 p-4 mt-1 h-full overflow-y-auto'>{eventDetailsComponent}</div>
    </div>
  )
}
