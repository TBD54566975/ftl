package xyz.block.ftl.gradle

import org.gradle.api.Plugin
import org.gradle.api.Project

class FTLPlugin : Plugin<Project> {
  private lateinit var extension: FTLExtension
  private lateinit var project: Project

  override fun apply(project: Project) {
    this.extension = project.extensions.create("ftl", FTLExtension::class.java, project)
    this.project = project

    project.tasks.register("deploy", FTLDeploy::class.java) {
      it.group = "FTL"
      it.description = "Deploy FTL module"
    }
    //
    // project.afterEvaluate {
    //   println("BEFORE")
    //   println(project.configurations.getByName("runtimeClasspath").files)
    //   project.tasks.register("run", JavaExec::class.java) {
    //     it.group = "Execution"
    //     it.description = "Run the FTL module locally."
    //     it.classpath = project.configurations.getByName("runtimeClasspath")
    //     it.mainClass.set("xyz.block.ftl.main.MainKt")
    //   }
    // }

    project.afterEvaluate {
      check(extension.endpoint != null && extension.endpoint?.isNotEmpty() == true) {
        "FTL endpoint must be set"
      }

      extension.endpoint?.let {
        println("FTL endpoint: $it")
      }
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

    val script = project.file("build/main")
    script.writeText(
      """
      #!/bin/bash
      java -cp ftl/jars/${jarFiles.joinToString(":ftl/jars/")}:classes xyz.block.ftl.main.MainKt
      """.trimIndent()
    )
    script.setExecutable(true)
  }
}
