plugins {}

rootProject.name = "echo"
includeBuild("../../kotlin-runtime/kotlin-runtime") {
  dependencySubstitution {
    substitute(module("xyz.block.ftl:ftl-runtime")).using(project(":"))
  }
}

includeBuild("../../kotlin-runtime/ftl-plugin")
