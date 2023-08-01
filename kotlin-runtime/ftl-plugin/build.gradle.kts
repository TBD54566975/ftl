plugins {
  kotlin("jvm") version "1.9.0"
  id("java-gradle-plugin")
}

repositories {
  // Use Maven Central for resolving dependencies.
  mavenCentral()
}

group = "xyz.block"
version = "1.0-SNAPSHOT"

gradlePlugin {
  plugins {
    create("ftl") {
      id = "xyz.block.ftl"
      displayName = "FTL"
      implementationClass = "xyz.block.ftl.gradle.FTLPlugin"
      description = "Generate FTL stubs"
    }
  }
}

dependencies {
  compileOnly(gradleApi())

  // Use the Kotlin JUnit 5 integration.
  testImplementation("org.jetbrains.kotlin:kotlin-test-junit5")

  // Use the JUnit 5 integration.
  testImplementation("org.junit.jupiter:junit-jupiter-engine:5.9.2")
  testRuntimeOnly("org.junit.platform:junit-platform-launcher")

  implementation(project(":ftl-protos"))
}

tasks.findByName("wrapper")?.enabled = false
