import SwiftCompilerPlugin
import SwiftSyntax
import SwiftSyntaxBuilder
import SwiftSyntaxMacros


/// Implementation of the `ftlVerb` macro,
public struct DataMacro: ExtensionMacro {
   public static func expansion(
      of node: AttributeSyntax,
      attachedTo declaration: some DeclGroupSyntax,
      providingExtensionsOf type: some TypeSyntaxProtocol,
      conformingTo protocols: [TypeSyntax],
      in context: some MacroExpansionContext
   ) throws -> [ExtensionDeclSyntax] {
      // Only on structs
      guard let structDecl = declaration.as(StructDeclSyntax.self) else {
         throw MacroError(message:"@ftlData only works on structs")
      }
      
      var variableSetters = [String]()
      var encoders = [String]()
      for m in structDecl.memberBlock.members {
         let member = MemberBlockItemSyntax(m)!
         if let decl = VariableDeclSyntax(member.decl) {
            for b in decl.bindings {
               let binding = PatternBindingSyntax(b)!
               
               guard let typeAnnotation = binding.typeAnnotation else {
                  throw MacroError(message: "could not find type for \(binding.pattern)")
               }
               //                    variableSetters.append("\(typeAnnotation.syntaxNodeType)")
               
               variableSetters.append("self.\(binding.pattern) = try \(typeAnnotation.type.description).ftlDecode(jsonDict[\"\(binding.pattern)\"])")
               encoders.append("output[\"\(binding.pattern)\"] = self.\(binding.pattern).ftlEncode()")
            }
         }
      }
      
      let equatableExtension = try ExtensionDeclSyntax(
"""
extension \(type.trimmed): FTLType {
    public static func ftlDecode(_ json:Any?) throws -> Self {
        return try \(type.trimmed)(ftlJson:json)
    }

    private init(ftlJson:Any?) throws {
        guard let jsonDict = ftlJson as? [String:Any] else {
            throw FTLError(message:"expected json object for \(Self.self) instead of \\(ftlJson)")
        }
        \(raw: variableSetters.joined(separator:"\n"))
    }

    public func ftlEncode() -> Any? {
        var output = [String:Any]()
        \(raw: encoders.joined(separator:"\n"))
        return output
    }
}
""")
      return [equatableExtension]
   }
}

