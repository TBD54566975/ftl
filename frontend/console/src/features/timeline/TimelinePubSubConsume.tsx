import { PubSubConsumeEvent } from "../../protos/xyz/block/ftl/v1/console/console_pb";

export const TimelinePubSubConsume = ({ pubSubConsume }: { pubSubConsume: PubSubConsumeEvent }) => {
  var title = `Topic ${pubSubConsume.topic} propagated by controller`
  var consumedBy = undefined
  if (pubSubConsume.destVerbName) {
    consumedBy = `${pubSubConsume.destVerbModule && pubSubConsume.destVerbModule + '.' || ''}.${pubSubConsume.destVerbName}`
    title = `Topic ${pubSubConsume.topic} consumed by ${consumedBy}`
  }

  return (
    <span title={title}>
      {'Topic '}
      <span className='text-indigo-500 dark:text-indigo-300'>${pubSubConsume.topic}</span>
      {consumedBy &&
        <>
          {' consumed by '}
          <span className='text-indigo-500 dark:text-indigo-300'>{pubSubConsume.destVerbModule && pubSubConsume.destVerbModule + '.' || ''}{pubSubConsume.destVerbName || 'unknown'}</span>
        </>
      || ' propagated by controller'}
    </span>
  )
}
