import Foundation

public struct FTLError: Error {
   let message: String
   
   public init(message: String) {
      self.message = message
   }
}
