package xyz.block.ftl.gradle

import org.gradle.api.Project

open class FTLExtension(project: Project) {
  var module: String? = null
  var endpoint: String? = null
  var outputDirectory: String = "build/generated/source"
}
