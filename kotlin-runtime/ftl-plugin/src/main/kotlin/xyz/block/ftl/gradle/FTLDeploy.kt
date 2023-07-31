package xyz.block.ftl.gradle

import org.gradle.api.DefaultTask
import org.gradle.api.tasks.TaskAction

abstract class FTLDeploy : DefaultTask() {
  @TaskAction
  fun deploy() {
    println("FTL deploy")
  }
}
