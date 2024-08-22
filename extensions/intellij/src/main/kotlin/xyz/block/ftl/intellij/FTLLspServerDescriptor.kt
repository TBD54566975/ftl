package xyz.block.ftl.intellij

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.execution.process.OSProcessHandler
import com.intellij.execution.process.ProcessAdapter
import com.intellij.execution.process.ProcessEvent
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Key
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.platform.lsp.api.LspServerNotificationsHandler
import com.intellij.platform.lsp.api.ProjectWideLspServerDescriptor
import xyz.block.ftl.intellij.toolWindow.FTLMessagesToolWindowFactory
import java.util.regex.Pattern

class FTLLspServerDescriptor(project: Project) : ProjectWideLspServerDescriptor(project, "FTL") {
  override fun isSupportedFile(file: VirtualFile) = file.extension == "go"

  override fun createLsp4jClient(handler: LspServerNotificationsHandler): CustomLsp4jClient {
    return CustomLsp4jClient(handler)
  }

  override fun createCommandLine(): GeneralCommandLine {
    val settings = AppSettings.getInstance().state
    val generalCommandLine =
      GeneralCommandLine(listOf(settings.lspServerPath) + settings.lspServerArguments.split(Pattern.compile("\\s+")))
    generalCommandLine.setWorkDirectory(project.basePath)
    displayMessageInToolWindow("LSP Server Command: " + generalCommandLine.commandLineString)
    displayMessageInToolWindow("Working Directory: " + generalCommandLine.workDirectory)
    return generalCommandLine
  }

  override fun startServerProcess(): OSProcessHandler {
    displayMessageInToolWindow("Starting FTL LSP Server")
    val processHandler = super.startServerProcess()
    processHandler.addProcessListener(object : ProcessAdapter() {
      override fun onTextAvailable(event: ProcessEvent, outputType: Key<*>) {
        val message = event.text.trim()
        if (message.isNotBlank()) {
          displayMessageInToolWindow(message)
        }
      }
    })
    return processHandler
  }

  private fun displayMessageInToolWindow(message: String) {
    FTLMessagesToolWindowFactory.Util.displayMessageInToolWindow(project, message)
  }
}
