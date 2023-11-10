package xyz.block.ftl.schemaextractor

import io.gitlab.arturbosch.detekt.api.*
import io.gitlab.arturbosch.detekt.api.internal.RequiresTypeResolution
import io.gitlab.arturbosch.detekt.rules.fqNameOrNull
import org.jetbrains.kotlin.cfg.getElementParentDeclaration
import org.jetbrains.kotlin.descriptors.ClassDescriptor
import org.jetbrains.kotlin.name.FqName
import org.jetbrains.kotlin.psi.*
import org.jetbrains.kotlin.psi.psiUtil.getValueParameters
import org.jetbrains.kotlin.resolve.BindingContext
import org.jetbrains.kotlin.resolve.calls.util.getResolvedCall
import org.jetbrains.kotlin.resolve.calls.util.getType
import org.jetbrains.kotlin.resolve.source.getPsi
import org.jetbrains.kotlin.resolve.typeBinding.createTypeBindingForReturnType
import org.jetbrains.kotlin.types.KotlinType
import org.jetbrains.kotlin.utils.addToStdlib.ifNotEmpty
import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.schemaextractor.SchemaExtractor.Companion.moduleName
import xyz.block.ftl.v1.schema.*
import xyz.block.ftl.v1.schema.Array
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
      val currentModuleName = annotationEntry.containingKtFile.packageFqName.moduleName()
      val extractor = SchemaExtractor(this.bindingContext, modules, currentModuleName, annotationEntry)
      extractor.extract()
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
class SchemaExtractor(
  private val bindingContext: BindingContext,
  private val modules: MutableMap<String, ModuleData>,
  private val currentModuleName: String,
  annotation: KtAnnotationEntry
) {
  private val verb: KtNamedFunction
  private val module: KtDeclaration
  fun extract() {
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

    val moduleData = ModuleData(
      decls = mutableSetOf(Decl(verb = verb), *extractDataDeclarations().toTypedArray()),
      comments = module.comments()
    )
    modules[currentModuleName]?.decls?.addAll(moduleData.decls) ?: run {
      modules[currentModuleName] = moduleData
    }
  }

  private fun extractDataDeclarations(): Set<Decl> {
    return verb.containingKtFile.children
      .filter { it is KtClass && it.isData() }
      .map { Decl(data_ = (it as KtClass).toSchemaData()) }
      .toSet()
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
        method = methodArg.substringAfter("."),
      )
    }
  }

  private fun extractCalls(): MetadataCalls? {
    val verbs = mutableSetOf<VerbRef>()
    extractCalls(verb, verbs)
    return verbs.ifNotEmpty { MetadataCalls(calls = verbs.toList()) }
  }

  private fun extractCalls(element: KtElement, calls: MutableSet<VerbRef>) {
    // Step into function calls inside this expression body to look for transitive calls.
    if (element is KtCallExpression) {
      val resolvedCall = element.getResolvedCall(bindingContext)?.candidateDescriptor?.source?.getPsi() as? KtFunction
      if (resolvedCall != null) {
        extractCalls(resolvedCall, calls)
      }
    }

    val func = element as? KtNamedFunction
    if (func != null) {
      val body = requireNotNull(func.bodyExpression) { "Could not parse empty function body" }
      val imports = func.containingKtFile.importList?.imports?.mapNotNull { it.importedFqName } ?: emptyList()

      // Look for all params of type Context and extract a matcher for each based on its variable name.
      // e.g. fun foo(ctx: Context) { ctx.call(...) } => "ctx.call(...)"
      val callMatchers = func.valueParameters.filter {
        it.typeReference?.resolveType()?.fqNameOrNull()?.asString() == Context::class.qualifiedName
      }.map { ctxParam -> getCallMatcher(ctxParam.text.split(":")[0].trim()) }

      val refs = callMatchers.flatMap { matcher ->
        matcher.findAll(body.text).map {
          val req = requireNotNull(it.groups["req"]?.value?.trim()) {
            "Could not extract request type for outgoing verb call from ${verb.name}"
          }
          val verbCall = requireNotNull(it.groups["fn"]?.value?.trim()) {
            "Could not extract module name for outgoing verb call from ${verb.name}"
          }
          // TODO(worstell): Figure out how to get module name when not imported from another Kt file
          val moduleRefName = imports.firstOrNull { import -> import.toString().contains(req) }
            ?.moduleName().takeIf { refModule -> refModule != currentModuleName }

          VerbRef(
            name = verbCall.split("::")[1].trim(),
            module = moduleRefName ?: "",
          )
        }
      }
      calls.addAll(refs)
    }

    element.children
      .filter { it is KtFunction || it is KtExpression }
      .mapNotNull { it as? KtElement }
      .forEach {
        extractCalls(it, calls)
      }
  }

  private fun KtClass.toSchemaData(): Data {
    return Data(
      name = this.name!!,
      fields = this.getValueParameters().map { param ->
        Field(
          name = param.name!!,
          type = param.typeReference?.resolveType()?.toSchemaType()
        )
      }.toList(),
      comments = this.comments()
    )
  }

  private fun KotlinType.toSchemaType(): Type {
    return when (this.fqNameOrNull()?.asString()) {
      String::class.qualifiedName -> Type(string = xyz.block.ftl.v1.schema.String())
      Int::class.qualifiedName -> Type(int = xyz.block.ftl.v1.schema.Int())
      Long::class.qualifiedName -> Type(int = xyz.block.ftl.v1.schema.Int())
      Double::class.qualifiedName -> Type(float = xyz.block.ftl.v1.schema.Float())
      Boolean::class.qualifiedName -> Type(bool = xyz.block.ftl.v1.schema.Bool())
      OffsetDateTime::class.qualifiedName -> Type(time = xyz.block.ftl.v1.schema.Time())
      Map::class.qualifiedName -> {
        return Type(
          map = xyz.block.ftl.v1.schema.Map(
            key = this.arguments.first().type.toSchemaType(),
            value_ = this.arguments.last().type.toSchemaType(),
          )
        )
      }

      List::class.qualifiedName -> {
        return Type(
          array = Array(
            element = this.arguments.first().type.toSchemaType()
          )
        )
      }

      else -> {
        require(
          this.toClassDescriptor().isData
            && (this.fqNameOrNull()?.asString()?.startsWith("ftl.") ?: false)
        ) { "Expected module name to be in the form ftl.<module>, but was ${this.fqNameOrNull()?.asString()}" }
        return Type(
          dataRef = DataRef(
            name = this.toClassDescriptor().name.asString(),
            module = this.fqNameOrNull()!!.moduleName().takeIf { it != currentModuleName } ?: "",
          )
        )
      }
    }
  }

  private fun KtTypeReference.resolveType(): KotlinType =
    bindingContext.get(BindingContext.TYPE, this)
      ?: throw IllegalStateException("Could not resolve type ${this.text}")

  init {
    requireNotNull(annotation.getElementParentDeclaration()) { "Could not extract $currentModuleName verb definition" }.let {
      require(it is KtNamedFunction) { "Verbs must be functions" }
      verb = it
    }
    module = requireNotNull(verb.getElementParentDeclaration()) { "Could not extract $currentModuleName definition" }

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
    }
  }

  companion object {
    private fun getCallMatcher(ctxVarName: String): Regex {
      return """${ctxVarName}.call\((?<fn>[^)]+),(?<req>[^)]+)\(.*\)\)""".toRegex(RegexOption.IGNORE_CASE)
    }

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

