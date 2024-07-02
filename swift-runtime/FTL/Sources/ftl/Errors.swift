import Foundation

public struct FTLError: Error {
   public let message: String
   
   public init(message: String) {
      self.message = message
   }
}
