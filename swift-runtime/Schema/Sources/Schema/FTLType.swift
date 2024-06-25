import Foundation

public indirect enum FTLType {
    case unit
    case int
    case float
    case string
    case bytes
    case bool
    case time
    case array(FTLType)
    case dict(FTLType, FTLType)
    case any
    case optional(FTLType)
    // TODO: refs have associated types
    case ref(Ref)
    
    func toProto() -> Xyz_Block_Ftl_V1_Schema_Type {
        var proto = Xyz_Block_Ftl_V1_Schema_Type()
        switch self {
        case .unit:
            proto.value = .unit(Xyz_Block_Ftl_V1_Schema_Unit())
        case .int:
            proto.value = .int(Xyz_Block_Ftl_V1_Schema_Int())
        case .float:
            proto.value = .float(Xyz_Block_Ftl_V1_Schema_Float())
        case .string:
            proto.value = .string(Xyz_Block_Ftl_V1_Schema_String())
        case .bytes:
            proto.value = .bytes(Xyz_Block_Ftl_V1_Schema_Bytes())
        case .bool:
            proto.value = .bool(Xyz_Block_Ftl_V1_Schema_Bool())
        case .time:
            proto.value = .time(Xyz_Block_Ftl_V1_Schema_Time())
        case .array(let memberType):
            var array = Xyz_Block_Ftl_V1_Schema_Array()
            array.element = memberType.toProto()
            proto.value = .array(array)
        case .dict(let keyType, let valueType):
            var map = Xyz_Block_Ftl_V1_Schema_Map()
            map.key = keyType.toProto()
            map.value = valueType.toProto()
            proto.value = .map(map)
        case .any:
            proto.value = .any(Xyz_Block_Ftl_V1_Schema_Any())
        case .optional(let wrapped):
            var optional = Xyz_Block_Ftl_V1_Schema_Optional()
            optional.type = wrapped.toProto()
            proto.value = .optional(optional)
        case .ref(let ref):
            var protoRef = Xyz_Block_Ftl_V1_Schema_Ref()
            protoRef.module = ref.module
            protoRef.name = ref.name
            proto.value = .ref(protoRef)
        }
        return proto
    }
}
