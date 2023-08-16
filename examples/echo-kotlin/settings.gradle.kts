plugins {}

rootProject.name = "echo"

includeBuild("../../kotlin-runtime/ftl-runtime") {
  dependencySubstitution {
    substitute(module("xyz.block.ftl:ftl-runtime")).using(project(":"))
  }
}

includeBuild("../../kotlin-runtime/ftl-plugin")

includeBuild("../../kotlin-runtime/ftl-schema") {
  dependencySubstitution {
    substitute(module("xyz.block.ftl:ftl-schema")).using(project(":"))
  }
}
