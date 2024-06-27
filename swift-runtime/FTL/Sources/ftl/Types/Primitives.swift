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
      throw FTLError(message: "not implemented")
   }
   
   public func ftlEncode() -> Any? {
      return nil
   }
}

extension Optional: FTLType {
   public static func ftlDecode(_ json:Any?) throws -> Self {
      throw FTLError(message: "not implemented")
   }
   
   public func ftlEncode() -> Any? {
      return nil
   }
}
