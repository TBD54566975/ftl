import Foundation

import SwiftSyntax
import SwiftSyntaxBuilder
import SwiftSyntaxMacros

func expandCall(node:some FreestandingMacroExpansionSyntax, verbType:String, includeRequest:Bool) throws -> ExprSyntax {
   // not sure how to access arguments by index
   var argumentList = [LabeledExprSyntax]()
   for a in node.argumentList {
      guard let argument = LabeledExprSyntax(a) else {
         throw MacroError(message: "expected LabeledExprSyntax")
      }
      argumentList.append(argument)
   }
   
   // verb
   let verbArgument = argumentList[0]
   var module: String?
   var name = ""
   if let expression = MemberAccessExprSyntax(verbArgument.expression) {
      module = expression.base?.trimmed.description.lowercased() ?? ""
      name = expression.declName.trimmed.description
   } else if let expression = DeclReferenceExprSyntax(verbArgument.expression) {
      name = expression.baseName.trimmed.description
   } else {
      throw MacroError(message: "expected MemberAccessExprSyntax for first argument")
   }
   
   if !includeRequest {
      return ExprSyntax("context.call(module:\"\(raw:module ?? "")\", name:\"\(raw:name)\", \(raw:verbType):\(raw:verbArgument.expression.description))")
   }
   
   // request
   let requestParameter = argumentList[1].expression.trimmed.description
   return ExprSyntax("context.call(module:\"\(raw:module ?? "")\", name:\"\(raw:name)\", \(raw:verbType):\(raw:verbArgument.expression.description), request:\(raw:requestParameter))")
}

public struct CallVerbMacro: ExpressionMacro {
   public static func expansion(
      of node: some FreestandingMacroExpansionSyntax,
      in context: some MacroExpansionContext
   ) throws -> ExprSyntax {
      return try expandCall(node: node, verbType: "verb", includeRequest: true)
   }
}

public struct CallSourceMacro: ExpressionMacro {
   public static func expansion(
      of node: some FreestandingMacroExpansionSyntax,
      in context: some MacroExpansionContext
   ) throws -> ExprSyntax {
      return try expandCall(node: node, verbType: "source", includeRequest: false)
   }
}

public struct CallSinkMacro: ExpressionMacro {
   public static func expansion(
      of node: some FreestandingMacroExpansionSyntax,
      in context: some MacroExpansionContext
   ) throws -> ExprSyntax {
      return try expandCall(node: node, verbType: "sink", includeRequest: true)
   }
}

public struct CallEmptyMacro: ExpressionMacro {
   public static func expansion(
      of node: some FreestandingMacroExpansionSyntax,
      in context: some MacroExpansionContext
   ) throws -> ExprSyntax {
      return try expandCall(node: node, verbType: "empty", includeRequest: false)
   }
}
