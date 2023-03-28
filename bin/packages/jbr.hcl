description = "Runtime environment based on OpenJDK for running IntelliJ Platform-based products on Windows, macOS, and Linux"
binaries = ["bin/*"]
provides = ["jdk", "jre"]
env = {
  "JAVA_HOME": "${root}",
}
test = "java -version"
strip = 1

platform "darwin" {
  root = "${dest}/Contents/Home"
  dest = "${HOME}/Library/Java/JavaVirtualMachines/jbr-${version}.jdk"
}

on unpack {
  copy { from = "jbr/post-unpack.sh" to = "${root}/post-unpack.sh" mode = 0750 }
  run { cmd = "${root}/post-unpack.sh ${dest}" }
}

version "17.0.6.b469.82" {
  platform "darwin" "arm64" {
    source = "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-osx-aarch64-b469.82.tar.gz"
  }

  platform "darwin" "amd64" {
    source = "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-osx-x64-b469.82.tar.gz"
  }

  platform "linux" "amd64" {
    source = "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-linux-x64-b469.82.tar.gz"
  }
}

sha256sums = {
  "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-osx-x64-b469.82.tar.gz": "ed073e407d7bef4634d3a1bfe79b70560db50a1a74e94c2aca80bc7c7666b951",
  "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-osx-aarch64-b469.82.tar.gz": "7abd6420edfc89d3c197aa50a100488452418afd4b2fe5701be43fed3eacc5b2",
  "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-linux-x64-b469.82.tar.gz": "fb93918c7a8acb56ba3a9aafd7453294d63677914486d20e427111eaa18f05d1",
}
