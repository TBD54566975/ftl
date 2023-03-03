package xyz.block.ftl.drive.adapter

import com.google.gson.Gson
import java.io.Reader
import java.io.Writer
import kotlin.reflect.KClass

class JsonAdapter {
  private val gson = Gson()

  fun readAs(reader: Reader, dataType: KClass<*>): Any {
    return gson.fromJson<Any>(reader, dataType.java)
  }

  fun write(any: Any, writer: Writer) {
    writer.write(gson.toJson(any))
  }
}
