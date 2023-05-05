import { MetadataCalls, Module, Verb } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { Card } from '../../components/Card'
import { Link } from 'react-router-dom'
import { classNames } from '../../utils/react.utils'

type Props = {
  module?: Module
  verb?: Verb
}

export const VerbCard: React.FC<Props> = ({ module, verb }) => {
  const calls = verb?.metadata
    .filter(meta => meta.value.case === 'calls')
    .map(meta => meta.value.value as MetadataCalls)

  return (
    <Card>
      <div className="min-w-0 flex-1">
        <Link to={`/modules/${module?.name}/verbs/${verb?.name}`} className="focus:outline-none">
          <p className="text-sm font-medium text-gray-900 dark:text-gray-300">{verb?.name}</p>
          {(calls?.length ?? 0) > 0 && (
            <li className="flex items-center space-x-4 pt-2">
              <div className="relative flex items-center space-x-4">
                <div
                  className={classNames(
                    'text-green-400 bg-green-400/10 ring-green-400/30',
                    'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset',
                  )}
                >
                  <code className="text-xs">{calls?.map(call => call.calls.map(call => call.name))}</code>
                </div>
              </div>
            </li>
          )}
        </Link>
      </div>
    </Card>
  )
}
