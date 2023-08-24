import { Duration, Timestamp } from '@bufbuild/protobuf'
import { Call } from '../../protos/xyz/block/ftl/v1/console/console_pb'

interface CallBlockProps {
  call: Call;
  selectedCall?: Call;
  firstTimeStamp: Timestamp;
  firstDuration: Duration;
}

const CallBlock: React.FC<CallBlockProps> = ({ call, selectedCall, firstTimeStamp, firstDuration }) => {
  const totalDurationMillis = (firstDuration.nanos ?? 0) / 1000000
  const durationInMillis = (call.duration?.nanos ?? 0) / 1000000
  const width = (durationInMillis / totalDurationMillis) * 100

  const callTime = call.timeStamp?.toDate() ?? new Date()
  const initialTime = firstTimeStamp?.toDate() ?? new Date()
  const offsetInMillis = callTime.getTime() - initialTime.getTime()
  const leftOffsetPercentage = (offsetInMillis / totalDurationMillis) * 100

  const barColor = call.equals(selectedCall) ? 'bg-green-500' : 'bg-indigo-500'

  return (
    <div className='relative my-0.5 h-4 flex' title={`${call.destinationVerbRef?.module} : ${call.destinationVerbRef?.name}`}>
      <div className='flex-grow relative'>
        <div
          className={`absolute h-4 ${barColor} rounded-sm`}
          style={{
            width: `${width}%`,
            left: `${leftOffsetPercentage}%`,
          }}
        />
      </div>

      <div className='text-gray-900 dark:text-gray-300 self-center text-xs pl-2'>
        {`${durationInMillis}(ms)`}
      </div>
    </div>
  )
}

type Props = {
  calls: Call[]
  call?: Call
  setSelectedCall: React.Dispatch<React.SetStateAction<Call>>
}

export const RequestGraph: React.FC<Props> = ({ calls, call, setSelectedCall }) => {
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
        <div key={index} className='flex hover:bg-indigo-500/10' onClick={() => setSelectedCall(c)}>
          <div className='w-full relative'>
            <CallBlock
              call={c}
              selectedCall={call}
              firstTimeStamp={firstTimeStamp}
              firstDuration={firstDuration}
            />
          </div>
        </div>
      ))}
    </div>
  )
}

