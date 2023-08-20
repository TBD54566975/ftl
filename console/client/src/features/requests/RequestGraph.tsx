import { Duration, Timestamp } from '@bufbuild/protobuf'
import { Call } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const colors = [ 'bg-indigo-500', 'bg-green-500/70', 'bg-pink-500/70', 'bg-blue-500/70', 'bg-yello-500/70' ]

interface CallBlockProps {
  call: Call;
  index: number;
  firstTimeStamp: Timestamp;
  firstDuration: Duration;
}

const CallBlock: React.FC<CallBlockProps> = ({ call, index, firstTimeStamp, firstDuration }) => {
  const totalDurationMillis = (firstDuration.nanos ?? 0) / 1000000
  const durationInMillis = (call.duration?.nanos ?? 0) / 1000000
  const width = (durationInMillis / totalDurationMillis) * 100

  const colorClass = colors[index % colors.length]

  const callTime = call.timeStamp?.toDate() ?? new Date()
  const initialTime = firstTimeStamp?.toDate() ?? new Date()
  const offsetInMillis = initialTime.getTime() - callTime.getTime()
  const leftOffsetPercentage = (offsetInMillis / totalDurationMillis) * 100

  return (
    <div className='relative my-1 h-4'>
      <div
        className={`absolute h-4 ${colorClass} rounded-md`}
        style={{
          width: `${width}%`,
          left: `${leftOffsetPercentage}%`,
        }}
        title={`Duration: ${call.duration}`}
      />
      <div
        className='absolute text-gray-900 right-0 top-1/2 transform -translate-y-1/2 text-xs pr-1'
      >
        {durationInMillis}ms
      </div>
    </div>
  )
}

type Props = {
  calls: Call[]
}

export const RequestGraph: React.FC<Props> = ({ calls }) => {
  if (calls.length === 0) {
    return <></>
  }

  const reversedCalls = calls.slice().reverse()
  const firstTimeStamp = reversedCalls[0].timeStamp
  const firstDuration = reversedCalls[0].duration
  if (firstTimeStamp === undefined || firstDuration === undefined) {
    return <></>
  }

  return (
    <div className='flex flex-col'>
      {reversedCalls.map((call, index) => (
        <div key={index}
          className='flex'
        >
          <div className='w-full relative'>
            <CallBlock call={call}
              index={index}
              firstTimeStamp={firstTimeStamp}
              firstDuration={firstDuration}
            />
          </div>
        </div>
      ))}
    </div>
  )
}


