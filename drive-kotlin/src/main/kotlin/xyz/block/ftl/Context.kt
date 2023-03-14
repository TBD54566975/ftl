package xyz.block.ftl

import jakarta.servlet.http.HttpServletRequest

class Context {
  companion object {
    fun fromHttpRequest(request: HttpServletRequest): Context {
      return Context()
    }

    fun fromLocal() = Context()
  }
}
