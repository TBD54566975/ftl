package xyz.block.ftl.intellij

import com.intellij.platform.lsp.api.Lsp4jClient
import com.intellij.platform.lsp.api.LspServerNotificationsHandler

class CustomLsp4jClient(handler: LspServerNotificationsHandler) : Lsp4jClient(handler) {
  override fun telemetryEvent(`object`: Any) {
    super.telemetryEvent(`object`)
  }
}
