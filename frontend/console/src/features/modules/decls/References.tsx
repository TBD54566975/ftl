import type { Ref } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { DeclLink } from './DeclLink'

export const References = ({ references }: { references: Ref[] }) => {
  return (
    <div className='mt-8 text-sm'>
      Referenced By:
      {references.length === 0
        ? ' None'
        : references.map((r, i) => (
            <div key={i} className='font-mono text-xs mt-2 ml-3.5'>
              <DeclLink moduleName={r.module} declName={r.name} />
            </div>
          ))}
    </div>
  )
}
