package xyz.block.ftl.drive

import com.squareup.ftldemo.Order
import com.squareup.ftldemo.makePizza
import jakarta.servlet.http.HttpServlet
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import jakarta.servlet.http.HttpServletResponse.SC_OK
import xyz.block.ftl.drive.verb.VerbCassette

class DriveServlet : HttpServlet() {
  private val cassette = VerbCassette(::makePizza)

  override fun doGet(request: HttpServletRequest?, response: HttpServletResponse?) {
    response!!.apply {
      contentType = "text/html"
      status = SC_OK

      // Use format adapters here to translate the request
      writer.println(cassette.invokeVerb(Order(request!!.getParameter("topping"))))
    }
  }
}