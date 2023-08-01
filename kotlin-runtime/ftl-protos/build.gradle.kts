buildscript {
  dependencies {
    classpath("com.squareup.wire:wire-gradle-plugin:4.7.2")
  }
}

plugins {
  id("com.squareup.wire") version "4.7.2"
  kotlin("jvm") version "1.9.0"
}

group = "xyz.block"
version = "1.0-SNAPSHOT"

repositories {
  mavenCentral()
}

dependencies {
  implementation(libs.wireRuntime)
  implementation(libs.wireGrpcServer)
  implementation(libs.wireGrpcClient)
  implementation(libs.grpcNetty)
  implementation(libs.grpcProtobuf)
  implementation(libs.grpcStub)
}

// Disable gradlew because we use a Hermit-provided gradle.
tasks.findByName("wrapper")?.enabled = false

wire {
  kotlin {
    rpcRole = "server"
    rpcCallStyle = "blocking"
    grpcServerCompatible = true
    includes = listOf(
      "xyz.block.ftl.v1.ControllerService",
      "xyz.block.ftl.v1.RunnerService",
      "xyz.block.ftl.v1.VerbService"
    )
    exclusive = false
    out = "src/main/kotlin/generated"
  }
  kotlin {
    rpcRole = "client"
    rpcCallStyle = "blocking"
    out = "src/main/kotlin/generated"
  }
  sourcePath {
    srcDir("../../protos")
  }
}
