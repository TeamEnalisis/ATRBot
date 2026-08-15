package com.dpsharma.tradingbot.ui

/**
 * Delta Exchange's raw API responses are decoded by Gson into generic
 * Map/List trees (since we don't hard-model every field). These helpers dig
 * out the handful of numbers the dashboard cares about, failing soft (to
 * null / "—") instead of crashing if a field name differs from expected —
 * double check field names against https://docs.delta.exchange for your
 * account/product type and adjust here if needed.
 */

@Suppress("UNCHECKED_CAST")
fun asMap(v: Any?): Map<String, Any?>? = v as? Map<String, Any?>

@Suppress("UNCHECKED_CAST")
fun asList(v: Any?): List<Any?>? = v as? List<Any?>

fun asDouble(v: Any?): Double? = when (v) {
    is Double -> v
    is Int -> v.toDouble()
    is String -> v.toDoubleOrNull()
    else -> null
}

fun resultList(root: Any?): List<Any?> = asList(asMap(root)?.get("result")) ?: emptyList()

/** Sums `available_balance` (falls back to `balance`) across all wallet entries. */
fun totalAvailableBalance(balancesRoot: Any?): Double? {
    val entries = resultList(balancesRoot)
    if (entries.isEmpty()) return null
    var sum = 0.0
    var found = false
    for (e in entries) {
        val m = asMap(e) ?: continue
        val v = asDouble(m["available_balance"]) ?: asDouble(m["balance"])
        if (v != null) {
            sum += v
            found = true
        }
    }
    return if (found) sum else null
}

data class PositionsSummary(val count: Int, val unrealizedPnl: Double)

fun summarizePositions(positionsRoot: Any?): PositionsSummary {
    val entries = resultList(positionsRoot)
    var count = 0
    var pnl = 0.0
    for (e in entries) {
        val m = asMap(e) ?: continue
        val size = asDouble(m["size"]) ?: 0.0
        if (size == 0.0) continue
        count++
        pnl += asDouble(m["unrealized_pnl"]) ?: 0.0
    }
    return PositionsSummary(count, pnl)
}

data class MarketSummary(
    val markPrice: Double?,
    val spotPrice: Double?,
    val high: Double?,
    val low: Double?,
    val volume: Double?,
    val openInterest: Double?,
    val change24h: Double?,
)

fun summarizeMarket(market: Map<String, Any?>?): MarketSummary {
    if (market == null) return MarketSummary(null, null, null, null, null, null, null)
    return MarketSummary(
        markPrice = asDouble(market["mark_price"]) ?: asDouble(market["close"]),
        spotPrice = asDouble(market["spot_price"]),
        high = asDouble(market["high"]),
        low = asDouble(market["low"]),
        volume = asDouble(market["volume"]),
        openInterest = asDouble(market["oi"]) ?: asDouble(market["open_interest"]),
        change24h = asDouble(market["mark_change_24h"]) ?: asDouble(market["change_24h"]),
    )
}

fun fmtMoney(v: Double?, prefix: String = "$"): String =
    if (v == null) "—" else String.format("%s%.2f", prefix, v)

fun fmtPct(v: Double?): String =
    if (v == null) "—" else String.format("%.2f%%", v)
