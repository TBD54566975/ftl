import Foundation

extension String: FTLType {
   public static func ftlDecode(_ json:Any?) throws -> Self {
      guard let jsonString = json as? String else {
         throw FTLError(message: "expected string for string")
      }
      return jsonString
   }
   
   public func ftlEncode() -> Any? {
      return self
   }
}

extension Date: FTLType {
   public static func ftlDecode(_ json:Any?) throws -> Self {
      guard let raw = json as? String else {
         throw FTLError(message:"expected date to be represented by a json string instead of \(json)")
      }
      return try Date(raw, strategy:.iso8601)
   }
   
   public func ftlEncode() -> Any? {
      return self.formatted(.iso8601)
   }
}

extension Optional: FTLType where Wrapped : FTLType {
   public static func ftlDecode(_ json:Any?) throws -> Self {
      guard let json = json else {
         return .none
      }
      if let _ = json as? NSNull {
         return .none
      }
      return .some(try Wrapped.ftlDecode(json))
   }
   
   public func ftlEncode() -> Any? {
      switch self {
      case .none:
         return nil
      case .some(let wrapped):
         return wrapped.ftlEncode()
      }
   }
}
