plugins {
  kotlin("jvm") version "1.9.0"
  // Apply the java-library plugin for API and implementation separation.
  `java-library`
}

repositories {
  // Use Maven Central for resolving dependencies.
  mavenCentral()
}

dependencies {
  implementation(project(":ftl-runtime"))
}

tasks.register<JavaExec>("run") {
  group = "Execution"
  description = "Run the module."
  classpath = sourceSets["main"].runtimeClasspath
  mainClass.set("xyz.block.ftl.main.MainKt")
}

tasks.findByName("wrapper")?.enabled = false
