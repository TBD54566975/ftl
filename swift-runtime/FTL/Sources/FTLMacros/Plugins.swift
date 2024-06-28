import SwiftCompilerPlugin
import SwiftSyntax
import SwiftSyntaxBuilder
import SwiftSyntaxMacros

@main
struct MacroPlugin: CompilerPlugin {
   let providingMacros: [Macro.Type] = [
      VerbMacro.self,
      DataMacro.self,
      CallVerbMacro.self,
      CallSourceMacro.self,
      CallSinkMacro.self,
      CallEmptyMacro.self,
   ]
}


