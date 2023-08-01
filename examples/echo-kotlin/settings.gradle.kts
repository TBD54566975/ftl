plugins {}

rootProject.name = "echo"
includeBuild("../../kotlin-runtime") {
  dependencySubstitution {
    substitute(module("xyz.block.ftl:ftl-runtime")).using(project(":"))
  }
}

includeBuild("../../kotlin-runtime/ftl-plugin")
include(":ftl-protos")
project(":ftl-protos").projectDir = File("../../kotlin-runtime/ftl-protos")
