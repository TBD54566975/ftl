package xyz.block.ftl.gradle

import org.gradle.api.Project

open class FTLExtension(project: Project) {
  var endpoint: String? = null
  var module: String = project.rootProject.name
  var outputDirectory: String = "build/generated/source"
}
