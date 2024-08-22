package xyz.block.ftl.intellij.toolWindow

import com.intellij.execution.filters.TextConsoleBuilderFactory
import com.intellij.execution.ui.ConsoleView
import com.intellij.execution.ui.ConsoleViewContentType
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.SimpleToolWindowPanel
import xyz.block.ftl.intellij.runOnEDT
import java.awt.BorderLayout
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter

class FTLMessagesPanel(project: Project) : SimpleToolWindowPanel(false, false) {
  val consoleView: ConsoleView = TextConsoleBuilderFactory.getInstance().createBuilder(project).console
  var autoScrollEnabled = true

  init {
    layout = BorderLayout()
    add(consoleView.component, BorderLayout.CENTER)
  }

  fun addMessage(message: String) {
    runOnEDT {
      val timestamp = ZonedDateTime.now().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME)
      val messageWithTimestamp = "[$timestamp] $message\n"
      consoleView.print(messageWithTimestamp, ConsoleViewContentType.NORMAL_OUTPUT)

      if (autoScrollEnabled) {
        consoleView.requestScrollingToEnd()
      }
    }
  }
}
