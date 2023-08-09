import { Handle, Position } from 'reactflow'

export function VerbNode({ data }) {
  return (
    <>
      <Handle
        type='target'
        position={Position.Top}
        style={{ border: 0 }}
        className='bg-indigo-600'
        isConnectable={true}
      />

      <div className='grid h-full w-full bg-indigo-600 rounded-md'>
        <div className='place-self-center text-xs text-gray-100'>{data.title}</div>
      </div>

      <Handle
        type='source'
        position={Position.Bottom}
        style={{ border: 0 }}
        className='bg-indigo-600'
        isConnectable={true}
      />
    </>
  )
}
