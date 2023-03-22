package com.squareup.ftldemo

import xyz.block.ftl.SimpleVerb
import xyz.block.ftl.Verb

data class NotifyRequest(val recipient: String, val amount: Long)

data class NotifyResponse(val message: String)

class NotifyVerb : SimpleVerb() {
  @Verb fun notifyCustomer(notify: NotifyRequest): NotifyResponse {
    println("Sending Push notification to ${notify.recipient}")
    return NotifyResponse("OK")
  }
}
