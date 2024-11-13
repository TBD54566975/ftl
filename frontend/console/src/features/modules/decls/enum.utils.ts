import type { Enum } from '../../../protos/xyz/block/ftl/schema/v1/schema_pb'

export function enumType(value: Enum): string {
  if (!value.type) {
    return 'Type'
  }
  return value.type.value.case === 'string' ? 'String' : 'Int'
}
