package com.squareup.ftldemo

import xyz.block.ftl.Context
import xyz.block.ftl.Ftl
import xyz.block.ftl.SimpleVerb
import xyz.block.ftl.Verb
import java.util.UUID

data class PaymentRequest(val recipient: String, val amount: Long)

data class PaymentResponse(val id: String, val status: String)

class PaymentVerb : SimpleVerb() {
  @Verb fun pay(context: Context, request: PaymentRequest): PaymentResponse {
    Ftl.call(
      context,
      NotifyVerb::notifyCustomer,
      NotifyRequest(request.recipient, 200_00L)
    )

    return PaymentResponse(UUID.randomUUID().toString(), "INITIATED")
  }
}
