package xyz.block.ftl.intellij

import com.intellij.openapi.components.Service
import com.intellij.openapi.project.Project

@Service(Service.Level.PROJECT)
class FTLLspServerService(val project: Project) {
  val lspServerSupportProvider = FTLLspServerSupportProvider()

  companion object {
    fun getInstance(project: Project): FTLLspServerService {
      return project.getService(FTLLspServerService::class.java)
    }
  }
}
