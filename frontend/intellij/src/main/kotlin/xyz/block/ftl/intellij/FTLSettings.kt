package xyz.block.ftl.intellij

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.Service
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import org.jetbrains.annotations.NonNls

@State(
  name = "org.intellij.sdk.settings.AppSettings",
  storages = [Storage("SdkSettingsPlugin.xml")]
)
@Service
class AppSettings : PersistentStateComponent<AppSettings.State> {

  data class State(
    @NonNls var lspServerPath: String = "ftl",
    var lspServerArguments: String = "--lsp",
    var lspServerStopArguments: String = "serve --stop",
    var autoRestartLspServer: Boolean = false,
  )

  private var myState = State()

  companion object {
    fun getInstance(): AppSettings {
      return ApplicationManager.getApplication().getService(AppSettings::class.java)
    }
  }

  override fun getState(): State {
    return myState
  }

  override fun loadState(state: State) {
    myState = state
  }
}
