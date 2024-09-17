import type { Enum } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'

export function enumType(value: Enum): string {
  if (!value.type) {
    return 'Type'
  }
  return value.type.value.case === 'string' ? 'String' : 'Int'
}
