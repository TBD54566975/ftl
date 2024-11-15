import type { Ref } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { DeclLink } from './DeclLink'

export const References = ({ references }: { references: Ref[] }) => {
  return (
    <div className='text-sm space-y-2'>
      {references.length === 0
        ? ' None'
        : references.map((r, i) => (
            <div key={i} className='font-mono text-xs'>
              <DeclLink moduleName={r.module} declName={r.name} />
            </div>
          ))}
    </div>
  )
}
