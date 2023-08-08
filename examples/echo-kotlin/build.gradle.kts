plugins {
  kotlin("jvm") version "1.9.0"
  // Apply the java-library plugin for API and implementation separation.
  `java-library`
  id("xyz.block.ftl")
}

repositories {
  // Use Maven Central for resolving dependencies.
  mavenCentral()
}

dependencies {
  implementation("xyz.block.ftl:ftl-runtime")
}

ftl {
  endpoint = "http://localhost:8892"
}

tasks.findByName("wrapper")?.enabled = false
