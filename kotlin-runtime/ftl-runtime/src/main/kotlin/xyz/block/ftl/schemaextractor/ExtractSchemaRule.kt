package xyz.block.ftl.schemaextractor

import io.gitlab.arturbosch.detekt.api.Debt
import io.gitlab.arturbosch.detekt.api.Issue
import io.gitlab.arturbosch.detekt.api.Rule
import io.gitlab.arturbosch.detekt.api.Severity
import io.gitlab.arturbosch.detekt.api.config
import io.gitlab.arturbosch.detekt.api.internal.RequiresTypeResolution
import io.gitlab.arturbosch.detekt.rules.fqNameOrNull
import org.jetbrains.kotlin.backend.jvm.ir.psiElement
import org.jetbrains.kotlin.cfg.getDeclarationDescriptorIncludingConstructors
import org.jetbrains.kotlin.com.intellij.openapi.util.TextRange
import org.jetbrains.kotlin.com.intellij.psi.PsiComment
import org.jetbrains.kotlin.com.intellij.psi.PsiElement
import org.jetbrains.kotlin.descriptors.ClassDescriptor
import org.jetbrains.kotlin.descriptors.ClassKind
import org.jetbrains.kotlin.descriptors.impl.referencedProperty
import org.jetbrains.kotlin.diagnostics.DiagnosticUtils.getLineAndColumnInPsiFile
import org.jetbrains.kotlin.diagnostics.PsiDiagnosticUtils.LineAndColumn
import org.jetbrains.kotlin.name.FqName
import org.jetbrains.kotlin.psi.KtAnnotationEntry
import org.jetbrains.kotlin.psi.KtCallExpression
import org.jetbrains.kotlin.psi.KtClass
import org.jetbrains.kotlin.psi.KtDeclaration
import org.jetbrains.kotlin.psi.KtDotQualifiedExpression
import org.jetbrains.kotlin.psi.KtElement
import org.jetbrains.kotlin.psi.KtEnumEntry
import org.jetbrains.kotlin.psi.KtExpression
import org.jetbrains.kotlin.psi.KtFile
import org.jetbrains.kotlin.psi.KtFunction
import org.jetbrains.kotlin.psi.KtNamedFunction
import org.jetbrains.kotlin.psi.KtProperty
import org.jetbrains.kotlin.psi.KtSuperTypeCallEntry
import org.jetbrains.kotlin.psi.KtTypeParameterList
import org.jetbrains.kotlin.psi.KtTypeReference
import org.jetbrains.kotlin.psi.KtValueArgument
import org.jetbrains.kotlin.psi.ValueArgument
import org.jetbrains.kotlin.psi.psiUtil.children
import org.jetbrains.kotlin.psi.psiUtil.getValueParameters
import org.jetbrains.kotlin.psi.psiUtil.startOffset
import org.jetbrains.kotlin.resolve.BindingContext
import org.jetbrains.kotlin.resolve.calls.util.getResolvedCall
import org.jetbrains.kotlin.resolve.calls.util.getType
import org.jetbrains.kotlin.resolve.descriptorUtil.fqNameSafe
import org.jetbrains.kotlin.resolve.source.getPsi
import org.jetbrains.kotlin.resolve.typeBinding.createTypeBindingForReturnType
import org.jetbrains.kotlin.types.KotlinType
import org.jetbrains.kotlin.types.checker.SimpleClassicTypeSystemContext.getClassFqNameUnsafe
import org.jetbrains.kotlin.types.checker.SimpleClassicTypeSystemContext.isTypeParameterTypeConstructor
import org.jetbrains.kotlin.types.checker.anySuperTypeConstructor
import org.jetbrains.kotlin.types.isNullable
import org.jetbrains.kotlin.types.typeUtil.builtIns
import org.jetbrains.kotlin.types.typeUtil.isAny
import org.jetbrains.kotlin.types.typeUtil.isAnyOrNullableAny
import org.jetbrains.kotlin.types.typeUtil.isSubtypeOf
import org.jetbrains.kotlin.util.containingNonLocalDeclaration
import org.jetbrains.kotlin.utils.addToStdlib.ifNotEmpty
import xyz.block.ftl.Context
import xyz.block.ftl.Database
import xyz.block.ftl.HttpIngress
import xyz.block.ftl.Json
import xyz.block.ftl.Method
import xyz.block.ftl.schemaextractor.SchemaExtractor.Companion.extractModuleName
import xyz.block.ftl.v1.schema.Array
import xyz.block.ftl.v1.schema.Config
import xyz.block.ftl.v1.schema.Data
import xyz.block.ftl.v1.schema.Decl
import xyz.block.ftl.v1.schema.Enum
import xyz.block.ftl.v1.schema.EnumVariant
import xyz.block.ftl.v1.schema.Field
import xyz.block.ftl.v1.schema.IngressPathComponent
import xyz.block.ftl.v1.schema.IngressPathLiteral
import xyz.block.ftl.v1.schema.IngressPathParameter
import xyz.block.ftl.v1.schema.IntValue
import xyz.block.ftl.v1.schema.Metadata
import xyz.block.ftl.v1.schema.MetadataAlias
import xyz.block.ftl.v1.schema.MetadataCalls
import xyz.block.ftl.v1.schema.MetadataIngress
import xyz.block.ftl.v1.schema.Module
import xyz.block.ftl.v1.schema.Optional
import xyz.block.ftl.v1.schema.Position
import xyz.block.ftl.v1.schema.Ref
import xyz.block.ftl.v1.schema.Secret
import xyz.block.ftl.v1.schema.StringValue
import xyz.block.ftl.v1.schema.Type
import xyz.block.ftl.v1.schema.TypeParameter
import xyz.block.ftl.v1.schema.Unit
import xyz.block.ftl.v1.schema.Value
import xyz.block.ftl.v1.schema.Verb
import java.io.File
import java.io.FileOutputStream
import java.nio.file.Path
import java.time.OffsetDateTime
import kotlin.io.path.createDirectories
import io.gitlab.arturbosch.detekt.api.Config as DetektConfig

data class ModuleData(var comments: List<String> = emptyList(), val decls: MutableSet<Decl> = mutableSetOf())

// Helpers
private fun Ref.compare(module: String, name: String): Boolean = this.name == name && this.module == module

@RequiresTypeResolution
class ExtractSchemaRule(config: DetektConfig) : Rule(config) {
  private val output: String by config(defaultValue = ".")
  private val modules: MutableMap<String, ModuleData> = mutableMapOf()
  private var extractor = SchemaExtractor(modules)

  override val issue = Issue(
    javaClass.simpleName,
    Severity.Performance,
    "Verifies and extracts FTL Schema",
    Debt.FIVE_MINS,
  )

  override fun preVisit(root: KtFile) {
    extractor.setBindingContext(bindingContext)
    extractor.addModuleComments(root)
  }

  override fun visitAnnotationEntry(annotationEntry: KtAnnotationEntry) {
    if (
      bindingContext.get(
        BindingContext.ANNOTATION,
        annotationEntry
      )?.fqName?.asString() != xyz.block.ftl.Export::class.qualifiedName
    ) {
      return
    }

    // Skip if annotated with @Ignore
    if (
      annotationEntry.containingNonLocalDeclaration()!!.annotationEntries.any {
        bindingContext.get(
          BindingContext.ANNOTATION,
          it
        )?.fqName?.asString() == xyz.block.ftl.Ignore::class.qualifiedName
      }
    ) {
      return
    }

    when (val element = annotationEntry.parent.parent) {
      is KtNamedFunction -> extractor.addVerbToSchema(element)
      is KtClass -> {
        when {
          element.isData() -> extractor.addDataToSchema(element)
          element.isEnum() -> extractor.addEnumToSchema(element)
        }
      }
    }
  }

  override fun visitProperty(property: KtProperty) {
    when (property.getDeclarationDescriptorIncludingConstructors(bindingContext)?.referencedProperty?.returnType
      ?.fqNameOrNull()?.asString()) {
      Database::class.qualifiedName -> extractor.addDatabaseToSchema(property)
      xyz.block.ftl.secrets.Secret::class.qualifiedName -> extractor.addSecretToSchema(property)
      xyz.block.ftl.config.Config::class.qualifiedName -> extractor.addConfigToSchema(property)
    }
  }

  override fun postVisit(root: KtFile) {
    val moduleName = root.extractModuleName()
    modules[moduleName]?.let {
      val module = it.toModule(moduleName)
      val outputDirectory = File(output).also { f -> Path.of(f.absolutePath).createDirectories() }
      val file = File(outputDirectory.absolutePath, OUTPUT_FILENAME)
      file.createNewFile()
      val os = FileOutputStream(file)
      os.write(module.encode())
      os.close()
    }
  }

  private fun ModuleData.toModule(moduleName: String): Module = Module(
    name = moduleName,
    decls = this.decls.sortedBy { it.data_ == null },
    comments = this.comments
  )

  companion object {
    const val OUTPUT_FILENAME = "schema.pb"
  }
}

class SchemaExtractor(
  private val modules: MutableMap<String, ModuleData>,
) {
  private var bindingContext = BindingContext.EMPTY

  fun setBindingContext(bindingContext: BindingContext) {
    this.bindingContext = bindingContext
  }

  fun addModuleComments(file: KtFile) {
    val module = file.extractModuleName()
    val comments = file.children
      .filterIsInstance<PsiComment>()
      .flatMap { it.text.normalizeFromDocComment() }
    modules[module]?.let { it.comments = comments } ?: run {
      modules[module] = ModuleData(comments = comments)
    }
  }

  fun addVerbToSchema(verb: KtNamedFunction) {
    validateVerb(verb)
    addDecl(verb.extractModuleName(), Decl(verb = extractVerb(verb)))
  }

  fun addDataToSchema(data: KtClass) {
    addDecl(data.extractModuleName(), Decl(data_ = data.toSchemaData()))
  }

  fun addEnumToSchema(enum: KtClass) {
    addDecl(enum.extractModuleName(), Decl(enum_ = enum.toSchemaEnum()))
  }

  fun addConfigToSchema(config: KtProperty) {
    extractSecretOrConfig(config).let {
      val decl = Decl(
        config = Config(
          pos = it.position,
          name = it.name,
          type = it.type
        )
      )

      addDecl(config.extractModuleName(), decl)
    }
  }

  fun addSecretToSchema(secret: KtProperty) {
    extractSecretOrConfig(secret).let {
      val decl = Decl(
        secret = Secret(
          pos = it.position,
          name = it.name,
          type = it.type
        )
      )

      addDecl(secret.extractModuleName(), decl)
    }
  }

  fun addDatabaseToSchema(database: KtProperty) {
    val decl = database.children.single().let {
      val sourcePos = it.getPosition()
      val dbName = (it as? KtCallExpression).getResolvedCall(bindingContext)?.valueArguments?.entries?.single { e ->
        e.key.name.asString() == "name"
      }
        ?.value?.toString()
        ?.trim('"')
      requireNotNull(dbName) { "$sourcePos $dbName Could not extract database name" }

      Decl(
        database = xyz.block.ftl.v1.schema.Database(
          pos = sourcePos,
          name = dbName
        )
      )
    }
    addDecl(database.extractModuleName(), decl)
  }

  private fun addDecl(module: String, decl: Decl) {
    modules[module]?.decls?.add(decl) ?: run {
      modules[module] = ModuleData(decls = mutableSetOf(decl))
    }
  }

  private fun validateVerb(verb: KtNamedFunction) {
    val verbSourcePos = verb.getLineAndColumn()
    requireNotNull(verb.fqName?.asString()) {
      "Verbs must be defined in a package"
    }.let { fqName ->
      require(fqName.split(".").let { it.size >= 2 && it.first() == "ftl" }) {
        "$verbSourcePos Expected exported function to be in package ftl.<module>, but was $fqName"
      }

      // Validate parameters
      require(verb.valueParameters.size >= 1) { "$verbSourcePos Verbs must have at least one argument, ${verb.name} did not" }
      require(verb.valueParameters.size <= 2) { "$verbSourcePos Verbs must have at most two arguments, ${verb.name} did not" }
      val ctxParam = verb.valueParameters.first()

      require(ctxParam.typeReference?.resolveType()?.fqNameOrNull()?.asString() == Context::class.qualifiedName) {
        "${verb.valueParameters.first().getLineAndColumn()} First argument of verb must be Context"
      }

      if (verb.valueParameters.size == 2) {
        val reqParam = verb.valueParameters.last()
        require(reqParam.typeReference?.resolveType()
          ?.let { it.toClassDescriptor().isData || it.isEmptyBuiltin() }
          ?: false
        ) {
          "${verb.valueParameters.last().getLineAndColumn()} Second argument of ${verb.name} must be a data class or " +
            "builtin.Empty"
        }
      }

      // Validate return type
      verb.createTypeBindingForReturnType(bindingContext)?.type?.let {
        require(it.toClassDescriptor().isData || it.isEmptyBuiltin() || it.isUnit()) {
          "${verbSourcePos}: return type of ${verb.name} must be a data class or builtin.Empty but is ${
            it.fqNameOrNull()?.asString()
          }"
        }
      }
    }
  }

  private fun extractVerb(verb: KtNamedFunction): Verb {
    val verbSourcePos = verb.getLineAndColumn()

    val requestRef = verb.valueParameters.takeIf { it.size > 1 }?.last()?.let {
      val position = it.getPosition()
      return@let it.typeReference?.resolveType()?.toSchemaType(position)
    } ?: Type(unit = xyz.block.ftl.v1.schema.Unit())

    val returnRef = verb.createTypeBindingForReturnType(bindingContext)?.let {
      val position = it.psiElement.getPosition()
      return@let it.type.toSchemaType(position)
    } ?: Type(unit = Unit())

    val metadata = mutableListOf<Metadata>()
    extractIngress(verb, requestRef, returnRef)?.apply { metadata.add(Metadata(ingress = this)) }
    extractCalls(verb)?.apply { metadata.add(Metadata(calls = this)) }

    return Verb(
      name = requireNotNull(verb.name) { "$verbSourcePos Verbs must be named" },
      request = requestRef,
      response = returnRef,
      metadata = metadata,
      comments = verb.comments(),
    )
  }

  data class SecretConfigData(val name: String, val type: Type, val position: Position)

  private fun extractSecretOrConfig(property: KtProperty): SecretConfigData {
    return property.children.single().let {
      val position = it.getPosition()
      var type: KotlinType? = null
      var name = ""
      when (it) {
        is KtCallExpression -> {
          it.getResolvedCall(bindingContext)?.valueArguments?.entries?.forEach { arg ->
            if (arg.key.name.asString() == "name") {
              name = arg.value.toString().trim('"')
            } else if (arg.key.name.asString() == "cls") {
              type = (arg.key.varargElementType ?: arg.key.type).arguments.single().type
            }
          }
        }

        is KtDotQualifiedExpression -> {
          it.getResolvedCall(bindingContext)?.let { call ->
            name = call.valueArguments.entries.single().value.toString().trim('"')
            type = call.typeArguments.values.single()
          }
        }

        else -> {
          throw IllegalArgumentException("$position: Could not extract secret or config")
        }
      }

      SecretConfigData(name, type!!.toSchemaType(position), position)
    }
  }

  private fun extractIngress(verb: KtNamedFunction, requestType: Type, responseType: Type): MetadataIngress? {
    return verb.annotationEntries.firstOrNull {
      bindingContext.get(BindingContext.ANNOTATION, it)?.fqName?.asString() == HttpIngress::class.qualifiedName
    }?.let { annotationEntry ->
      val sourcePos = annotationEntry.getLineAndColumn()
      require(requestType.ref != null) {
        "$sourcePos ingress ${verb.name} request must be a data class"
      }
      require(responseType.ref != null) {
        "$sourcePos ingress ${verb.name} response must be a data class"
      }
      require(requestType.ref.compare("builtin", "HttpRequest")) {
        "$sourcePos @HttpIngress-annotated ${verb.name} request must be ftl.builtin.HttpRequest"
      }
      require(responseType.ref.compare("builtin", "HttpResponse")) {
        "$sourcePos @HttpIngress-annotated ${verb.name} response must be ftl.builtin.HttpResponse"
      }
      require(annotationEntry.valueArguments.size >= 2) {
        "$sourcePos ${verb.name} @HttpIngress annotation requires at least 2 arguments"
      }

      val args = annotationEntry.valueArguments.partition { arg ->
        // Method arg is named "method" or is of type xyz.block.ftl.Method (in the case where args are
        // positional rather than named).
        arg.getArgumentName()?.asName?.asString() == "method"
          || arg.getArgumentExpression()?.getType(bindingContext)?.fqNameOrNull()
          ?.asString() == Method::class.qualifiedName
      }

      val methodArg = requireNotNull(args.first.single().getArgumentExpression()?.text?.substringAfter(".")) {
        "Could not extract method from ${verb.name} @HttpIngress annotation"
      }
      val pathArg = requireNotNull(args.second.single().getArgumentExpression()?.text?.let {
        extractPathComponents(it.trim('\"'))
      }) {
        "Could not extract path from ${verb.name} @HttpIngress annotation"
      }

      MetadataIngress(
        type = "http",
        path = pathArg,
        method = methodArg,
        pos = sourcePos.toPosition(verb.containingFile.name),
      )
    }
  }


  private fun extractPathComponents(path: String): List<IngressPathComponent> {
    return path.split("/").filter { it.isNotEmpty() }.map { part ->
      if (part.startsWith("{") && part.endsWith("}")) {
        IngressPathComponent(ingressPathParameter = IngressPathParameter(name = part.substring(1, part.length - 1)))
      } else {
        IngressPathComponent(ingressPathLiteral = IngressPathLiteral(text = part))
      }
    }
  }

  private fun extractCalls(verb: KtNamedFunction): MetadataCalls? {
    val verbs = mutableSetOf<Ref>()
    extractCalls(verb, verbs)
    return verbs.ifNotEmpty { MetadataCalls(calls = verbs.toList()) }
  }

  private fun extractCalls(element: KtElement, calls: MutableSet<Ref>) {
    // Step into function calls inside this expression body to look for transitive calls.
    if (element is KtCallExpression) {
      val resolvedCall = element.getResolvedCall(bindingContext)?.candidateDescriptor?.source?.getPsi() as? KtFunction
      if (resolvedCall != null) {
        extractCalls(resolvedCall, calls)
      }
    }

    val func = element as? KtNamedFunction
    if (func != null) {
      val funcSourcePos = func.getLineAndColumn()
      val body =
        requireNotNull(func.bodyExpression) { "$funcSourcePos Function body cannot be empty; was in ${func.name}" }

      val commentRanges = body.node.children()
        .filterIsInstance<PsiComment>()
        .map { it.textRange.shiftLeft(it.startOffset).shiftRight(it.startOffsetInParent) }

      val imports = func.containingKtFile.importList?.imports ?: emptyList()

      // Look for all params of type Context and extract a matcher for each based on its variable name.
      // e.g. fun foo(ctx: Context) { ctx.call(...) } => "ctx.call(...)"
      val callMatchers = func.valueParameters.filter {
        it.typeReference?.resolveType()?.fqNameOrNull()?.asString() == Context::class.qualifiedName
      }.map { ctxParam -> getCallMatcher(ctxParam.text.split(":")[0].trim()) }

      val refs = callMatchers.flatMap { matcher ->
        matcher.findAll(body.text).mapNotNull { match ->
          // ignore commented out matches
          if (commentRanges.any { it.contains(TextRange(match.range.first, match.range.last)) }) return@mapNotNull null

          val verbCall = requireNotNull(match.groups["fn"]?.value?.substringAfter("::")?.trim()) {
            "Error processing function defined at $funcSourcePos: Could not extract outgoing verb call"
          }
          imports.firstOrNull { import ->
            // if aliased import, match the alias
            (import.text.split(" ").takeIf { it.size > 2 }?.last()
            // otherwise match the last part of the import
              ?: import.importedFqName?.asString()?.split(".")?.last()) == verbCall
          }?.let { import ->
            val moduleRefName = import.importedFqName?.asString()?.extractModuleName()
              .takeIf { refModule -> refModule != element.extractModuleName() }
            Ref(
              name = import.importedFqName!!.asString().split(".").last(),
              module = moduleRefName ?: "",
            )
          } ?: let {
            // if no matching import, validate that the referenced verb is in the same module
            element.containingFile.children.singleOrNull {
              (it is KtNamedFunction) && it.name == verbCall && it.annotationEntries.any {
                bindingContext.get(
                  BindingContext.ANNOTATION,
                  it
                )?.fqName?.asString() == xyz.block.ftl.Export::class.qualifiedName
              }
            } ?: throw IllegalArgumentException(
              "Error processing function defined at $funcSourcePos: Could not resolve outgoing verb call"
            )

            Ref(
              name = verbCall,
              module = element.extractModuleName(),
            )
          }
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
        // Metadata containing JSON alias if present.
        val metadata = param.annotationEntries.firstOrNull {
          bindingContext.get(BindingContext.ANNOTATION, it)?.fqName?.asString() == Json::class.qualifiedName
        }?.valueArguments?.single()?.let {
          listOf(Metadata(alias = MetadataAlias(alias = (it as KtValueArgument).text.trim('"', ' '))))
        } ?: listOf()
        Field(
          name = param.name!!,
          type = param.typeReference?.let {
            return@let it.resolveType().toSchemaType(it.getPosition())
          },
          metadata = metadata,
        )
      }.toList(),
      comments = this.comments(),
      typeParameters = this.children.flatMap { (it as? KtTypeParameterList)?.parameters ?: emptyList() }.map {
        TypeParameter(
          name = it.name!!,
          pos = getLineAndColumnInPsiFile(it.containingFile, it.textRange).toPosition(this.containingFile.name),
        )
      }.toList(),
      pos = getLineAndColumnInPsiFile(this.containingFile, this.textRange).toPosition(this.containingFile.name),
    )
  }

  private fun KtClass.toSchemaEnum(): Enum {
    val variants: List<EnumVariant>
    require(this.getValueParameters().isEmpty() || this.getValueParameters().size == 1) {
      "${this.getLineAndColumn()}: Enums can have at most one value parameter, of type string or number"
    }

    if (this.getValueParameters().isEmpty()) {
      var ordinal = 0L
      variants = this.declarations.filterIsInstance<KtEnumEntry>().map {
        val variant = EnumVariant(
          name = it.name!!,
          value_ = Value(intValue = IntValue(value_ = ordinal)),
          comments = it.comments(),
        )
        ordinal = ordinal.inc()
        return@map variant
      }
    } else {
      variants = this.declarations.filterIsInstance<KtEnumEntry>().map { entry ->
        val pos: Position = entry.getPosition()
        val name: String = entry.name!!
        val arg: ValueArgument = entry.initializerList?.initializers?.single().let {
          (it as KtSuperTypeCallEntry).valueArguments.single()
        }

        val value: Value
        try {
          value = arg.getArgumentExpression()?.text?.let {
            if (it.startsWith('"')) {
              return@let Value(stringValue = StringValue(value_ = it.trim('"')))
            } else {
              return@let Value(intValue = IntValue(value_ = it.toLong()))
            }
          } ?: throw IllegalArgumentException("${pos}: Could not extract enum variant value")
        } catch (e: NumberFormatException) {
          throw IllegalArgumentException("${pos}: Enum variant value must be a string or number")
        }

        EnumVariant(
          name = name,
          value_ = value,
          pos = pos,
          comments = entry.comments(),
        )
      }
    }

    return Enum(
      name = this.name!!,
      variants = variants,
      comments = this.comments(),
      pos = getLineAndColumnInPsiFile(this.containingFile, this.textRange).toPosition(this.containingFile.name),
    )
  }

  private fun KotlinType.toSchemaType(position: Position): Type {
    if (this.unwrap().constructor.isTypeParameterTypeConstructor()) {
      return Type(
        ref = Ref(
          name = this.constructor.declarationDescriptor?.name?.asString() ?: "T",
          pos = position,
        )
      )
    }
    val builtIns = this.builtIns
    val type = this.constructor.declarationDescriptor!!.defaultType
    val schemaType = when {
      type.isSubtypeOf(builtIns.stringType) -> Type(string = xyz.block.ftl.v1.schema.String())
      type.isSubtypeOf(builtIns.intType) -> Type(int = xyz.block.ftl.v1.schema.Int())
      type.isSubtypeOf(builtIns.longType) -> Type(int = xyz.block.ftl.v1.schema.Int())
      type.isSubtypeOf(builtIns.floatType) -> Type(float = xyz.block.ftl.v1.schema.Float())
      type.isSubtypeOf(builtIns.doubleType) -> Type(float = xyz.block.ftl.v1.schema.Float())
      type.isSubtypeOf(builtIns.booleanType) -> Type(bool = xyz.block.ftl.v1.schema.Bool())
      type.isSubtypeOf(builtIns.unitType) -> Type(unit = xyz.block.ftl.v1.schema.Unit())
      type.anySuperTypeConstructor {
        it.getClassFqNameUnsafe().asString() == ByteArray::class.qualifiedName
      } -> Type(bytes = xyz.block.ftl.v1.schema.Bytes())

      type.anySuperTypeConstructor {
        it.getClassFqNameUnsafe().asString() == builtIns.list.fqNameSafe.asString()
      } -> Type(
        array = Array(
          element = this.arguments.first().type.toSchemaType(position)
        )
      )

      type.anySuperTypeConstructor {
        it.getClassFqNameUnsafe().asString() == builtIns.map.fqNameSafe.asString()
      } -> Type(
        map = xyz.block.ftl.v1.schema.Map(
          key = this.arguments.first().type.toSchemaType(position),
          value_ = this.arguments.last().type.toSchemaType(position),
        )
      )

      this.isAnyOrNullableAny() -> Type(any = xyz.block.ftl.v1.schema.Any())
      this.fqNameOrNull()
        ?.asString() == OffsetDateTime::class.qualifiedName -> Type(time = xyz.block.ftl.v1.schema.Time())

      else -> {
        val descriptor = this.toClassDescriptor()
        require(
          descriptor.isData
            || descriptor.kind == ClassKind.ENUM_CLASS
            || this.isEmptyBuiltin()
        ) {
          "(${position.line},${position.column}) Expected type to be a data class or builtin.Empty, but was ${
            this.fqNameOrNull()?.asString()
          }"
        }

        val refName = descriptor.name.asString()
        val fqName = this.fqNameOrNull()!!.asString()
        require(fqName.startsWith("ftl.")) {
          "(${position.line},${position.column}) Expected module name to be in the form ftl.<module>, " +
            "but was ${this.fqNameOrNull()?.asString()}"
        }

        // add all referenced types to the schema
        // TODO: remove once we require explicit exporting of types
        (descriptor.psiElement as? KtClass)?.let {
          when {
            it.isData() -> addDecl(it.extractModuleName(), Decl(data_ = it.toSchemaData()))
            it.isEnum() -> addDecl(it.extractModuleName(), Decl(enum_ = it.toSchemaEnum()))
          }
        }

        Type(
          ref = Ref(
            name = refName,
            module = fqName.extractModuleName(),
            pos = position,
            typeParameters = this.arguments.map { it.type.toSchemaType(position) }.toList(),
          )
        )
      }
    }
    if (this.isNullable()) {
      return Type(optional = Optional(type = schemaType))
    }
    if (this.isAny()) {
      return Type(any = xyz.block.ftl.v1.schema.Any())
    }
    return schemaType
  }

  private fun KtTypeReference.resolveType(): KotlinType =
    bindingContext.get(BindingContext.TYPE, this)
      ?: throw IllegalStateException("${this.getLineAndColumn()} Could not resolve type ${this.text}")

  companion object {
    private fun KtElement.getPosition() = this.getLineAndColumn().toPosition(this.containingFile.name)

    private fun PsiElement.getPosition() = this.getLineAndColumn().toPosition(this.containingFile.name)
    private fun PsiElement.getLineAndColumn(): LineAndColumn =
      getLineAndColumnInPsiFile(this.containingFile, this.textRange)

    private fun LineAndColumn.toPosition(filename: String) =
      Position(
        filename = filename,
        line = this.line.toLong(),
        column = this.column.toLong(),
      )

    private fun getCallMatcher(ctxVarName: String): Regex {
      return """${ctxVarName}\.call\((?<fn>[^,]+),\s*(?<req>[^,]+?)\s*[()]""".toRegex(RegexOption.IGNORE_CASE)
    }

    private fun KotlinType.toClassDescriptor(): ClassDescriptor =
      this.unwrap().constructor.declarationDescriptor as? ClassDescriptor
        ?: throw IllegalStateException("Could not resolve KotlinType to class")

    fun KtElement.extractModuleName(): String {
      return this.containingKtFile.packageFqName.extractModuleName()
    }

    fun FqName.extractModuleName(): String {
      return this.asString().extractModuleName()
    }

    private fun String.extractModuleName(): String {
      return this.split(".")[1]
    }

    private fun KtDeclaration.comments(): List<String> {
      return this.docComment?.text?.trim()?.normalizeFromDocComment() ?: emptyList()
    }

    private fun String.normalizeFromDocComment(): List<String> {
      // get comments without comment markers
      return this.lines()
        .mapNotNull { line ->
          line.removePrefix("/**")
            .removePrefix("/*")
            .removeSuffix("*/")
            .takeIf { it.isNotBlank() }
        }
        .map { it.trim('*', '/', ' ') }
        .toList()
    }

    private fun KotlinType.isEmptyBuiltin(): Boolean {
      return this.fqNameOrNull()?.asString() == "ftl.builtin.Empty"
    }

    private fun KotlinType.isUnit(): Boolean {
      return this.fqNameOrNull()?.asString() == "kotlin.Unit"
    }
  }
}

