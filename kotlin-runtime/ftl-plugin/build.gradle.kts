buildscript {
  dependencies {
    classpath("com.squareup.wire:wire-gradle-plugin:4.7.2")
  }
}

plugins {
  kotlin("jvm") version "1.9.0"
  id("java-gradle-plugin")
  id("com.squareup.wire") version "4.7.2"
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
  testImplementation(libs.kotlinTestJunit5)

  // Use the JUnit 5 integration.
  testImplementation(libs.junitJupiterEngine)
  testRuntimeOnly(libs.junitPlatformLauncher)

  implementation(libs.kotlinPoet)
  implementation(libs.kotlinReflect)
  implementation(libs.kotlinxCoroutinesCore)
  implementation(libs.wireRuntime)
  implementation(libs.wireGrpcClient)
}

tasks.findByName("wrapper")?.enabled = false

wire {
  kotlin {
    rpcRole = "client"
    rpcCallStyle = "blocking"
  }
  sourcePath {
    srcDir("../../protos")
  }
}
