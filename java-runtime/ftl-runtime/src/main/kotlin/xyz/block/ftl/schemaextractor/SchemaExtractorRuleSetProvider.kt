package xyz.block.ftl.schemaextractor

import io.gitlab.arturbosch.detekt.api.Config
import io.gitlab.arturbosch.detekt.api.RuleSet
import io.gitlab.arturbosch.detekt.api.RuleSetProvider

class SchemaExtractorRuleSetProvider : RuleSetProvider {
  override val ruleSetId: String = "SchemaExtractorRuleSet"

  override fun instance(config: Config): RuleSet {
    return RuleSet(
      ruleSetId,
      listOf(
        ExtractSchemaRule(config),
      ),
    )
  }
}
