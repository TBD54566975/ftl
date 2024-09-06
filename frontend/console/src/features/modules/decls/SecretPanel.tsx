import type { Secret } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'
import { TypeEl } from './TypeEl'

export const SecretPanel = ({ value, moduleName, declName }: { value: Secret; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={false} comments={value.comments}>
        Secret: {moduleName}.{declName}
      </PanelHeader>
      <div className='text-sm my-4'>
        Type: <TypeEl t={value.type} />
      </div>
    </div>
  )
}
