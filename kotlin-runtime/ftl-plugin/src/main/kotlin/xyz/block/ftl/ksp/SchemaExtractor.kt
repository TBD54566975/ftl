package xyz.block.ftl.ksp

import com.google.devtools.ksp.KspExperimental
import com.google.devtools.ksp.closestClassDeclaration
import com.google.devtools.ksp.getAnnotationsByType
import com.google.devtools.ksp.processing.*
import com.google.devtools.ksp.symbol.KSAnnotated
import com.google.devtools.ksp.symbol.KSFunctionDeclaration
import com.google.devtools.ksp.symbol.KSVisitorVoid
import com.google.devtools.ksp.validate
import xyz.block.ftl.Ignore
import xyz.block.ftl.Ingress
import xyz.block.ftl.v1.schema.*

class Visitor(val logger: KSPLogger) : KSVisitorVoid() {
  @OptIn(KspExperimental::class)
  override fun visitFunctionDeclaration(function: KSFunctionDeclaration, data: Unit) {
    // Skip ignored classes.
    if (function.closestClassDeclaration()?.getAnnotationsByType(Ignore::class)?.firstOrNull() != null) {
      return
    }

    val metadata = mutableListOf<Metadata>()
    val qualifiedName = function.qualifiedName!!.asString()
    val parts = qualifiedName.split(".")
    if (parts.size < 2 || parts[0] != "ftl") {
      logger.error("Expected @Verb to be in package ftl.<module>, but was $qualifiedName")
      return
    }
    val module = parts[1]

    logger.info("Found @Verb $qualifiedName")

    function.getAnnotationsByType(Ingress::class).firstOrNull()?.apply {
      logger.info("  Found ingress ${this.path} ${this.method}")
      metadata += Metadata(
        ingress = MetadataIngress(
          path = this.path,
          method = this.method.toString()
        )
      )

      val verb = Verb(
        name = function.simpleName.asString(),
        metadata = metadata
      )

      val module = Module(
        name = module,
        decls = listOf(Decl(verb = verb))
      )

      logger.info("  Module $module")
    }
  }
}

class SchemaExtractor(val logger: KSPLogger) : SymbolProcessor {
  override fun process(resolver: Resolver): List<KSAnnotated> {
    val symbols = resolver.getSymbolsWithAnnotation("xyz.block.ftl.Verb")
    val ret = symbols.filter { !it.validate() }.toList()
    symbols
      .filter { it is KSFunctionDeclaration && it.validate() }
      .forEach { it.accept(Visitor(logger), Unit) }
    return ret
  }
}

class SchemaExtractorProvider : SymbolProcessorProvider {
  override fun create(environment: SymbolProcessorEnvironment): SymbolProcessor {
    return SchemaExtractor(environment.logger)
  }
}
