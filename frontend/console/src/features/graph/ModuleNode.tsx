import { Handle, type NodeProps, Position } from 'reactflow'
import type { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'

interface Props extends NodeProps {
  data: {
    title: string
    item: Module
    selected: boolean
  }
}

export const ModuleNode = ({ data }: Props) => {
  const handleColor = 'rgb(49 46 129)'
  /*const decls = []
  for (const d of data.item.configs) {
    decls.push(<div key={d.config?.name || 'd'} className='ml-4 text-xs'>config: {d.config?.name}</div>)
  }
  for (const d of data.item.data) {
    decls.push(<div key={d.data?.name || 'd'} className='ml-4 text-xs'>data: {d.data?.name}</div>)
  }
  for (const d of data.item.databases) {
    decls.push(<div key={d.database?.name || 'd'} className='ml-4 text-xs'>database: {d.database?.name}</div>)
  }
  for (const d of data.item.enums) {
    decls.push(<div key={d.enum?.name || 'd'} className='ml-4 text-xs'>enum: {d.enum?.name}</div>)
  }
  for (const d of data.item.secrets) {
    decls.push(<div key={d.secret?.name || 'd'} className='ml-4 text-xs'>secret: {d.secret?.name}</div>)
  }
  for (const d of data.item.subscriptions) {
    decls.push(<div key={d.subscription?.name || 'd'} className='ml-4 text-xs'>subscription: {d.subscription?.name}</div>)
  }
  for (const d of data.item.topics) {
    decls.push(<div key={d.topic?.name || 'd'} className='ml-4 text-xs'>topic: {d.topic?.name}</div>)
  }
  for (const d of data.item.typealiases) {
    decls.push(<div key={d.typealias?.name || 'd'} className='ml-4 text-xs'>typealias: {d.typealias?.name}</div>)
  }
  for (const d of data.item.verbs) {
    decls.push(<div key={d.verb?.name || 'd'} className='ml-4 text-xs'>verb: {d.verb?.name}</div>)
  }*/
  return (
    <>
      <Handle type='target' position={Position.Left} style={{ border: 0, backgroundColor: handleColor }} isConnectable={true} />

      <div className={`justify-center grid h-full w-full bg-indigo-600 rounded-md ${data.selected ? 'bg-pink-600' : ''}`}>
        <div className='mt-1 text-xs text-gray-100 truncate'>{data.title}</div>
      </div>

      <Handle type='source' position={Position.Right} style={{ border: 0, backgroundColor: handleColor }} className='bg-indigo-600' isConnectable={true} />
    </>
  )
}
