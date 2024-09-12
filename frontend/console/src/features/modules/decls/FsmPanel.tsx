import type { FSM } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'
import { RefLink } from './TypeEl'

export const FsmPanel = ({ value, moduleName, declName }: { value: FSM; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={false} comments={value.comments}>
        FSM: {moduleName}.{declName}
      </PanelHeader>
      <div className='mt-8'>
        Start{value.start.length === 1 ? '' : 's'}: {value.start.map((r, i) => [<RefLink key={i} r={r} />, i === value.start.length - 1 ? '' : ', '])}
      </div>
      <div className='mt-8'>
        Transitions
        <ul className='list-disc ml-8 text-sm'>
          {value.transitions.map((t, i) => (
            <li key={i} className='mt-2'>
              From <RefLink r={t.from} /> to <RefLink r={t.to} />
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
