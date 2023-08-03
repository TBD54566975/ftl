package xyz.block.ftl.gradle

import org.gradle.api.Plugin
import org.gradle.api.Project
import org.jetbrains.kotlin.gradle.dsl.KotlinJvmProjectExtension

class FTLPlugin : Plugin<Project> {
  private lateinit var extension: FTLExtension
  private lateinit var project: Project

  override fun apply(project: Project) {
    project.plugins.apply("org.jetbrains.kotlin.jvm")
    this.extension = project.extensions.create("ftl", FTLExtension::class.java, project)
    this.project = project

    project.tasks.register("deploy", FTLDeploy::class.java) {
      it.group = "FTL"
      it.description = "Deploy FTL module"
    }

    project.afterEvaluate {
      check(extension.endpoint != null && extension.endpoint?.isNotEmpty() == true) {
        "FTL endpoint must be set"
      }

      extension.endpoint?.let {
        val client = FTLClient(it)
        val schemas = client.pullSchemas()
        val outputDirectory = project.file(extension.outputDirectory)
        outputDirectory.mkdir()
        extension.module.let { module ->
          ModuleGenerator().run(schemas, outputDirectory, module)
        }
      }

      // Add generated files to sourceSets
      val kotlinExtension = project.extensions.getByType(KotlinJvmProjectExtension::class.java)
      kotlinExtension.sourceSets.findByName("main")?.kotlin?.srcDir(extension.outputDirectory)
    }

    project.tasks.getByName("classes").doLast { prepareFtlRoot(project) }
  }

  // Gather all the JAR files in the runtime classpath and copy them to
  // the build/ftl directory. These will be part of the deployment.
  private fun prepareFtlRoot(project: Project) {
    val jarFiles = mutableListOf<String>()
    val classes = project.mkdir("build/ftl/jars")
    project.configurations.getByName("runtimeClasspath")
      .exclude(mapOf("group" to "xyz.block"))
      .exclude(mapOf("group" to "org.jetbrains.kotlin"))
      .forEach { file ->
        project.copy {
          jarFiles += file.name
          it.from(file)
          it.into(classes)
        }
      }

    val config = project.file("build/ftl.toml")
    config.writeText(
      """
      module = "${extension.module}"
      language = "kotlin"
      deploy = ["main", "classes/kotlin/main", "ftl"]
      """.trimIndent()
    )

    val script = project.file("build/main")
    script.writeText(
      """
      #!/bin/bash
      java -cp ftl/jars/ftl-runtime.jar:ftl/jars/${jarFiles.joinToString(":ftl/jars/")}:classes/kotlin/main xyz.block.ftl.main.MainKt
      """.trimIndent()
    )
    script.setExecutable(true)
  }
}
