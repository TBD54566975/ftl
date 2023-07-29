buildscript {
  dependencies {
    classpath("com.squareup.wire:wire-gradle-plugin:4.7.2")
  }
}

group = "xyz.block"
version = "0.1.0-SNAPSHOT"

plugins {
  id("com.squareup.wire") version "4.7.2"
  kotlin("jvm") version "1.9.0"
  // Apply the java-library plugin for API and implementation separation.
  `java-library`
}

repositories {
  // Use Maven Central for resolving dependencies.
  mavenCentral()
}

dependencies {
  compileOnly("org.hotswapagent:hotswap-agent-core:1.4.1")

  // Use the Kotlin JUnit 5 integration.
  testImplementation("org.jetbrains.kotlin:kotlin-test-junit5")

  // Use the JUnit 5 integration.
  testImplementation("org.junit.jupiter:junit-jupiter-engine:5.9.2")
  testRuntimeOnly("org.junit.platform:junit-platform-launcher")

  // These dependencies are used internally, and not exposed to consumers on their own compile classpath.
  implementation("io.github.classgraph:classgraph:4.8.157")
  implementation("ch.qos.logback:logback-classic:1.4.5")
  implementation("ch.qos.logback:logback-core:1.4.5")
  implementation("org.jetbrains.kotlin:kotlin-reflect:1.8.22")
  implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.6.4")
  implementation("com.google.code.gson:gson:2.10.1")
  implementation("com.squareup.wire:wire-runtime:4.7.2")
  implementation("com.squareup.wire:wire-grpc-server:4.7.2")
  implementation("io.grpc:grpc-netty:1.56.1")
  implementation("io.grpc:grpc-protobuf:1.56.1")
  implementation("io.grpc:grpc-stub:1.56.1")
}

// Disable gradlew because we use a Hermit-provided gradle.
tasks.findByName("wrapper")?.enabled = false

wire {
  kotlin {
    rpcRole = "server"
    rpcCallStyle = "blocking"
    grpcServerCompatible = true
  }
  sourcePath {
    srcDir("src/main/proto")
  }
}

tasks.named<Test>("test") {
  // Use JUnit Platform for unit tests.
  useJUnitPlatform()
  testLogging {
    events("passed", "skipped", "failed")
  }
}

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
  from({
    configurations.runtimeClasspath.get().filter { it.name.endsWith("jar") }.map { zipTree(it) }
  }) {
    exclude("META-INF/*.RSA", "META-INF/*.SF", "META-INF/*.DSA")
  }
}