package xyz.block.ftl.gradle

import org.gradle.api.Plugin
import org.gradle.api.Project

class FTLPlugin : Plugin<Project> {
  private lateinit var extension: FTLExtension
  private lateinit var project: Project

  override fun apply(project: Project) {
    this.extension = project.extensions.create("ftl", FTLExtension::class.java, project)
    this.project = project

    project.afterEvaluate {
      check(extension.endpoint != null && extension.endpoint?.isNotEmpty() == true) {
        "FTL endpoint must be set"
      }

      println("FTL endpoint: " + extension.endpoint)
    }
  }
}