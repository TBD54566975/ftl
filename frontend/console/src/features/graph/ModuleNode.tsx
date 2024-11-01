import { Handle, type NodeProps, Position } from 'reactflow'

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
  }
}

export const ModuleNode = ({ data }: Props) => {
  const handleColor = 'rgb(49 46 129)'
  const decls = []
  for (const d of data.item.configs) {
    decls.push(
      <div className='ml-4 text-xs'>
        config: {d.config.name}
      </div>
    )
  }
  for (const d of data.item.data) {
    decls.push(
      <div className='ml-4 text-xs'>
        data: {d.data.name}
      </div>
    )
  }
  for (const d of data.item.databases) {
    decls.push(
      <div className='ml-4 text-xs'>
        database: {d.database.name}
      </div>
    )
  }
  for (const d of data.item.enums) {
    decls.push(
      <div className='ml-4 text-xs'>
        enum: {d.enum.name}
      </div>
    )
  }
  for (const d of data.item.secrets) {
    decls.push(
      <div className='ml-4 text-xs'>
        secret: {d.secret.name}
      </div>
    )
  }
  for (const d of data.item.subscriptions) {
    decls.push(
      <div className='ml-4 text-xs'>
        subscription: {d.subscription.name}
      </div>
    )
  }
  for (const d of data.item.topics) {
    decls.push(
      <div className='ml-4 text-xs'>
        topic: {d.topic.name}
      </div>
    )
  }
  for (const d of data.item.typealiases) {
    decls.push(
      <div className='ml-4 text-xs'>
        typealias: {d.typealias.name}
      </div>
    )
  }
  for (const d of data.item.verbs) {
    decls.push(
      <div className='ml-4 text-xs'>
        verb: {d.verb.name}
      </div>
    )
  }
  return (
    <>
      <Handle type='target' position={Position.Left} style={{ border: 0, backgroundColor: handleColor }} isConnectable={true} />

      <div className={`grid h-full w-full bg-indigo-600 rounded-md ${data.selected ? 'bg-pink-600' : ''}`}>
        <div className='place-self-center text-xs text-gray-100 truncate max-w-[90%]'>{data.title}</div>
        {decls}
      </div>

      <Handle type='source' position={Position.Right} style={{ border: 0, backgroundColor: handleColor }} className='bg-indigo-600' isConnectable={true} />
    </>
  )
}
