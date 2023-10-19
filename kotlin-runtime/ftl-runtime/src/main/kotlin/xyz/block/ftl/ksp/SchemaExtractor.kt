package xyz.block.ftl.ksp

import com.google.devtools.ksp.KspExperimental
import com.google.devtools.ksp.closestClassDeclaration
import com.google.devtools.ksp.getAnnotationsByType
import com.google.devtools.ksp.processing.*
import com.google.devtools.ksp.symbol.*
import com.google.devtools.ksp.validate
import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Ingress
import xyz.block.ftl.v1.schema.*
import xyz.block.ftl.v1.schema.Array
import java.io.File
import java.io.FileOutputStream
import java.nio.file.Path
import java.time.OffsetDateTime
import kotlin.io.path.createDirectories
import kotlin.reflect.KClass

data class ModuleData(val comments: List<String> = emptyList(), val decls: MutableSet<Decl> = mutableSetOf())

class Visitor(val logger: KSPLogger, val modules: MutableMap<String, ModuleData>) :
  KSVisitorVoid() {
  @OptIn(KspExperimental::class)
  override fun visitFunctionDeclaration(function: KSFunctionDeclaration, data: Unit) {
    // Skip ignored classes.
    if (function.closestClassDeclaration()?.getAnnotationsByType(Ignore::class)?.firstOrNull() != null) {
      return
    }

    validateVerb(function)

    val metadata = mutableListOf<Metadata>()
    val moduleName = function.qualifiedName!!.moduleName()
    val requestType = function.parameters.last().type.resolve().declaration
    val responseType = function.returnType!!.resolve().declaration
   if (modules[moduleName] == null) {
      modules[moduleName] = ModuleData(comments = function.closestClassDeclaration()?.comments() ?: emptyList())
    }

    function.getAnnotationsByType(Ingress::class).firstOrNull()?.apply {
      metadata += Metadata(
        ingress = MetadataIngress(
          path = this.path,
          method = this.method.toString()
        )
      )

      val verb = Verb(
        name = function.simpleName.asString(),
        request = requestType.toSchemaType().dataRef,
        response = responseType.toSchemaType().dataRef,
        metadata = metadata,
        comments = function.comments(),
      )

      val requestData = Decl(data_ = requestType.closestClassDeclaration()!!.toSchemaData())
      val responseData = Decl(data_ = responseType.closestClassDeclaration()!!.toSchemaData())
      val decls = mutableSetOf(Decl(verb = verb), requestData, responseData)
      modules[moduleName]!!.decls.addAll(decls)
    }
  }

  private fun validateVerb(verb: KSFunctionDeclaration) {
    val params = verb.parameters.map { it.type.resolve().declaration }
    require(params.size == 2) { "Verbs must have exactly two arguments" }
    require(params.first().toKClass() == Context::class) { "First argument of verb must be Context" }
    require(params.last().modifiers.contains(Modifier.DATA)) { "Second argument of verb must be a data class" }
    require(verb.returnType?.resolve()?.declaration?.modifiers?.contains(Modifier.DATA) == true) {
      "Return type of verb must be a data class"
    }

    val qualifiedName = verb.qualifiedName!!.asString()
    require(qualifiedName.split(".").let { it.size >= 2 && it.first() == "ftl" }) {
      "Expected @Verb to be in package ftl.<module>, but was $qualifiedName"
    }
  }

  private fun KSClassDeclaration.toSchemaData(): Data {
    return Data(
      name = this.simpleName.asString(),
      fields = this.getAllProperties()
        .map { param ->
          Field(
            name = param.simpleName.asString(),
            type = param.type.resolve().declaration.toSchemaType(param.type.element?.typeArguments)
          )
        }.toList(),
      comments = this.comments(),
    )
  }

  private fun KSDeclaration.toSchemaType(typeArguments: List<KSTypeArgument>? = emptyList()): Type {
    return when (this.qualifiedName!!.asString()) {
      String::class.qualifiedName -> Type(string = xyz.block.ftl.v1.schema.String())
      Int::class.qualifiedName -> Type(int = xyz.block.ftl.v1.schema.Int())
      Long::class.qualifiedName -> Type(int = xyz.block.ftl.v1.schema.Int())
      Boolean::class.qualifiedName -> Type(bool = Bool())
      OffsetDateTime::class.qualifiedName -> Type(time = Time())
      Map::class.qualifiedName -> {
        return Type(
          map = xyz.block.ftl.v1.schema.Map(
            key = typeArguments!!.first()
              .let { it.type?.resolve()?.declaration?.toSchemaType(it.type?.element?.typeArguments) },
            value_ = typeArguments.last()
              .let { it.type?.resolve()?.declaration?.toSchemaType(it.type?.element?.typeArguments) },
          )
        )
      }

      List::class.qualifiedName -> {
        return Type(
          array = Array(
            element = typeArguments!!.first()
              .let { it.type?.resolve()?.declaration?.toSchemaType(it.type?.element?.typeArguments) }
          )
        )
      }

      else -> {
        this.closestClassDeclaration()?.let {
          if (it.simpleName != this.simpleName) {
            return@let
          }

          // Make sure any nested data classes are included in the module schema.
          val decl = Decl(data_ = it.toSchemaData())
          val moduleName = it.qualifiedName!!.moduleName()
          modules[moduleName]!!.decls.add(decl)
        }
        return Type(dataRef = DataRef(name = this.simpleName.asString()))
      }
    }
  }

  companion object {
    private fun KSDeclaration.toKClass(): KClass<*> {
      return Class.forName(this.qualifiedName?.asString()).kotlin
    }

    private fun KSDeclaration.comments(): List<String> {
      return this.docString?.trim()?.let { listOf(it) } ?: emptyList()
    }

    private fun KSName.moduleName(): String {
      return this.asString().split(".")[1]
    }
  }
}

class SchemaExtractor(val logger: KSPLogger, val options: Map<String, String>) : SymbolProcessor {
  override fun process(resolver: Resolver): List<KSAnnotated> {
    val dest = requireNotNull(options["dest"]) { "Must provide output directory for generated schemas" }
    val outputDirectory = File(dest).also { Path.of(it.absolutePath).createDirectories() }
    val modules = mutableMapOf<String, ModuleData>()

    val symbols = resolver.getSymbolsWithAnnotation("xyz.block.ftl.Verb")
    val ret = symbols.filter { !it.validate() }.toList()
    symbols
      .filter { it is KSFunctionDeclaration && it.validate() }
      .forEach { it.accept(Visitor(logger, modules), Unit) }

    modules.map {
      Module(name = it.key, decls = it.value.decls.sortedBy { it.data_ == null }, comments = it.value.comments)
    }.forEach {
      val file = File(outputDirectory.absolutePath, it.name)
      file.createNewFile()
      val os = FileOutputStream(file)
      os.write(it.encode())
      os.close()
    }

    return ret
  }
}

class SchemaExtractorProvider : SymbolProcessorProvider {
  override fun create(environment: SymbolProcessorEnvironment): SymbolProcessor {
    return SchemaExtractor(environment.logger, environment.options)
  }
}
