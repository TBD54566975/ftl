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

version "17.0.6.b469.82" {
  platform "darwin" "arm64" {
    source = "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-osx-aarch64-fastdebug-b469.82.tar.gz"
  }

  platform "darwin" "amd64" {
    source = "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-osx-x64-fastdebug-b469.82.tar.gz"
  }

  platform "linux" "amd64" {
    source = "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-linux-x64-fastdebug-b469.82.tar.gz"
  }
}

sha256sums = {
  "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-linux-x64-fastdebug-b469.82.tar.gz": "636675c1372d0df8a5461eac1c4938e1383fba19dcfb8d879b8533ede6dee354",
  "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-osx-x64-fastdebug-b469.82.tar.gz": "a6ea022c5b26c471519d6e31a6c205327d8b1bf18005025a92b7f0ae63d97473",
  "https://cache-redirector.jetbrains.com/intellij-jbr/jbrsdk-17.0.6-osx-aarch64-fastdebug-b469.82.tar.gz": "280ce36d97589ff789c8655916e507a9e64ea715b3ef3db75462bf2ee4fbba79",
}
