rootProject.name = "ftl-schema"

dependencyResolutionManagement {
  versionCatalogs {
    create("libs") {
      from(files("../gradle/libs.versions.toml"))
    }
  }
}

includeBuild("../ftl-runtime") {
  dependencySubstitution {
    substitute(module("xyz.block.ftl:ftl-runtime")).using(project(":"))
  }
}
