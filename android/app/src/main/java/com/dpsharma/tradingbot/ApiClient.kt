package com.dpsharma.tradingbot

import android.content.Context
import android.content.SharedPreferences
import com.google.gson.annotations.SerializedName
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST

/**
 * Everything the Android app needs to know about the trading backend lives
 * here (base URL + shared secret). Both are set on-device from the
 * Configuration screen -- the Delta Exchange API key/secret NEVER touch the
 * phone, they're stored only on the backend/VPS.
 */
class Settings(context: Context) {
    private val prefs: SharedPreferences =
        context.getSharedPreferences("dpsharma_tradingbot", Context.MODE_PRIVATE)

    var backendUrl: String
        get() = prefs.getString("backend_url", "http://localhost:8080") ?: "http://localhost:8080"
        set(value) = prefs.edit().putString("backend_url", value).apply()

    var appToken: String
        get() = prefs.getString("app_token", "") ?: ""
        set(value) = prefs.edit().putString("app_token", value).apply()
}

// ---------- Data models (mirror the backend's JSON) ----------

data class StatusResponse(
    val running: Boolean,
    val position: String,
    @SerializedName("entry_price") val entryPrice: Double
)

data class DashboardResponse(
    val balances: Any?,
    val positions: Any?,
    val market: Map<String, Any?>?,
    @SerializedName("bot_position") val botPosition: String?,
    @SerializedName("bot_entry_price") val botEntryPrice: Double?,
    @SerializedName("bot_running") val botRunning: Boolean?
)

data class LogsResponse(val logs: List<String>)

data class BotConfig(
    @SerializedName("delta_api_key") var deltaApiKey: String = "",
    @SerializedName("delta_api_secret") var deltaApiSecret: String = "",
    @SerializedName("base_url") var baseUrl: String = "https://api.india.delta.exchange",
    @SerializedName("product_id") var productId: Int = 0,
    @SerializedName("product_symbol") var productSymbol: String = "BTCUSD",
    @SerializedName("resolution") var resolution: String = "5m",
    @SerializedName("key_value") var keyValue: Double = 1.0,
    @SerializedName("atr_period") var atrPeriod: Int = 10,
    @SerializedName("use_heikin_ashi") var useHeikinAshi: Boolean = false,
    @SerializedName("lot_size") var lotSize: Int = 1,
    @SerializedName("num_lots") var numLots: Int = 1,
    @SerializedName("fee_amount") var feeAmount: Double = 0.0,
    @SerializedName("funding_amount") var fundingAmount: Double = 0.0,
    @SerializedName("brokerage_amount") var brokerageAmount: Double = 0.0,
    @SerializedName("beta") var beta: Double = 2.0,
    @SerializedName("support_level") var supportLevel: Double = 0.0,
    @SerializedName("resistance_level") var resistanceLevel: Double = 0.0,
    @SerializedName("sr_lookback") var srLookback: Int = 50,
    @SerializedName("bot_running") var botRunning: Boolean = false,
    @SerializedName("poll_interval_seconds") var pollIntervalSeconds: Int = 30
)

data class SimpleStatus(val status: String?)

interface BackendApi {
    @GET("/api/status")
    suspend fun status(): StatusResponse

    @GET("/api/dashboard")
    suspend fun dashboard(): DashboardResponse

    @GET("/api/config")
    suspend fun getConfig(): BotConfig

    @POST("/api/config")
    suspend fun setConfig(@Body cfg: BotConfig): SimpleStatus

    @POST("/api/bot/start")
    suspend fun start(): SimpleStatus

    @POST("/api/bot/stop")
    suspend fun stop(): SimpleStatus

    @GET("/api/logs")
    suspend fun logs(): LogsResponse
}

object ApiClientFactory {
    /** Builds a fresh Retrofit client each time so a changed backend URL/token in
     * Settings takes effect immediately without restarting the app. */
    fun build(settings: Settings): BackendApi {
        val tokenInterceptor = Interceptor { chain ->
            val req = chain.request().newBuilder()
                .apply {
                    if (settings.appToken.isNotBlank()) {
                        addHeader("X-App-Token", settings.appToken)
                    }
                }
                .build()
            chain.proceed(req)
        }
        val logging = HttpLoggingInterceptor().apply { level = HttpLoggingInterceptor.Level.BASIC }
        val client = OkHttpClient.Builder()
            .addInterceptor(tokenInterceptor)
            .addInterceptor(logging)
            .build()

        var base = settings.backendUrl
        if (!base.endsWith("/")) base += "/"

        return Retrofit.Builder()
            .baseUrl(base)
            .client(client)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(BackendApi::class.java)
    }
}
