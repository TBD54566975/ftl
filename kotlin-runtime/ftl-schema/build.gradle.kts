plugins {
  kotlin("jvm") version "1.9.0"
  id("com.squareup.wire") version "4.7.2"
  `java-library`
}

group = "xyz.block"
version = "1.0-SNAPSHOT"

buildscript {
  dependencies {
    classpath(kotlin("gradle-plugin", version = "1.9.0"))
  }
}

tasks.findByName("wrapper")?.enabled = false

repositories {
  mavenCentral()
}

dependencies {
  implementation("com.google.devtools.ksp:symbol-processing-api:1.9.0-1.0.11")
  implementation(libs.wireGrpcClient)
  implementation("xyz.block.ftl:ftl-runtime")
  testImplementation(kotlin("test"))
}

tasks.test {
  useJUnitPlatform()
}

wire {
  kotlin {
    rpcRole = "client"
    rpcCallStyle = "blocking"
  }
  sourcePath {
    srcDir("../../protos")
  }
}
