import type { IngressEvent } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { verbRefString } from '../verbs/verb.utils'

export const TimelineIngress = ({ ingress }: { ingress: IngressEvent }) => {
  const title = `Ingress ${ingress.method} ${ingress.path}`
  return (
    <span title={title}>
      {`${ingress.method} `}
      <span className='text-indigo-500 dark:text-indigo-300'>{ingress.path}</span>
      {` (${ingress.statusCode}) -> `}
      {ingress.verbRef?.module && <span className='text-indigo-500 dark:text-indigo-300'>{verbRefString(ingress.verbRef)}</span>}
    </span>
  )
}
