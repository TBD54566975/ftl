import { CodeCircleIcon } from 'hugeicons-react'
import { Handle, type NodeProps, Position } from 'reactflow'

export const declHeight = 32

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
    icon?: React.ComponentType<{ className?: string }>
  }
}

export const DeclNode = ({ data }: Props) => {
  const handleColor = data.selected ? 'rgb(251 113 133)' : 'rgb(79 70 229)'
  const Icon = data.icon || CodeCircleIcon
  return (
    <>
      <Handle type='target' position={Position.Left} style={{ border: 0, backgroundColor: handleColor }} isConnectable={true} />

      <div className={`flex h-full w-full bg-indigo-600 rounded-md ${data.selected ? 'bg-pink-600' : ''}`}>
        <div className='flex items-center text-gray-200 px-2 gap-1 w-full overflow-hidden'>
          <Icon className='size-4 flex-shrink-0' />
          <div className='text-xs truncate min-w-0 flex-1'>{data.title}</div>
        </div>
      </div>

      <Handle type='source' position={Position.Right} style={{ border: 0, backgroundColor: handleColor }} isConnectable={true} />
    </>
  )
}
