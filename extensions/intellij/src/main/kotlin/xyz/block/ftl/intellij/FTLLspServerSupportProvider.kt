package xyz.block.ftl.intellij

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.execution.process.OSProcessHandler
import com.intellij.icons.AllIcons.Icons
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.platform.lsp.api.LspServer
import com.intellij.platform.lsp.api.LspServerDescriptor.Companion.LOG
import com.intellij.platform.lsp.api.LspServerManager
import com.intellij.platform.lsp.api.LspServerManagerListener
import com.intellij.platform.lsp.api.LspServerState
import com.intellij.platform.lsp.api.LspServerSupportProvider
import com.intellij.platform.lsp.api.lsWidget.LspServerWidgetItem
import com.intellij.util.io.BaseOutputReader
import com.intellij.util.messages.Topic
import xyz.block.ftl.intellij.toolWindow.FTLMessagesToolWindowFactory.Util.displayMessageInToolWindow
import java.util.regex.Pattern

interface FTLLSPNotifier {
  fun lspServerStateChange(state: LspServerState)

  companion object {
    @Topic.ProjectLevel
    val SERVER_STATE_CHANGE_TOPIC: Topic<FTLLSPNotifier> = Topic.create(
      "FTL Server State Changed",
      FTLLSPNotifier::class.java
    )
  }
}

class FTLLspServerSupportProvider : LspServerSupportProvider {
  private var listenerAdded: Boolean = false

  override fun createLspServerWidgetItem(lspServer: LspServer, currentFile: VirtualFile?): LspServerWidgetItem =
    LspServerWidgetItem(
      lspServer = lspServer,
      currentFile = currentFile,
      settingsPageClass = FTLSettingsConfigurable::class.java,
      widgetMainActionBaseIcon = Icons.Ide.MenuArrow
    )

  override fun fileOpened(
    project: Project,
    file: VirtualFile,
    serverStarter: LspServerSupportProvider.LspServerStarter
  ) {
    if (!listenerAdded) {
      try {
        listenerAdded = true
        val lspServerManager = LspServerManager.getInstance(project)
        lspServerManager.addLspServerManagerListener(listener = object : LspServerManagerListener {
          override fun serverStateChanged(lspServer: LspServer) {
            val publisher = project.messageBus.syncPublisher(FTLLSPNotifier.SERVER_STATE_CHANGE_TOPIC)
            publisher.lspServerStateChange(lspServer.state)
          }
        }, parentDisposable = { }, sendEventsForExistingServers = true)
      } catch (e: Exception) {
        listenerAdded = false
      }
    }

    val isFtlSupportLanguage = file.extension == "go" || file.extension == "kt" || file.extension == "java"
    if (isFtlSupportLanguage && hasFtlProjectFile(project)) {
      serverStarter.ensureServerStarted(FTLLspServerDescriptor(project))
    }
  }

  private fun hasFtlProjectFile(project: Project): Boolean {
    val projectBaseDir = project.baseDir ?: return false
    val ftlProjectFile = projectBaseDir.findChild("ftl-project.toml")
    return ftlProjectFile != null && ftlProjectFile.exists()
  }

  fun startLspServer(project: Project) {
    val lspServerManager = LspServerManager.getInstance(project)
    lspServerManager.startServersIfNeeded(FTLLspServerSupportProvider::class.java)
  }

  fun stopLspServer(project: Project): OSProcessHandler? {
    return when (getLspServerStatus(project)) {
      LspServerState.ShutdownUnexpectedly -> {
        stopViaCommand(project)
      }

      else -> {
        val lspServerManager = LspServerManager.getInstance(project)
        lspServerManager.stopServers(FTLLspServerSupportProvider::class.java)
        null
      }
    }
  }

  private fun stopViaCommand(project: Project): OSProcessHandler {
    val settings = AppSettings.getInstance().state
    val generalCommandLine =
      GeneralCommandLine(listOf(settings.lspServerPath) + settings.lspServerStopArguments.split(Pattern.compile("\\s+"))).withCharset(
        Charsets.UTF_8
      )
    generalCommandLine.setWorkDirectory(project.basePath)
    displayMessageInToolWindow(project, "LSP Server Command: " + generalCommandLine.commandLineString)
    displayMessageInToolWindow(project, "Working Directory: " + generalCommandLine.workDirectory)

    LOG.info("$this: stopping LSP server: $generalCommandLine")
    val process: OSProcessHandler = object : OSProcessHandler(generalCommandLine) {
      override fun readerOptions(): BaseOutputReader.Options = BaseOutputReader.Options.forMostlySilentProcess()
    }

    return process
  }

  fun restartLspServer(project: Project) {
    val lspServerManager = LspServerManager.getInstance(project)
    lspServerManager.stopAndRestartIfNeeded(FTLLspServerSupportProvider::class.java)
  }

  fun getLspServerStatus(project: Project): LspServerState {
    val lspServerManager = LspServerManager.getInstance(project)
    val server = lspServerManager.getServersForProvider(FTLLspServerSupportProvider::class.java).firstOrNull()

    return server?.state ?: LspServerState.ShutdownNormally
  }
}
