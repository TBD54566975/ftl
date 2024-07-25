package xyz.block.ftl.serializer

import com.google.gson.Gson
import com.google.gson.GsonBuilder
import com.google.gson.JsonDeserializer
import com.google.gson.JsonPrimitive
import com.google.gson.JsonSerializer
import java.time.OffsetDateTime
import java.time.format.DateTimeFormatter
import java.util.*

private val FORMATTER: DateTimeFormatter = DateTimeFormatter.ISO_OFFSET_DATE_TIME!!

fun makeGson(): Gson = GsonBuilder()
  .registerTypeAdapter(OffsetDateTime::class.java, JsonSerializer<OffsetDateTime> { src, _, _ ->
    JsonPrimitive(FORMATTER.format(src))
  })
  .registerTypeAdapter(OffsetDateTime::class.java, JsonDeserializer { json, _, _ ->
    OffsetDateTime.parse(json.asString, DateTimeFormatter.ISO_OFFSET_DATE_TIME)
  })
  .registerTypeAdapter(ByteArray::class.java, JsonSerializer<ByteArray> { src, _, _ ->
    JsonPrimitive(Base64.getEncoder().encodeToString(src))
  })
  .registerTypeAdapter(ByteArray::class.java, JsonDeserializer { json, _, _ ->
    Base64.getDecoder().decode(json.asString)
  })
  .create()
