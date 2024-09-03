package xyz.block.ftl.intellij

import com.intellij.openapi.application.ApplicationManager

fun runOnEDT(runnable: () -> Unit) {
  ApplicationManager.getApplication().invokeLater {
    runnable()
  }
}
