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

tasks.jar {
  enabled = true
  isZip64 = true
  duplicatesStrategy = DuplicatesStrategy.EXCLUDE

  archiveFileName.set("${project.name}.jar")

  manifest {
    attributes["Main-Class"] = "xyz.block.ftl.main.MainKt"
  }

  from(sourceSets.main.get().output)
  dependsOn(configurations.compileClasspath)
  from(configurations.runtimeClasspath.get().map { if (it.isDirectory) it else zipTree(it) }) {
    exclude("META-INF/*.RSA", "META-INF/*.SF", "META-INF/*.DSA")
  }
}

tasks.named("jar") {
  dependsOn(":ftl-runtime:jar")
}
