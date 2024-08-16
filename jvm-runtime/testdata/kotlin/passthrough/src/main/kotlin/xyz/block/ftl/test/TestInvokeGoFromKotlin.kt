package xyz.block.ftl.test

import ftl.gomodule.*
import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import java.time.ZonedDateTime

@Export
@Verb
fun emptyVerb(emptyVerbClient: EmptyVerbClient) {
  emptyVerbClient.call()
}

@Export
@Verb
fun sinkVerb(input: String, sinkVerbClient: SinkVerbClient) {
  sinkVerbClient.call(input)
}

@Export
@Verb
fun sourceVerb(sourceVerbClient: SourceVerbClient): String {
  return sourceVerbClient.call()
}

@Export
@Verb
fun errorEmptyVerb(client: ErrorEmptyVerbClient) {
  client.call()
}

@Export
@Verb
fun intVerb(payload: Long, client: IntVerbClient): Long {
  return client.call(payload)
}

@Export
@Verb
fun floatVerb(payload: Double, client: FloatVerbClient): Double {
  return client.call(payload)
}

@Export
@Verb
fun stringVerb(payload: String, client: StringVerbClient): String {
  return client.call(payload)
}

@Export
@Verb
fun bytesVerb(payload: ByteArray, client: BytesVerbClient): ByteArray {
  return client.call(payload)
}

@Export
@Verb
fun boolVerb(payload: Boolean, client: BoolVerbClient): Boolean {
  return client.call(payload)
}

@Export
@Verb
fun stringArrayVerb(payload: List<String>, client: StringArrayVerbClient): List<String> {
  return client.call(payload)
}

@Export
@Verb
fun stringMapVerb(payload: Map<String, String>, client: StringMapVerbClient): Map<String, String> {
  return client.call(payload)
}

@Export
@Verb
fun timeVerb(instant: ZonedDateTime, client: TimeVerbClient): ZonedDateTime {
  return client.call(instant)
}

@Export
@Verb
fun testObjectVerb(payload: TestObject, client: TestObjectVerbClient): TestObject {
  return client.call(payload)
}

@Export
@Verb
fun testObjectOptionalFieldsVerb(
  payload: TestObjectOptionalFields,
  client: TestObjectOptionalFieldsVerbClient
): TestObjectOptionalFields {
  return client.call(payload)
}

// now the same again but with option return / input types
@Export
@Verb
fun optionalIntVerb(payload: Long, client: OptionalIntVerbClient): Long {
  return client.call(payload)
}

@Export
@Verb
fun optionalFloatVerb(payload: Double, client: OptionalFloatVerbClient): Double {
  return client.call(payload)
}

@Export
@Verb
fun optionalStringVerb(payload: String, client: OptionalStringVerbClient): String {
  return client.call(payload)
}

@Export
@Verb
fun optionalBytesVerb(payload: ByteArray?, client: OptionalBytesVerbClient): ByteArray {
  return client.call(payload!!)
}

@Export
@Verb
fun optionalBoolVerb(payload: Boolean, client: OptionalBoolVerbClient): Boolean {
  return client.call(payload)
}

@Export
@Verb
fun optionalStringArrayVerb(payload: List<String>, client: OptionalStringArrayVerbClient): List<String> {
  return client.call(payload)
}

@Export
@Verb
fun optionalStringMapVerb(payload: Map<String, String>, client: OptionalStringMapVerbClient): Map<String, String> {
  return client.call(payload)
}

@Export
@Verb
fun optionalTimeVerb(instant: ZonedDateTime?, client: OptionalTimeVerbClient): ZonedDateTime {
  return client.call(instant!!)
}

@Export
@Verb
fun optionalTestObjectVerb(payload: TestObject?, client: OptionalTestObjectVerbClient): TestObject {
  return client.call(payload!!)
}

@Export
@Verb
fun optionalTestObjectOptionalFieldsVerb(
  payload: TestObjectOptionalFields?,
  client: OptionalTestObjectOptionalFieldsVerbClient
): TestObjectOptionalFields {
  return client.call(payload!!)
}
