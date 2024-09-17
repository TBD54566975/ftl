import type { Enum, Type, Value } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { classNames } from '../../../utils'
import { TypeEl } from './TypeEl'
import { enumType } from './enum.utils'

const VariantComments = ({ comments, fullRow }: { comments?: string[]; fullRow?: boolean }) => {
  if (!comments) {
    return
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
    return
  }
  const valueText = value?.value.case === 'intValue' ? v.toString() : `"${v}"`
  return (
    <div className='mb-2'>
      {name && `${name} = `}
      {valueText}
    </div>
  )
}

const VariantNameAndType = ({ name, t }: { name: string; t: Type }) => {
  return [
    <span key='n' className='mb-2'>
      {name}
    </span>,
    <TypeEl key='t' t={t} />,
  ]
}

const ValueEnumSnippet = ({ value }: { value: Enum }) => {
  return (
    <div>
      <div>
        {value.export ? 'export ' : ''}
        enum {value.name}: {`${enumType(value)} {`}
      </div>
      <div className='my-2 ml-8'>
        {value.variants.map((v) => [<VariantComments key='c' comments={v.comments} />, <VariantValue key='v' name={v.name} value={v.value} />])}
      </div>
      <div>{'}'}</div>
    </div>
  )
}

const TypeEnumSnippet = ({ value }: { value: Enum }) => {
  return (
    <div>
      <div>
        {value.export ? 'export ' : ''}
        enum {value.name}
        {' {'}
      </div>
      <div className='inline-grid grid-cols-2 gap-x-4 mt-2 ml-8' style={{ gridTemplateColumns: 'auto auto' }}>
        {value.variants.map((v) => [
          <VariantComments key='c' fullRow comments={v.comments} />,
          <VariantNameAndType key='n' name={v.name} t={v.value?.value.value?.value as Type} />,
        ])}
      </div>
      <div>{'}'}</div>
    </div>
  )
}

export const EnumSnippet = ({ value }: { value: Enum }) => {
  const isValueEnum = value.type !== undefined
  return <div className='text-xs font-mono'>{isValueEnum ? <ValueEnumSnippet value={value} /> : <TypeEnumSnippet value={value} />}</div>
}
