repositories {
  flatDir {
    dirs("../../kotlin-runtime/ftl-plugin/build/libs")
  }
  mavenCentral()
}

plugins {
  kotlin("jvm") version "1.9.0"
  `java-library`
  id("xyz.block.ftl")
}

dependencies {
  implementation("xyz.block:ftl-runtime")
}

ftl {
  endpoint = "http://localhost:8892"
}

tasks.findByName("wrapper")?.enabled = false
