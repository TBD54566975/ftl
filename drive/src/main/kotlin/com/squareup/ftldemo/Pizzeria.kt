package com.squareup.ftldemo

import xyz.block.ftl.drive.Verb

data class Order(val topping: String)

data class Pizza(val base: String, val size: String, val topping: String)

@Verb fun makePizza(order: Order): Pizza {
  return Pizza(base = "wholewheat", size = "Large", topping = order.topping)
}