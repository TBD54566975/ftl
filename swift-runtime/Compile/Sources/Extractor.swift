import Foundation
import SwiftSyntax
import SwiftParser
import Schema

/// Extractor takes a module directory and extracts an FTL schema
public class Extractor {
    struct ExtractorError: Error {
        let message: String
        let wrappedError: Error?
        
        init(message: String, wrappedError:Error? = nil) {
            self.message = message
            self.wrappedError = wrappedError
        }
    }
    
    let rootURL: URL
    let deployURL: URL
    let name: String

    private var module: Module
    private var errors = [Error]()

    
    public init(rootURL: URL, deployURL: URL, name:String) {
        self.rootURL = rootURL
        self.deployURL = deployURL
        self.name = name
        self.module = Module(name: name)
    }
    
    public func extract() throws -> Module {
        let fileManager = FileManager.default
        let resourceKeys: [URLResourceKey] = [.isDirectoryKey]
        
        var fileError:Error?
        let enumerator = fileManager.enumerator(at: self.rootURL, includingPropertiesForKeys: resourceKeys, options: [.skipsHiddenFiles]) { (url, error) -> Bool in
            fileError = error
            return false
        }
        if let fileError = fileError {
            throw fileError
        }
        
        while let fileURL = enumerator?.nextObject() as? URL {
            if fileURL.pathExtension != "swift" {
                continue
            }
            if fileURL.path().hasSuffix("Tests.swift") {
                continue
            }
            if fileURL.absoluteString.hasPrefix(self.deployURL.absoluteString) {
                // TODO: make sure deploy url ends in /
                continue
            }
            let resourceValues = try fileURL.resourceValues(forKeys: Set(resourceKeys))
            if resourceValues.isDirectory == true {
                continue
            }
            try processSwiftFile(fileURL)
        }
        self.resolveVisibility()
        
        return module
    }
    
    func processSwiftFile(_ url:URL) throws {
        let data = try Data(contentsOf: url)
        guard let code = String(data: data, encoding: .utf8) else {
            throw ExtractorError(message: "Could not read code at \(url)")
        }
        let rootNode = Parser.parse(source: code)
        
        extractDataStructs(Syntax(rootNode))
        extractVerbs(Syntax(rootNode))
    }
    
    // MARK: - Extract specific declarations
    
    func extractVerbs(_ root: Syntax) {
        traverseTree(root) { node in
            guard let funcNode = node.as(FunctionDeclSyntax.self) else {
                return
            }
            extractVerb(funcNode)
        }
    }
    
    func extractVerb(_ node: FunctionDeclSyntax) {
        var errors = [Error]()
        let requestType = {
            for (_, parameter) in node.signature.parameterClause.parameters.enumerated() {
                // TODO: check param index
                do {
                    return try self.ftlTypeFrom(parameter.type)
                }
                catch {
                    errors.append(ExtractorError(message: "Could not determine verb request type", wrappedError:error))
                }
            }
            return FTLType.unit
        }()
        
        let responseType = {
            if let returnNode = node.signature.returnClause {
                do {
                    return try self.ftlTypeFrom(returnNode.type)
                }
                catch {
                    // TODO: handle type error
                }
            }
            return FTLType.unit
        }()
        
        var verb: Verb?
        // var metas = [Metadata]()
        errors.append(contentsOf: self.visitAttributesIn(list: node.attributes) { attributeName, attribute in
            switch attributeName {
            case "FTLVerb":
                guard verb == nil else {
                    return ExtractorError(message: "Could not parse more than one FTLVerb macro")
                }
                // TODO: parse export
                let name = node.name.description.trimmingCharacters(in: .whitespaces)
                verb = Verb(name: name, isExported: true, requestType: requestType, responseType: responseType)
                return nil
            default:
                return ExtractorError(message: "Unexpected \(attributeName) macro")
            }
        })
        
        guard let verb = verb else {
            return
        }
        
        self.module.decls.append(.verb(verb))
    }
    
    func extractDataStructs(_ root: Syntax) {
        traverseTree(root) { node in
            guard let structNode = node.as(StructDeclSyntax.self) else {
                return
            }
            extractDataStruct(structNode)
        }
    }
    
    func extractDataStruct(_ node: StructDeclSyntax) {
        var errors = [Error]()
        let name = node.name.description.trimmingCharacters(in: .whitespaces)
        
        var dataStruct: DataStruct?
        // var metas = [Metadata]()
        errors.append(contentsOf: self.visitAttributesIn(list: node.attributes) { attributeName, attribute in
            switch attributeName {
            case "FTLData":
                guard dataStruct == nil else {
                    return ExtractorError(message: "Could not parse more than one FTLData macro")
                }
                // TODO: parse export
                dataStruct = DataStruct(name: name, isExported: true)
                return nil
            default:
                return ExtractorError(message: "Unexpected \(attributeName) macro")
            }
        })
        guard var dataStruct = dataStruct else {
            return
        }
        
        for m in node.memberBlock.members {
            let member = MemberBlockItemSyntax(m)!
            guard let declNode = VariableDeclSyntax(member.decl) else {
                errors.append(ExtractorError(message: "Could not parse member: \(member.description)"))
                continue
            }
            for (i, b) in declNode.bindings.enumerated() {
                guard i == 0 else {
                    errors.append(ExtractorError(message: "Unexpected multiple declarations"))
                    continue
                }
                let binding = PatternBindingSyntax(b)!
                guard let memberIdentifier = IdentifierPatternSyntax(binding.pattern) else {
                    errors.append(ExtractorError(message: "Expected an identifier for the struct member"))
                    continue
                }
                guard let typeAnnotation = TypeAnnotationSyntax(binding.typeAnnotation) else {
                    errors.append(ExtractorError(message: "Expected a type annotation for the struct member"))
                    continue
                }
                do {
                    let ftlType = try ftlTypeFrom(typeAnnotation.type)
                    dataStruct.fields.append((name: memberIdentifier.description, type: ftlType))
                }
                catch {
                    errors.append(ExtractorError(message: "Could not determine field type", wrappedError: error))
                }
            }
        }
        self.errors.append(contentsOf: errors)
        self.module.decls.append(.dataStruct(dataStruct))
    }
    
    func resolveVisibility() {
        
        
    }
    
    // MARK: - Helpers

func traverseTree(_ node:Syntax, _ visitor: (Syntax) -> ()) {
    visitor(node)
    
    let children = node.children(viewMode: .fixedUp)
    for child in children {
        traverseTree(child, visitor)
    }
}

func visitAttributesIn(list: AttributeListSyntax, _ visitor:(String, AttributeSyntax) -> (Error?)) -> [Error] {
    var errors = [Error]()
    for attributeNode in list {
        guard let attributeNode = AttributeSyntax(attributeNode) else {
            continue
        }
        let attributeName = attributeNode.attributeName.description.trimmingCharacters(in: .whitespaces)
        if attributeName.hasPrefix("FTL") {
            if let error = visitor(attributeName, attributeNode) {
                errors.append(error)
            }
        }
    }
    return errors
}
    
    func ftlTypeFrom(_ type:TypeSyntax) throws -> FTLType {
        if let type = ArrayTypeSyntax(type) {
            return .array(try ftlTypeFrom(type.element))
        } else if let type = DictionaryTypeSyntax(type) {
            return .dict(try ftlTypeFrom(type.key),
                         try ftlTypeFrom(type.value))
        } else if let type = IdentifierTypeSyntax(type) {
            let name = type.description.trimmingCharacters(in: .whitespaces)
            // TODO: figure out proper packages for names
            switch name {
            case "Date":
                return .time
            default:
                return .ref(Ref(module: self.module.name, name: name))
            }
        } else if let type = OptionalTypeSyntax(type) {
            return .optional(try ftlTypeFrom(type.wrappedType))
            //        } else if let type = AttributedTypeSyntax(type) {
            //        } else if let type = ClassRestrictionTypeSyntax(type) {
            //        } else if let type = CompositionTypeSyntax(type) {
            //        } else if let type = FunctionTypeSyntax(type) {
            //        } else if let type = ImplicitlyUnwrappedOptionalTypeSyntax(type) {
            //        } else if let type = MemberTypeSyntax(type) {
            //        } else if let type = MetatypeTypeSyntax(type) {
            //        } else if let type = MissingTypeSyntax(type) {
            //        } else if let type = NamedOpaqueReturnTypeSyntax(type) {
            //        } else if let type = PackElementTypeSyntax(type) {
            //        } else if let type = PackExpansionTypeSyntax(type) {
            //        } else if let type = SomeOrAnyTypeSyntax(type) {
            //        } else if let type = SuppressedTypeSyntax(type) {
            //        } else if let type = TupleTypeSyntax(type) {
        }
        // TODO: use proper errors
        throw NSError(domain: "ftl.compile.extractor", code: 10, userInfo: [NSLocalizedDescriptionKey:"Could not determine type \(type)"])
    }
}


extension Syntax {
    /// Returns the source code of this syntax node.
    public func source() -> String {
        var result = ""
        write(to: &result)
        return result
    }
}
