package xyz.block.ftl.schemaextractor

import io.gitlab.arturbosch.detekt.api.*
import io.gitlab.arturbosch.detekt.api.internal.RequiresTypeResolution
import io.gitlab.arturbosch.detekt.rules.fqNameOrNull
import org.jetbrains.kotlin.backend.common.serialization.metadata.findKDocString
import org.jetbrains.kotlin.cfg.getElementParentDeclaration
import org.jetbrains.kotlin.descriptors.ClassDescriptor
import org.jetbrains.kotlin.descriptors.impl.referencedProperty
import org.jetbrains.kotlin.name.FqName
import org.jetbrains.kotlin.psi.*
import org.jetbrains.kotlin.resolve.BindingContext
import org.jetbrains.kotlin.resolve.calls.util.getResolvedCall
import org.jetbrains.kotlin.resolve.calls.util.getType
import org.jetbrains.kotlin.resolve.scopes.DescriptorKindFilter
import org.jetbrains.kotlin.resolve.scopes.getDescriptorsFiltered
import org.jetbrains.kotlin.resolve.source.getPsi
import org.jetbrains.kotlin.resolve.typeBinding.createTypeBindingForReturnType
import org.jetbrains.kotlin.types.KotlinType
import org.jetbrains.kotlin.types.TypeProjection
import org.jetbrains.kotlin.utils.addToStdlib.ifNotEmpty
import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.schemaextractor.SchemaExtractor.Companion.moduleName
import xyz.block.ftl.v1.schema.*
import java.io.File
import java.io.FileOutputStream
import java.nio.file.Path
import java.time.OffsetDateTime
import kotlin.Boolean
import kotlin.String
import kotlin.collections.Map
import kotlin.collections.set
import kotlin.io.path.createDirectories

data class ModuleData(val comments: List<String> = emptyList(), val decls: MutableSet<Decl> = mutableSetOf())

@RequiresTypeResolution
class ExtractSchemaRule(config: Config) : Rule(config) {
  private val output: String by config(defaultValue = ".")
  private val modules: MutableMap<String, ModuleData> = mutableMapOf()

  override val issue = Issue(
    javaClass.simpleName,
    Severity.Performance,
    "Verifies and extracts FTL Schema",
    Debt.FIVE_MINS,
  )

  override fun visitAnnotationEntry(annotationEntry: KtAnnotationEntry) {
    if (
      bindingContext.get(
        BindingContext.ANNOTATION,
        annotationEntry
      )?.fqName?.asString() != xyz.block.ftl.Verb::class.qualifiedName
    ) {
      return
    }

    runCatching {
      val extractor = SchemaExtractor(this.bindingContext, annotationEntry)
      val moduleName = annotationEntry.containingKtFile.packageFqName.moduleName()
      val moduleData = extractor.extract()
      modules[moduleName]?.let { it.decls += moduleData.decls }
        ?: run { modules[moduleName] = moduleData }
    }.onFailure {
      when (it) {
        is IgnoredModuleException -> return
        else -> throw it
      }
    }
  }

  override fun postVisit(root: KtFile) {
    val outputDirectory = File(output).also { Path.of(it.absolutePath).createDirectories() }

    modules.toModules().forEach {
      val file = File(outputDirectory.absolutePath, OUTPUT_FILENAME)
      file.createNewFile()
      val os = FileOutputStream(file)
      os.write(it.encode())
      os.close()
    }
  }

  companion object {
    const val OUTPUT_FILENAME = "schema.pb"

    private fun Map<String, ModuleData>.toModules(): List<Module> {
      return this.map {
        Module(name = it.key, decls = it.value.decls.sortedBy { it.data_ == null }, comments = it.value.comments)
      }
    }
  }
}

class IgnoredModuleException : Exception()
class SchemaExtractor(val bindingContext: BindingContext, annotation: KtAnnotationEntry) {
  private val callMatcher: Regex
  private val verb: KtNamedFunction
  private val module: KtDeclaration
  private val decls: MutableSet<Decl> = mutableSetOf()
  fun extract(): ModuleData {
    val requestType = requireNotNull(verb.valueParameters.last().typeReference?.resolveType()) {
      "Could not resolve verb request type"
    }
    val responseType = requireNotNull(verb.createTypeBindingForReturnType(bindingContext)?.type) {
      "Could not resolve verb response type"
    }

    val metadata = mutableListOf<Metadata>()
    extractIngress()?.apply { metadata.add(Metadata(ingress = this)) }
    extractCalls()?.apply { metadata.add(Metadata(calls = this)) }

    val verb = Verb(
      name = requireNotNull(verb.name) { "Verbs must be named" },
      request = requestType.toSchemaType().dataRef,
      response = responseType.toSchemaType().dataRef,
      metadata = metadata,
      comments = verb.comments(),
    )

    decls.addAll(
      listOf(
        Decl(verb = verb),
        Decl(data_ = requestType.toSchemaData()),
        Decl(data_ = responseType.toSchemaData())
      )
    )

    return ModuleData(decls = decls, comments = module.comments())
  }

  private fun extractIngress(): MetadataIngress? {
    return verb.annotationEntries.firstOrNull {
      bindingContext.get(BindingContext.ANNOTATION, it)?.fqName?.asString() == Ingress::class.qualifiedName
    }?.let {
      val argumentLists = it.valueArguments.partition { arg ->
        // Method arg is named "method" or is of type xyz.block.ftl.Method (in the case where args are
        // positional rather than named).
        arg.getArgumentName()?.asName?.asString() == "method"
          || arg.getArgumentExpression()?.getType(bindingContext)?.fqNameOrNull()
          ?.asString() == Method::class.qualifiedName
      }
      val methodArg = requireNotNull(argumentLists.first.single().getArgumentExpression()?.text) {
        "Could not extract method from ${verb.name} @Ingress annotation"
      }
      // NB: trim leading/trailing double quotes because KtStringTemplateExpression.text includes them
      val pathArg = requireNotNull(argumentLists.second.single().getArgumentExpression()?.text?.trim { it == '\"' }) {
        "Could not extract path from ${verb.name} @Ingress annotation"
      }
      MetadataIngress(
        path = pathArg,
        method = methodArg,
      )
    }
  }

  private fun extractCalls(): MetadataCalls? {
    val verbs = mutableListOf<VerbRef>()
    extractCalls(verb, verbs)
    return verbs.ifNotEmpty { MetadataCalls(calls = verbs) }
  }

  private fun extractCalls(func: KtNamedFunction, calls: MutableList<VerbRef>) {
    val body = requireNotNull(func.bodyExpression) { "Verbs must have a body" }
    val imports = func.containingKtFile.importList?.imports?.mapNotNull { it.importedFqName } ?: emptyList()

    val refs = callMatcher.findAll(body.text).map {
      val req = requireNotNull(it.groups["req"]?.value?.trim()) {
        "Could not extract request type for outgoing verb call from ${verb.name}"
      }
      val verbCall = requireNotNull(it.groups["fn"]?.value?.trim()) {
        "Could not extract module name for outgoing verb call from ${verb.name}"
      }
      // TODO(worstell): Figure out how to get module name when not imported from another Kt file
      val moduleRefName = imports.filter { it.toString().contains(req) }.firstOrNull()?.moduleName()

      VerbRef(
        name = verbCall.split("::")[1].trim(),
        module = moduleRefName ?: "",
      )
    }
    calls.addAll(refs)

    // Step into function calls inside this expression body to look for transitive calls.
    body.children.mapNotNull {
      (it as? KtCallExpression)
        ?.getResolvedCall(bindingContext)?.candidateDescriptor?.source?.getPsi() as? KtNamedFunction
    }.forEach {
      extractCalls(it, calls)
    }
  }

  private fun KotlinType.toSchemaData(): Data {
    return Data(
      name = this.toClassDescriptor().name.asString(),
      fields = this.memberScope.getDescriptorsFiltered(DescriptorKindFilter.VARIABLES).map { property ->
        val param = requireNotNull(property.referencedProperty?.type) { "Could not resolve data class property type" }
        Field(
          name = property.name.asString(),
          type = param.toSchemaType(param.arguments)
        )
      }.toList(),
      comments = this.toClassDescriptor().findKDocString()?.trim()?.let { listOf(it) } ?: emptyList()
    )
  }

  private fun KotlinType.toSchemaType(typeArguments: List<TypeProjection> = emptyList()): Type {
    return when (this.fqNameOrNull()?.asString()) {
      String::class.qualifiedName -> Type(string = xyz.block.ftl.v1.schema.String())
      Long::class.qualifiedName -> Type(int = xyz.block.ftl.v1.schema.Int())
      Double::class.qualifiedName -> Type(float = xyz.block.ftl.v1.schema.Float())
      Boolean::class.qualifiedName -> Type(bool = xyz.block.ftl.v1.schema.Bool())
      OffsetDateTime::class.qualifiedName -> Type(time = xyz.block.ftl.v1.schema.Time())
      Map::class.qualifiedName -> {
        return Type(
          map = xyz.block.ftl.v1.schema.Map(
            key = typeArguments.first().let { it.type.toSchemaType(it.type.getTypeArguments()) },
            value_ = typeArguments.last().let { it.type.toSchemaType(it.type.getTypeArguments()) },
          )
        )
      }

      List::class.qualifiedName -> {
        return Type(
          array = Array(
            element = typeArguments.first().let { it.type.toSchemaType(it.type.getTypeArguments()) }
          )
        )
      }

      else -> {
        require(
          this.toClassDescriptor().isData
            && (this.fqNameOrNull()?.asString()?.startsWith("ftl.") ?: false)
        ) { "${this.fqNameOrNull()?.asString()} type is not supported in FTL schema" }

        // Make sure any nested data classes are included in the module schema.
        decls.add(Decl(data_ = this.toSchemaData()))
        return Type(
          dataRef = DataRef(
            name = this.toClassDescriptor().name.asString(),
            module = this.fqNameOrNull()!!.moduleName()
          )
        )
      }
    }
  }

  private fun KtTypeReference.resolveType(): KotlinType =
    bindingContext.get(BindingContext.TYPE, this)
      ?: throw IllegalStateException("Could not resolve type ${this.text}")

  init {
    val moduleName = annotation.containingKtFile.packageFqName.moduleName()
    requireNotNull(annotation.getElementParentDeclaration()) { "Could not extract $moduleName verb definition" }.let {
      require(it is KtNamedFunction) { "Verbs must be functions" }
      verb = it
    }
    module = requireNotNull(verb.getElementParentDeclaration()) { "Could not extract $moduleName definition" }

    // Skip ignored modules.
    if (module.annotationEntries.firstOrNull {
        bindingContext.get(BindingContext.ANNOTATION, it)?.fqName?.asString() == Ignore::class.qualifiedName
      } != null) {
      throw IgnoredModuleException()
    }

    requireNotNull(verb.fqName?.asString()) {
      "Verbs must be defined in a package"
    }.let { fqName ->
      require(fqName.split(".").let { it.size >= 2 && it.first() == "ftl" }) {
        "Expected @Verb to be in package ftl.<module>, but was $fqName"
      }

      // Validate parameters
      require(verb.valueParameters.size == 2) { "Verbs must have exactly two arguments" }
      val ctxParam = verb.valueParameters.first()
      val reqParam = verb.valueParameters.last()
      require(ctxParam.typeReference?.resolveType()?.fqNameOrNull()?.asString() == Context::class.qualifiedName) {
        "First argument of verb must be Context"
      }
      require(reqParam.typeReference?.resolveType()?.toClassDescriptor()?.isData ?: false) {
        "Second argument of verb must be a data class"
      }

      // Validate return type
      val respClass = verb.createTypeBindingForReturnType(bindingContext)?.type?.toClassDescriptor()
        ?: throw IllegalStateException("Could not resolve verb return type")
      require(respClass.isData) { "Return type of verb must be a data class" }

      val ctxVarName = ctxParam.text.split(":")[0].trim()
      callMatcher = """${ctxVarName}.call\((?<fn>[^)]+),(?<req>[^)]+)\(\)\)""".toRegex(RegexOption.IGNORE_CASE)
    }
  }

  companion object {
    private fun KotlinType.getTypeArguments(): List<TypeProjection> =
      this.memberScope.getDescriptorsFiltered(DescriptorKindFilter.VARIABLES)
        .flatMap { it.referencedProperty!!.type.arguments }

    private fun KotlinType.toClassDescriptor(): ClassDescriptor =
      this.unwrap().constructor.declarationDescriptor as? ClassDescriptor
        ?: throw IllegalStateException("Could not resolve KotlinType to class")

    fun FqName.moduleName(): String {
      return this.asString().split(".")[1]
    }

    private fun KtDeclaration.comments(): List<String> {
      return this.docComment?.text?.trim()?.let { listOf(it) } ?: emptyList()
    }
  }
}

