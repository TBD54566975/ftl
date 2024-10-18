package xyz.block.ftl.test

class Cat(val furLength: Int, val name: String, val breed: String): Animal {
  override fun isCat() = true

  override fun isDog() = false

  override fun getCat() = this

  override fun getDog(): Dog = throw RuntimeException("Cat is not Dog")
}
