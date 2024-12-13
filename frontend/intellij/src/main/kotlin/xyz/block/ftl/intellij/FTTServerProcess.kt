package xyz.block.ftl.intellij

import com.intellij.execution.configurations.GeneralCommandLine
import com.redhat.devtools.lsp4ij.server.OSProcessStreamConnectionProvider

class FTTServerProcess(commandLine: GeneralCommandLine?) : OSProcessStreamConnectionProvider(commandLine) {
}
