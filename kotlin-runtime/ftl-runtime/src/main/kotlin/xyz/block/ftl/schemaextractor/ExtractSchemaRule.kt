package xyz.block.ftl.schemaextractor

import io.gitlab.arturbosch.detekt.api.Config
import io.gitlab.arturbosch.detekt.api.Debt
import io.gitlab.arturbosch.detekt.api.Issue
import io.gitlab.arturbosch.detekt.api.Rule
import io.gitlab.arturbosch.detekt.api.Severity
import io.gitlab.arturbosch.detekt.api.config
import io.gitlab.arturbosch.detekt.api.internal.RequiresTypeResolution
import io.gitlab.arturbosch.detekt.rules.fqNameOrNull
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
import org.jetbrains.kotlin.psi.KtTypeAlias
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
import xyz.block.ftl.secrets.Secret
import xyz.block.ftl.v1.schema.*
import xyz.block.ftl.v1.schema.Enum
import java.io.File
import java.io.FileOutputStream
import java.nio.file.Path
import java.time.OffsetDateTime
import kotlin.io.path.createDirectories

data class ModuleData(val comments: List<String> = emptyList(), val decls: MutableSet<Decl> = mutableSetOf())

// Helpers
private fun Ref.compare(module: String, name: String): Boolean = this.name == name && this.module == module

@RequiresTypeResolution
class ExtractSchemaRule(config: Config) : Rule(config) {
  private val output: String by config(defaultValue = ".")
  private val visited: MutableSet<KtFile> = mutableSetOf()
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

    // Skip if the verb is annotated with @Ignore
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

    runCatching {
      val file = annotationEntry.containingKtFile
      if (!visited.contains(file)) {
        SchemaExtractor(this.bindingContext, modules, file).extract()
        visited.add(file)
      }
    }.onFailure {
      throw it
    }
  }

  override fun postVisit(root: KtFile) {
    modules.toModules().ifNotEmpty {
      require(modules.size == 1) {
        "Each FTL module must define its own pom.xml; cannot be shared across" +
          " multiple modules"
      }
      val outputDirectory = File(output).also { Path.of(it.absolutePath).createDirectories() }
      val file = File(outputDirectory.absolutePath, OUTPUT_FILENAME)
      file.createNewFile()
      val os = FileOutputStream(file)
      os.write(this.single().encode())
      os.close()
    }
  }

  private fun Map<String, ModuleData>.toModules(): List<Module> {
    return this.map {
      xyz.block.ftl.v1.schema.Module(
        name = it.key,
        decls = it.value.decls.sortedBy { it.data_ == null },
        comments = it.value.comments
      )
    }
  }

  companion object {
    const val OUTPUT_FILENAME = "schema.pb"
  }
}

class SchemaExtractor(
  private val bindingContext: BindingContext,
  private val modules: MutableMap<String, ModuleData>,
  private val file: KtFile
) {
  private val currentModuleName = file.packageFqName.extractModuleName()
  fun extract() {
    val moduleComments = file.children
      .filterIsInstance<PsiComment>()
      .flatMap { it.text.normalizeFromDocComment() }

    val moduleData = ModuleData(
      decls = mutableSetOf(
        *extractVerbs().toTypedArray(),
        *extractDataDeclarations().toTypedArray(),
        *extractDatabases().toTypedArray(),
        *extractEnums().toTypedArray(),
        *extractConfigs().toTypedArray(),
        *extractSecrets().toTypedArray(),
      ),
      comments = moduleComments
    )
    modules[currentModuleName]?.decls?.addAll(moduleData.decls) ?: run {
      modules[currentModuleName] = moduleData
    }
  }

  private fun extractVerbs(): Set<Decl> {
    val verbs = file.children.mapNotNull { c ->
      (c as? KtNamedFunction)?.takeIf { verb ->
        verb.annotationEntries.any {
          bindingContext.get(
            BindingContext.ANNOTATION,
            it
          )?.fqName?.asString() == xyz.block.ftl.Verb::class.qualifiedName
        } && verb.annotationEntries.none {
          bindingContext.get(
            BindingContext.ANNOTATION,
            it
          )?.fqName?.asString() == xyz.block.ftl.Ignore::class.qualifiedName
        }
      }
    }
    return verbs.map {
      validateVerb(it)
      Decl(verb = extractVerb(it))
    }.toSet()
  }

  private fun validateVerb(verb: KtNamedFunction) {
    val verbSourcePos = verb.getLineAndColumn()
    requireNotNull(verb.fqName?.asString()) {
      "Verbs must be defined in a package"
    }.let { fqName ->
      require(fqName.split(".").let { it.size >= 2 && it.first() == "ftl" }) {
        "$verbSourcePos Expected @Verb to be in package ftl.<module>, but was $fqName"
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
      val position = it.getLineAndColumn().toPosition()
      return@let it.typeReference?.resolveType()?.toSchemaType(position)
    } ?: Type(unit = xyz.block.ftl.v1.schema.Unit())

    val returnRef = verb.createTypeBindingForReturnType(bindingContext)?.let {
      val position = it.psiElement.getLineAndColumn().toPosition()
      return@let it.type.toSchemaType(position)
    } ?: Type(unit = xyz.block.ftl.v1.schema.Unit())

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

  private fun extractDatabases(): Set<Decl> {
    return file.declarations
      .filter {
        (it as? KtProperty)
          ?.getDeclarationDescriptorIncludingConstructors(bindingContext)?.referencedProperty?.returnType
          ?.fqNameOrNull()?.asString() == Database::class.qualifiedName
      }
      .flatMap { it.children.asSequence() }
      .map {
        val sourcePos = it.getLineAndColumn()
        val dbName = (it as? KtCallExpression).getResolvedCall(bindingContext)?.valueArguments?.entries?.single { e ->
          e.key.name.asString() == "name"
        }
          ?.value?.toString()
          ?.trim('"')
        requireNotNull(dbName) { "$sourcePos $dbName Could not extract database name" }

        Decl(
          database = xyz.block.ftl.v1.schema.Database(
            pos = sourcePos.toPosition(),
            name = dbName
          )
        )
      }
      .toSet()
  }

  private fun extractConfigs(): Set<Decl> {
    return extractSecretsOrConfigs(xyz.block.ftl.config.Config::class.qualifiedName!!).map {
      Decl(
        config = Config(
          pos = it.position,
          name = it.name,
          type = it.type
        )
      )
    }.toSet()
  }

  private fun extractSecrets(): Set<Decl> {
    return extractSecretsOrConfigs(Secret::class.qualifiedName!!).map {
      Decl(
        secret = Secret(
          pos = it.position,
          name = it.name,
          type = it.type
        )
      )
    }.toSet()
  }

  data class SecretConfigData(val name: String, val type: Type, val position: Position)

  private fun extractSecretsOrConfigs(qualifiedPropertyName: String): List<SecretConfigData> {
    return file.declarations
      .filter {
        (it as? KtProperty)
          ?.getDeclarationDescriptorIncludingConstructors(bindingContext)?.referencedProperty?.returnType
          ?.fqNameOrNull()?.asString() == qualifiedPropertyName
      }.flatMap { it.children.asSequence() }
      .map {
        val position = it.getLineAndColumn().toPosition()
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

  private fun extractDataDeclarations(): Set<Decl> {
    return file.children
      .filter { (it is KtClass && it.isData()) || it is KtTypeAlias }
      .mapNotNull {
        val data = (it as? KtClass)?.toSchemaData() ?: (it as? KtTypeAlias)?.toSchemaData()
        data?.let { Decl(data_ = data) }
      }
      .toSet()
  }

  private fun extractEnums(): Set<Decl> {
    return file.children
      .filter { it is KtClass && it.isEnum() }
      .mapNotNull {
        (it as? KtClass)?.toSchemaEnum()?.let { enum -> Decl(enum_ = enum) }
      }
      .toSet()
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
        pos = sourcePos.toPosition(),
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
              .takeIf { refModule -> refModule != currentModuleName }
            Ref(
              name = import.importedFqName!!.asString().split(".").last(),
              module = moduleRefName ?: "",
            )
          } ?: let {
            // if no matching import, validate that the referenced verb is in the same module
            element.containingKtFile.children.singleOrNull {
              (it is KtNamedFunction) && it.name == verbCall && it.annotationEntries.any {
                bindingContext.get(
                  BindingContext.ANNOTATION,
                  it
                )?.fqName?.asString() == xyz.block.ftl.Verb::class.qualifiedName
              }
            } ?: throw IllegalArgumentException(
              "Error processing function defined at $funcSourcePos: Could not resolve outgoing verb call"
            )

            Ref(
              name = verbCall,
              module = currentModuleName,
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

  private fun KtTypeAlias.toSchemaData(): Data {
    return Data(
      name = this.name!!,
      comments = this.comments(),
      pos = getLineAndColumnInPsiFile(this.containingFile, this.textRange).toPosition(),
    )
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
            return@let it.resolveType().toSchemaType(
              getLineAndColumnInPsiFile(it.containingFile, it.textRange).toPosition()
            )
          },
          metadata = metadata,
        )
      }.toList(),
      comments = this.comments(),
      typeParameters = this.children.flatMap { (it as? KtTypeParameterList)?.parameters ?: emptyList() }.map {
        TypeParameter(
          name = it.name!!,
          pos = getLineAndColumnInPsiFile(it.containingFile, it.textRange).toPosition(),
        )
      }.toList(),
      pos = getLineAndColumnInPsiFile(this.containingFile, this.textRange).toPosition(),
    )
  }

  private fun KtClass.toSchemaEnum(): Enum {
    val variants: List<EnumVariant>
    require(this.getValueParameters().isEmpty() || this.getValueParameters().size == 1) {
      "${this.getLineAndColumn().toPosition()}: Enums can have at most one value parameter, of type string or number"
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
        val pos: Position = entry.getLineAndColumn().toPosition()
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
      pos = getLineAndColumnInPsiFile(this.containingFile, this.textRange).toPosition(),
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
        require(
          this.toClassDescriptor().isData
            || this.toClassDescriptor().kind == ClassKind.ENUM_CLASS
            || this.isEmptyBuiltin()
        ) {
          "(${position.line},${position.column}) Expected type to be a data class or builtin.Empty, but was ${
            this.fqNameOrNull()?.asString()
          }"
        }

        val refName = this.toClassDescriptor().name.asString()
        val fqName = this.fqNameOrNull()!!.asString()
        require(fqName.startsWith("ftl.")) {
          "(${position.line},${position.column}) Expected module name to be in the form ftl.<module>, " +
            "but was ${this.fqNameOrNull()?.asString()}"
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

  private fun LineAndColumn.toPosition() =
    Position(
      filename = file.name,
      line = this.line.toLong(),
      column = this.column.toLong(),
    )

  companion object {
    private fun PsiElement.getLineAndColumn(): LineAndColumn =
      getLineAndColumnInPsiFile(this.containingFile, this.textRange)

    private fun getCallMatcher(ctxVarName: String): Regex {
      return """${ctxVarName}\.call\((?<fn>[^,]+),\s*(?<req>[^,]+?)\s*[()]""".toRegex(RegexOption.IGNORE_CASE)
    }

    private fun KotlinType.toClassDescriptor(): ClassDescriptor =
      this.unwrap().constructor.declarationDescriptor as? ClassDescriptor
        ?: throw IllegalStateException("Could not resolve KotlinType to class")

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

