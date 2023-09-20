import React from 'react'
import { Handle, Position } from 'reactflow'
import { Item } from './create-layout-data-structure'

interface Node {
  id: string
  data: Item
  type: string
  xPos: number
  yPos: number
  zIndex: number
  selected: boolean
  sourcePosition: string
  targetPosition: string
  dragging: boolean
  isConnectable: boolean
  dragHandle: string
  name: string
}

export const ModuleNode = React.memo(({ data }: { data: Item }) => {
  console.log(data)
  return (
    <div className='min-w-fit w-44 border border-gray-100 dark:border-slate-700 rounded overflow-hidden inline-block'>
      <button className='text-gray-600 dark:text-gray-300 p-1 w-full text-left flex justify-between items-center'>
        {data.name}
      </button>
      <ul className='text-gray-400 dark:text-gray-400 text-xs p-1 space-y-1 list-inside'>
        {data.verbs.map(({ name, id }) => (
          <li key={name} className='flex items-center text-gray-900 dark:text-gray-400 gap-x-2'>
            <Handle
              type='target'
              position={Position.Left}
              id={`${id}-target`}
              isConnectable={true}
              style={{ position: 'static' }}
            />
            {name}
            <Handle
              type='source'
              position={Position.Right}
              id={`${id}-source`}
              isConnectable={true}
              style={{ position: 'static', marginLeft: 'auto' }}
            />
          </li>
        ))}
      </ul>
    </div>
  )
})

ModuleNode.displayName === 'ModuleNode'
