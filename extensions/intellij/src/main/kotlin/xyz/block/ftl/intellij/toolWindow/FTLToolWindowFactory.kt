package xyz.block.ftl.intellij.toolWindow

import com.intellij.execution.process.ProcessAdapter
import com.intellij.execution.process.ProcessEvent
import com.intellij.execution.ui.ConsoleView
import com.intellij.icons.AllIcons
import com.intellij.openapi.actionSystem.ActionManager
import com.intellij.openapi.actionSystem.ActionToolbar
import com.intellij.openapi.actionSystem.ActionUpdateThread
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.DefaultActionGroup
import com.intellij.openapi.actionSystem.ToggleAction
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.editor.ex.EditorEx
import com.intellij.openapi.options.ShowSettingsUtil
import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Key
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.openapi.wm.ToolWindowManager
import com.intellij.platform.lsp.api.LspServerState
import com.intellij.ui.content.ContentFactory
import com.intellij.ui.content.ContentManagerEvent
import com.intellij.ui.content.ContentManagerListener
import xyz.block.ftl.intellij.FTLLSPNotifier
import xyz.block.ftl.intellij.FTLLspServerService
import xyz.block.ftl.intellij.FTLSettingsConfigurable
import java.util.concurrent.Executors
import java.util.concurrent.ScheduledExecutorService

class FTLMessagesToolWindowFactory() : ToolWindowFactory, DumbAware {
  private var currentLspState: LspServerState = LspServerState.Initializing
  val scheduler: ScheduledExecutorService = Executors.newSingleThreadScheduledExecutor()

  private lateinit var startAction: AnAction
  private lateinit var stopAction: AnAction

  private lateinit var panel: FTLMessagesPanel

  override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
    panel = FTLMessagesPanel(project)

    val contentFactory = ContentFactory.getInstance()
    val actionManager = ActionManager.getInstance()
    val content = contentFactory.createContent(panel, "", false)

    val actionGroup = DefaultActionGroup().apply {

      startAction = object : DumbAwareAction("Start", "Start the process", AllIcons.Actions.Execute) {
        override fun actionPerformed(e: AnActionEvent) {
          panel.addMessage("Start action triggered")

          val service = FTLLspServerService.getInstance(project)
          panel.addMessage("Status is: ${service.lspServerSupportProvider.getLspServerStatus(project)}")
          service.lspServerSupportProvider.startLspServer(project)
        }

        override fun update(e: AnActionEvent) {
          e.presentation.isEnabled =
            LspServerState.Running != currentLspState && LspServerState.Initializing != currentLspState
        }

        override fun getActionUpdateThread(): ActionUpdateThread {
          return ActionUpdateThread.EDT
        }
      }
      add(startAction)

      stopAction = object : DumbAwareAction("Stop", "Stop the process", AllIcons.Actions.Suspend) {
        override fun actionPerformed(e: AnActionEvent) {
          panel.addMessage("Stop action triggered")
          val service = FTLLspServerService.getInstance(project)
          panel.addMessage("Status is: ${service.lspServerSupportProvider.getLspServerStatus(project)}")
          val processHandler = service.lspServerSupportProvider.stopLspServer(project)
          processHandler?.addProcessListener(object : ProcessAdapter() {
            override fun onTextAvailable(event: ProcessEvent, outputType: Key<*>) {
              val message = event.text.trim()
              if (message.isNotBlank()) {
                Util.displayMessageInToolWindow(project, message)
              }
            }
          })
        }

        override fun getActionUpdateThread(): ActionUpdateThread {
          return ActionUpdateThread.EDT
        }

        override fun update(e: AnActionEvent) {
          e.presentation.isEnabled = LspServerState.ShutdownNormally != currentLspState
        }
      }
      add(stopAction)

      add(object : DumbAwareAction("Restart", "Restart the process", AllIcons.Actions.Restart) {
        override fun actionPerformed(e: AnActionEvent) {
          panel.addMessage("Restart action triggered")
          val service = FTLLspServerService.getInstance(project)
          panel.addMessage("Status is: ${service.lspServerSupportProvider.getLspServerStatus(project)}")
          service.lspServerSupportProvider.restartLspServer(project)
        }

        override fun getActionUpdateThread(): ActionUpdateThread {
          return ActionUpdateThread.EDT
        }
      })

      add(object : AnAction("Toggle Soft Wrap", "Toggle soft wrap in console view", AllIcons.Actions.ToggleSoftWrap) {
        override fun actionPerformed(e: AnActionEvent) {
          val editor = getEditorFromConsoleView(panel.consoleView)
          if (editor != null) {
            val settings = editor.settings
            settings.isUseSoftWraps = !settings.isUseSoftWraps
          }
        }

        override fun getActionUpdateThread(): ActionUpdateThread {
          return ActionUpdateThread.EDT
        }

        private fun getEditorFromConsoleView(consoleView: ConsoleView): EditorEx? {
          return try {
            val method = consoleView.javaClass.getMethod("getEditor")
            method.isAccessible = true
            method.invoke(consoleView) as? EditorEx
          } catch (e: Exception) {
            e.printStackTrace()
            null
          }
        }

        override fun update(e: AnActionEvent) {
          val editor = getEditorFromConsoleView(panel.consoleView)
          if (editor != null) {
            val isSoftWrapEnabled = editor.settings.isUseSoftWraps
            e.presentation.isEnabled = true
            e.presentation.putClientProperty("selected", isSoftWrapEnabled)
          } else {
            e.presentation.isEnabled = false
          }
        }
      })

      add(object : AnAction("Clear", "Clear the console", AllIcons.Actions.GC) {
        override fun actionPerformed(e: AnActionEvent) {
          panel.consoleView.clear()
        }

        override fun getActionUpdateThread(): ActionUpdateThread {
          return ActionUpdateThread.EDT
        }
      })

      add(object : ToggleAction("Auto Scroll", "Toggle auto scroll", AllIcons.RunConfigurations.Scroll_down) {
        override fun isSelected(e: AnActionEvent): Boolean {
          return panel.autoScrollEnabled
        }

        override fun getActionUpdateThread(): ActionUpdateThread {
          return ActionUpdateThread.EDT
        }

        override fun setSelected(e: AnActionEvent, state: Boolean) {
          panel.autoScrollEnabled = state
        }
      })

      add(object : AnAction("Settings", "Open Settings", AllIcons.General.Settings) {
        override fun getActionUpdateThread(): ActionUpdateThread {
          return ActionUpdateThread.EDT
        }

        override fun actionPerformed(e: AnActionEvent) {
          ShowSettingsUtil.getInstance().showSettingsDialog(project, FTLSettingsConfigurable::class.java)
        }
      })
    }

    val actionToolbar: ActionToolbar = actionManager.createActionToolbar("FTLToolbar", actionGroup, true)

    actionToolbar.targetComponent = panel.toolbar
    panel.toolbar = actionToolbar.component
    toolWindow.contentManager.addContent(content)

    project.messageBus.connect().subscribe(
      FTLLSPNotifier.SERVER_STATE_CHANGE_TOPIC,
      object : FTLLSPNotifier {
        override fun lspServerStateChange(state: LspServerState) {
          currentLspState = state
          panel.addMessage("State changed: ${state}")
        }
      })
  }

  override fun init(toolWindow: ToolWindow) {
    toolWindow.contentManager.addContentManagerListener(object : ContentManagerListener {
      override fun contentRemoveQuery(event: ContentManagerEvent) {
        scheduler.shutdown()
        super.contentRemoveQuery(event)
      }
    })
  }

  object Util {
    fun displayMessageInToolWindow(project: Project, message: String) {
      ApplicationManager.getApplication().invokeLater {
        val toolWindow = ToolWindowManager.getInstance(project).getToolWindow("FTL")
        if (toolWindow != null) {
          val content = toolWindow.contentManager.getContent(0)
          val panel = content?.component as? FTLMessagesPanel
          panel?.addMessage(message)
        }
      }
    }
  }
}

