import { Handle, NodeProps, Position } from 'reactflow'

export const verbHeight = 40

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
  }
}

export const VerbNode = ({ data }: Props) => {
  const handleColor = 'rgb(49 46 129)'
  return (
    <>
      <Handle
        type='target'
        position={Position.Left}
        style={{ border: 0, backgroundColor: handleColor }}
        isConnectable={true}
      />

      <div className={`grid h-full w-full bg-indigo-600 rounded-md ${data.selected ? 'bg-pink-600' : ''}`}>
        <div className='place-self-center text-xs text-gray-100 truncate max-w-[90%]'>{data.title}</div>
      </div>

      <Handle
        type='source'
        position={Position.Right}
        style={{ border: 0, backgroundColor: handleColor }}
        className='bg-indigo-600'
        isConnectable={true}
      />
    </>
  )
}
