package xyz.block.ftl.protoext

import xyz.block.ftl.v1.schema.DataRef
import xyz.block.ftl.v1.schema.VerbRef

/**
 * Return the fully qualified $module.$name for this VerbRef.
 */
val VerbRef.fullyQualified: String
  get() = "$module.$name"

/**
 * Return the fully qualified $module.$name for this DataRef.
 */
val DataRef.fullyQualified: String
  get() = "$module.$name"