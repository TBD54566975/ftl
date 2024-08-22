package xyz.block.ftl.intellij

import com.intellij.openapi.options.Configurable
import org.jetbrains.annotations.Nls
import javax.swing.JComponent

class FTLSettingsConfigurable : Configurable {

  private var mySettingsComponent: FTLSettingsComponent? = null

  @Nls(capitalization = Nls.Capitalization.Title)
  override fun getDisplayName(): String {
    return "FTL"
  }

  override fun getPreferredFocusedComponent(): JComponent? {
    return mySettingsComponent?.getPreferredFocusedComponent()
  }

  override fun createComponent(): JComponent? {
    mySettingsComponent = FTLSettingsComponent()
    return mySettingsComponent?.getPanel()
  }

  override fun isModified(): Boolean {
    val state = AppSettings.getInstance().state
    return mySettingsComponent?.getLspServerPath() != state.lspServerPath ||
      mySettingsComponent?.getLspServerArguments() != state.lspServerArguments ||
      mySettingsComponent?.getLspServerStopArguments() != state.lspServerStopArguments
  }

  override fun apply() {
    val state = AppSettings.getInstance().state
    state.lspServerPath = mySettingsComponent?.getLspServerPath() ?: "ftl"
    state.lspServerArguments = mySettingsComponent?.getLspServerArguments() ?: "--recreate --lsp"
    state.lspServerStopArguments = mySettingsComponent?.getLspServerStopArguments() ?: "serve --stop"
  }

  override fun reset() {
    val state = AppSettings.getInstance().state
    mySettingsComponent?.setLspServerPath(state.lspServerPath)
    mySettingsComponent?.setLspServerArguments(state.lspServerArguments)
    mySettingsComponent?.setLspServerStopArguments(state.lspServerStopArguments)
  }

  override fun disposeUIResources() {
    mySettingsComponent = null
  }
}
