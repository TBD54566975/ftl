package xyz.block.ftl.intellij

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.execution.process.OSProcessHandler
import com.intellij.execution.process.ProcessAdapter
import com.intellij.execution.process.ProcessEvent
import com.intellij.ide.DataManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Key
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.openapi.wm.ToolWindowManager
import com.intellij.platform.lsp.api.LspServerNotificationsHandler
import com.intellij.platform.lsp.api.ProjectWideLspServerDescriptor
import com.intellij.tools.ToolsCustomizer
import xyz.block.ftl.intellij.toolWindow.FTLMessagesToolWindowFactory
import java.util.concurrent.CompletableFuture
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
    try {
      // Hermit support, we need to get the environment variables so we use the correct FTL
      val result = CompletableFuture<GeneralCommandLine>()
      runOnEDT {
        val toolWindow = ToolWindowManager.getInstance(project).getToolWindow("FTL")
        if (toolWindow != null) {
          val dataContext = DataManager.getInstance().getDataContext(toolWindow.component)
          val customizeCommandLine =
            ToolsCustomizer.customizeCommandLine(generalCommandLine, dataContext)
          result.complete(customizeCommandLine)
        }
      }
      val res = result.get()
      return if (res != null) res else generalCommandLine
    } catch (e: Exception) {
      displayMessageInToolWindow("Failed to customize LSP Server Command: " + e.message)
    }
    return generalCommandLine
  }

  override fun startServerProcess(): OSProcessHandler {
    displayMessageInToolWindow("Starting FTL LSP Server")
    val processHandler = super.startServerProcess()
    processHandler.addProcessListener(object : ProcessAdapter() {

      override fun startNotified(event: ProcessEvent) {
        super.startNotified(event)
        displayMessageInToolWindow("LSP Started")
      }

      override fun processTerminated(event: ProcessEvent) {
        super.processTerminated(event)
        displayMessageInToolWindow("LSP Terminated")
      }

      override fun processWillTerminate(event: ProcessEvent, willBeDestroyed: Boolean) {
        super.processWillTerminate(event, willBeDestroyed)
        displayMessageInToolWindow("LSP Will Terminate")
      }

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
