import Foundation

public enum Decl {
    case verb(Verb)
    case dataStruct(DataStruct)
    
    
    func toProto() -> Xyz_Block_Ftl_V1_Schema_Decl {
        var out = Xyz_Block_Ftl_V1_Schema_Decl()
        switch self {
        case .verb(let verb):
            out.value = .verb(verb.toProto())
        case .dataStruct(let dataStruct):
            out.value = .data(dataStruct.toProto())
        }
        return out
    }
}
