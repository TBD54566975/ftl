import type { Enum, Type, Value } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { classNames } from '../../../utils'
import { PanelHeader } from './PanelHeader'
import { TypeEl } from './TypeEl'

const VariantComments = ({ comments, fullRow }: { comments?: string[]; fullRow?: boolean }) => {
  if (!comments) {
    return []
  }
  return comments.map((c, i) => (
    <div key={i} className={classNames('text-gray-500 dark:text-gray-400 mb-0.5', fullRow ? 'col-start-1 col-end-3' : '')}>
      {c}
    </div>
  ))
}

const VariantValue = ({ name, value }: { name?: string; value?: Value }) => {
  const v = value?.value.value?.value
  if (v === undefined) {
    return []
  }
  const valueText = value?.value.case === 'intValue' ? v.toString() : `"${v}"`
  return (
    <div className='mb-3'>
      {name && `${name} = `}
      {valueText}
    </div>
  )
}

const VariantNameAndType = ({ name, t }: { name: string; t: Type }) => {
  return [
    <span key='n' className='mb-3'>
      {name}
    </span>,
    <TypeEl key='t' t={t} />,
  ]
}

const ValueEnumVariants = ({ value }: { value: Enum }) => {
  return value.variants.map((v) => [<VariantComments key='c' comments={v.comments} />, <VariantValue key='v' name={v.name} value={v.value} />])
}

const TypeEnumVariants = ({ value }: { value: Enum }) => {
  return (
    <div className='inline-grid grid-cols-2 gap-x-4' style={{ gridTemplateColumns: 'auto auto' }}>
      {value.variants.map((v) => [
        <VariantComments key='c' fullRow comments={v.comments} />,
        <VariantNameAndType key='n' name={v.name} t={v.value?.value.value?.value as Type} />,
      ])}
    </div>
  )
}

function enumType(value: Enum): string {
  if (!value.type) {
    return 'Type'
  }
  return value.type.value.case === 'string' ? 'String' : 'Int'
}

export const EnumPanel = ({ value, moduleName, declName }: { value: Enum; moduleName: string; declName: string }) => {
  const isValueEnum = value.type !== undefined
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        {enumType(value)} Enum: {moduleName}.{declName}
      </PanelHeader>
      <div className='mt-8'>
        <div className='mb-2'>Variants</div>
        <div className='text-xs font-mono'>{isValueEnum ? <ValueEnumVariants value={value} /> : <TypeEnumVariants value={value} />}</div>
      </div>
    </div>
  )
}
