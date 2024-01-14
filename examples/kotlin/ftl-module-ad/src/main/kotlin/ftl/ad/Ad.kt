package ftl.ad

import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import xyz.block.ftl.Context
import xyz.block.ftl.Ingress
import xyz.block.ftl.Method
import xyz.block.ftl.Verb
import java.util.*

data class Ad(val redirectUrl: String, val text: String)
data class AdRequest(val contextKeys: List<String>? = null)
data class AdResponse(val ads: List<Ad>)

class AdModule {
  private val database: Map<String, Ad> = loadDatabase()

  @Verb
  @Ingress(Method.GET, "/get")
  fun get(context: Context, req: AdRequest): AdResponse {
    return when {
      req.contextKeys != null -> AdResponse(ads = contextualAds(req.contextKeys))
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
    private val DATABASE = mapOf(
      "hair" to Ad("/product/2ZYFJ3GM2N", "Hairdryer for sale. 50% off."),
      "clothing" to Ad("/product/66VCHSJNUP", "Tank top for sale. 20% off."),
      "accessories" to Ad("/product/1YMWWN1N4O", "Watch for sale. Buy one, get second kit for free"),
      "footwear" to Ad("/product/L9ECAV7KIM", "Loafers for sale. Buy one, get second one for free"),
      "decor" to Ad("/product/0PUK6V6EV0", "Candle holder for sale. 30% off."),
      "kitchen" to Ad("/product/9SIQT8TOJO", "Bamboo glass jar for sale. 10% off.")
    )

    private fun loadDatabase(): Map<String, Ad> {
      return DATABASE
    }

    inline fun <reified T> Gson.fromJson(json: String) = fromJson<T>(json, object : TypeToken<T>() {}.type)
  }
}
