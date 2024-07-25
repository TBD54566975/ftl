import { Duration, Timestamp } from '@bufbuild/protobuf'
import { CallEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { verbRefString } from '../verbs/verb.utils'

const CallBlock = ({
  call,
  selectedCall,
  firstTimeStamp,
  firstDuration,
}: {
  call: CallEvent
  selectedCall?: CallEvent
  firstTimeStamp: Timestamp
  firstDuration: Duration
}) => {
  const totalDurationMillis = (firstDuration.nanos ?? 0) / 1000000
  const durationInMillis = (call.duration?.nanos ?? 0) / 1000000
  let width = (durationInMillis / totalDurationMillis) * 100
  if (width < 1) {
    width = 1
  }

  const callTime = call.timeStamp?.toDate() ?? new Date()
  const initialTime = firstTimeStamp?.toDate() ?? new Date()
  const offsetInMillis = callTime.getTime() - initialTime.getTime()
  const leftOffsetPercentage = (offsetInMillis / totalDurationMillis) * 100

  const barColor = call.equals(selectedCall) ? 'bg-green-500' : 'bg-indigo-500'

  return (
    <div className='cursor-pointer group relative my-0.5 h-4 flex'>
      <div className='flex-grow relative'>
        <div
          className={`absolute h-4 ${barColor} rounded-sm`}
          style={{
            width: `${width}%`,
            left: `${leftOffsetPercentage}%`,
          }}
        />
      </div>

      <div className='text-gray-900 dark:text-gray-300 self-center text-xs p-1'>{`${durationInMillis}ms`}</div>
      {call.destinationVerbRef && (
        <span
          className='text-white pointer-events-none absolute pl-1 top-1/2 left-0 transform -translate-y-1/2
        self-center text-xs w-max opacity-0 transition-opacity group-hover:opacity-100'
        >
          {verbRefString(call.destinationVerbRef)}
        </span>
      )}
    </div>
  )
}

interface Props {
  calls: CallEvent[]
  call?: CallEvent
  setSelectedCall: React.Dispatch<React.SetStateAction<CallEvent>>
}

export const RequestGraph = ({ calls, call, setSelectedCall }: Props) => {
  if (calls.length === 0) {
    return <></>
  }

  const firstTimeStamp = calls[0].timeStamp
  const firstDuration = calls[0].duration
  if (firstTimeStamp === undefined || firstDuration === undefined) {
    return <></>
  }

  return (
    <div className='flex flex-col'>
      {calls.map((c, index) => (
        <div
          key={index}
          className='flex hover:bg-indigo-500/60 hover:dark:bg-indigo-500/10 rounded-sm'
          onClick={() => setSelectedCall(c)}
        >
          <div className='w-full relative'>
            <CallBlock call={c} selectedCall={call} firstTimeStamp={firstTimeStamp} firstDuration={firstDuration} />
          </div>
        </div>
      ))}
    </div>
  )
}
