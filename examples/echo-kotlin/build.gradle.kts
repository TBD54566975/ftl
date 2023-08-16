plugins {
  kotlin("jvm") version "1.9.0"
  // Apply the java-library plugin for API and implementation separation.
  `java-library`
  id("xyz.block.ftl")
  id("com.google.devtools.ksp") version "1.9.0-1.0.11"
}

repositories {
  // Use Maven Central for resolving dependencies.
  mavenCentral()
}

dependencies {
  implementation("xyz.block.ftl:ftl-runtime")
  implementation("xyz.block.ftl:ftl-schema")
  ksp("xyz.block.ftl:ftl-schema")
}

ftl {
  endpoint = "http://localhost:8892"
}

tasks.findByName("wrapper")?.enabled = false
