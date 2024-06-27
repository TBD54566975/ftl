import SwiftCompilerPlugin
import SwiftSyntax
import SwiftSyntaxBuilder
import SwiftSyntaxMacros


/// Implementation of the `ftlVerb` macro,
public struct VerbMacro: PeerMacro {
   public static func expansion(
      of node: AttributeSyntax,
      providingPeersOf declaration: some DeclSyntaxProtocol,
      in context: some MacroExpansionContext
   ) throws -> [DeclSyntax] {
      // Only on functions at the moment.
      guard let funcDecl = declaration.as(FunctionDeclSyntax.self) else {
         throw MacroError(message:"@ftlVerb only works on functions")
      }
      
      // check parameters
      guard funcDecl.signature.parameterClause.parameters.count <= 1 else {
         throw MacroError(message: "too many parameters")
      }
      for parameter in funcDecl.signature.parameterClause.parameters {
         
      }
      
      
      return []
   }
}
