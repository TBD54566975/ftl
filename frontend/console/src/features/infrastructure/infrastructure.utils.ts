import type { Value } from '@bufbuild/protobuf'

export const renderValue = (value: Value): string => {
  switch (value.kind?.case) {
    case 'numberValue':
      return value.kind.value.toString()
    case 'stringValue':
      return value.kind.value
    case 'boolValue':
      return value.kind.value ? 'true' : 'false'
    case 'structValue':
      return value.kind.value.toJsonString()
    case 'listValue':
      return value.kind.value.values.map(renderValue).join(', ')
    default:
      return ''
  }
}
