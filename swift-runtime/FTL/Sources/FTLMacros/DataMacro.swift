import SwiftCompilerPlugin
import SwiftSyntax
import SwiftSyntaxBuilder
import SwiftSyntaxMacros


/// Implementation of the `ftlVerb` macro,
public struct DataMacro: MemberMacro {
    public static func expansion(
        of node: AttributeSyntax,
        providingMembersOf declaration: some DeclGroupSyntax,
        in context: some MacroExpansionContext
    ) throws -> [DeclSyntax] {
        // Only on functions at the moment.
        guard let _ = declaration.as(StructDeclSyntax.self) else {
            throw MacroError(message:"@ftlData only works on structs")
        }
        return []
    }
}

