package ftl.ad

import com.google.gson.Gson
import com.google.gson.annotations.SerializedName
import com.google.gson.reflect.TypeToken
import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb
import xyz.block.ftl.serializer.makeGson
import java.util.*

data class Ad(@SerializedName("RedirectURL") val redirectUrl: String, @SerializedName("Text") val text: String)
data class AdRequest(val contextKeys: List<String>)
data class AdResponse(val ads: List<Ad>)

class AdModule {
  private val database: Map<String, Ad> = loadDatabase()

  @Throws(Exception::class)
  @Verb
  @Ingress(Type.FTL, Method.GET, "/get")
  fun get(context: Context, req: AdRequest): AdResponse {
    return when {
      req.contextKeys.isNotEmpty() -> AdResponse(ads = contextualAds(req.contextKeys))
      else -> AdResponse(ads = randomAds())
    }
  }

  private fun contextualAds(contextKeys: List<String>): List<Ad> {
    return contextKeys.map { database[it] ?: throw Exception("no ad registered for this context key") }
  }

  private fun randomAds(): List<Ad> {
    val ads = mutableListOf<Ad>()
    val random = Random()
    repeat(MAX_ADS_TO_SERVE) {
      ads.add(database.entries.elementAt(random.nextInt(database.size)).value)
    }
    return ads
  }

  companion object {
    private const val MAX_ADS_TO_SERVE = 2
    private const val DATABASE_JSON = "{\n" +
      "  \"hair\": {\n" +
      "    \"RedirectURL\": \"/product/2ZYFJ3GM2N\",\n" +
      "    \"Text\": \"Hairdryer for sale. 50% off.\"\n" +
      "  },\n" +
      "  \"clothing\": {\n" +
      "    \"RedirectURL\": \"/product/66VCHSJNUP\",\n" +
      "    \"Text\": \"Tank top for sale. 20% off.\"\n" +
      "  },\n" +
      "  \"accessories\": {\n" +
      "    \"RedirectURL\": \"/product/1YMWWN1N4O\",\n" +
      "    \"Text\": \"Watch for sale. Buy one, get second kit for free\"\n" +
      "  },\n" +
      "  \"footwear\": {\n" +
      "    \"RedirectURL\": \"/product/L9ECAV7KIM\",\n" +
      "    \"Text\": \"Loafers for sale. Buy one, get second one for free\"\n" +
      "  },\n" +
      "  \"decor\": {\n" +
      "    \"RedirectURL\": \"/product/0PUK6V6EV0\",\n" +
      "    \"Text\": \"Candle holder for sale. 30% off.\"\n" +
      "  },\n" +
      "  \"kitchen\": {\n" +
      "    \"RedirectURL\": \"/product/9SIQT8TOJO\",\n" +
      "    \"Text\": \"Bamboo glass jar for sale. 10% off.\"\n" +
      "  }\n" +
      "}"

    private fun loadDatabase(): Map<String, Ad> {
      return makeGson().fromJson<Map<String, Ad>>(DATABASE_JSON)
    }

    inline fun <reified T> Gson.fromJson(json: String) = fromJson<T>(json, object : TypeToken<T>() {}.type)
  }
}
