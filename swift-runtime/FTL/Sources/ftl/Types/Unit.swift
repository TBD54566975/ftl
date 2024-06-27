import Foundation

public struct Unit: FTLType {
   public init() {}
   
   public static func ftlDecode(_ json:Any?) throws -> Self {
      // TODO: validate json
      return Unit()
   }
   
   public func ftlEncode() -> Any? {
      return nil
   }
}
