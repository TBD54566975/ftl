package xyz.block.ftl.registry

import com.google.gson.Gson
import com.google.gson.GsonBuilder
import com.google.gson.JsonDeserializer
import com.google.gson.JsonPrimitive
import com.google.gson.JsonSerializer
import org.junit.jupiter.api.Test
import java.time.Instant
import java.time.OffsetDateTime

class VerbHandleTest {
  val gson = Gson()

  @Test
  fun testInvoke() {
    // val requestJson = gson.toJson(VerbRequest("a"))
    // val handle = VerbHandle(
    //   verbClass = ExampleVerb::class,
    //   verbFunction = ExampleVerb::verb,
    // )
    // val response = handle.invokeVerbInternal(
    //   context = Context(),
    //   argument = requestJson,
    // )
    // assertEquals(gson.toJson(VerbResponse("test")), response)
  }

  @Test
  fun testTimeSerializer() {
    val gson = GsonBuilder()
      .registerTypeAdapter(
        Instant::class.java, JsonSerializer<Instant> { src, _, _ -> JsonPrimitive(src.toString()) })
      .registerTypeAdapter(Instant::class.java, JsonDeserializer<Instant> { json, _, _ ->
        println(json)
        val offsetDateTime = OffsetDateTime.parse(json.asString)
        offsetDateTime.toInstant()
      })
      .create()

    val json = "\"2023-08-07T14:27:33.710111-07:00\""
    val instant = gson.fromJson(json, Instant::class.java)
    
    println(instant)
  }
}
