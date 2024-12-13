package xyz.block.ftl.intellij

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.ide.DataManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindowManager
import com.intellij.tools.ToolsCustomizer
import com.redhat.devtools.lsp4ij.LanguageServerFactory
import com.redhat.devtools.lsp4ij.server.StreamConnectionProvider
import xyz.block.ftl.intellij.toolWindow.FTLMessagesToolWindowFactory
import java.util.concurrent.CompletableFuture
import java.util.regex.Pattern

class FTLLanguageServerFactory : LanguageServerFactory {
  override fun createConnectionProvider(project: Project): StreamConnectionProvider {
    return FTTServerProcess(createCommandLine(project))
  }


  fun createCommandLine(project: Project): GeneralCommandLine {
    val settings = AppSettings.getInstance().state
    val generalCommandLine =
      GeneralCommandLine(listOf(settings.lspServerPath) + settings.lspServerArguments.split(Pattern.compile("\\s+")))
    generalCommandLine.setWorkDirectory(project.basePath)
    displayMessageInToolWindow(project, "LSP Server Command: " + generalCommandLine.commandLineString)
    displayMessageInToolWindow(project, "Working Directory: " + generalCommandLine.workDirectory)
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
      displayMessageInToolWindow(project, "Failed to customize LSP Server Command: " + e.message)
    }
    return generalCommandLine
  }

  private fun displayMessageInToolWindow(project: Project, message: String) {
    FTLMessagesToolWindowFactory.Util.displayMessageInToolWindow(project, message)
  }
}
