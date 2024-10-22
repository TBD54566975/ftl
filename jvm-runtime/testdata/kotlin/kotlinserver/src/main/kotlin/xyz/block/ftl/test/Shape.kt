package xyz.block.ftl.test

import xyz.block.ftl.Enum

@Enum
enum class Shape(val value: String) {
  Circle("circle"),
  Square("square"),
  Triangle("triangle");
}
