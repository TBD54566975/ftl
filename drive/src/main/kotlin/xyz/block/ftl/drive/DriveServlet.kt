package xyz.block.ftl.drive

import jakarta.servlet.http.HttpServlet
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import jakarta.servlet.http.HttpServletResponse.SC_OK

class DriveServlet : HttpServlet() {
  override fun doGet(req: HttpServletRequest?, response: HttpServletResponse?) {
    response!!.apply {
      contentType = "text/html"
      status = SC_OK
      writer.println("Faster Than Light!")
      
      // invoke Verb function here.
    }
  }
}