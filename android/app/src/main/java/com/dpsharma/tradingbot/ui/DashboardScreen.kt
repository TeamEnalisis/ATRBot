package com.dpsharma.tradingbot.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dpsharma.tradingbot.BotConfig
import com.dpsharma.tradingbot.DashboardResponse

@Composable
fun DashboardScreen(
    dashboard: DashboardResponse?,
    running: Boolean,
    config: BotConfig?,
    errorMessage: String?,
    onStart: () -> Unit,
    onStop: () -> Unit,
) {
    val positions = summarizePositions(dashboard?.positions)
    val balance = totalAvailableBalance(dashboard?.balances)
    val market = summarizeMarket(dashboard?.market)
    val entryPrice = dashboard?.botEntryPrice ?: 0.0
    val botPosition = dashboard?.botPosition ?: "flat"

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(bottom = 24.dp)
    ) {
        StatusRow(running = running, subtitle = if (running) "bot polling every ${config?.pollIntervalSeconds ?: 30}s" else "press Start to go live")

        if (errorMessage != null) {
            Text(errorMessage, color = NegativeRed, modifier = Modifier.padding(bottom = 12.dp))
        }

        Button(
            onClick = { if (running) onStop() else onStart() },
            modifier = Modifier.fillMaxWidth().height(52.dp),
            shape = RoundedCornerShape(14.dp),
            colors = ButtonDefaults.buttonColors(
                containerColor = if (running) NegativeRed else AccentTeal,
                contentColor = BgDark
            )
        ) {
            Text(if (running) "STOP BOT" else "START BOT", fontWeight = FontWeight.Bold)
        }

        Spacer(Modifier.height(20.dp))
        SectionHeader("Performance Overview", "Live position and account snapshot")

        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            StatCard(
                modifier = Modifier.weight(1f),
                label = "Bot Position",
                value = botPosition.uppercase(),
                subtitle = if (botPosition == "flat") "no open trade" else "entry ${fmtMoney(entryPrice)}",
                valueColor = when (botPosition) {
                    "long" -> PositiveGreen
                    "short" -> NegativeRed
                    else -> TextPrimary
                }
            )
            StatCard(
                modifier = Modifier.weight(1f),
                label = "Unrealized PnL",
                value = fmtMoney(positions.unrealizedPnl),
                subtitle = "Open positions",
                valueColor = if (positions.unrealizedPnl >= 0) PositiveGreen else NegativeRed
            )
        }
        Spacer(Modifier.height(12.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            StatCard(
                modifier = Modifier.weight(1f),
                label = "Available Balance",
                value = fmtMoney(balance),
                subtitle = "Wallet"
            )
            StatCard(
                modifier = Modifier.weight(1f),
                label = "Open Positions",
                value = positions.count.toString(),
                subtitle = "Active contracts"
            )
        }

        Spacer(Modifier.height(20.dp))

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(16.dp))
                .background(CardDark)
                .padding(16.dp)
        ) {
            Text(
                "MARKET DATA · ${config?.productSymbol ?: "—"}",
                color = TextSecondary,
                fontWeight = FontWeight.SemiBold,
                letterSpacing = 1.sp
            )
            Spacer(Modifier.height(12.dp))
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                DataTile(modifier = Modifier.weight(1f), label = "Mark Price", value = fmtMoney(market.markPrice))
                DataTile(modifier = Modifier.weight(1f), label = "24h Change", value = fmtPct(market.change24h), valueColor = if ((market.change24h ?: 0.0) >= 0) PositiveGreen else NegativeRed)
            }
            Spacer(Modifier.height(8.dp))
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                DataTile(modifier = Modifier.weight(1f), label = "24h Volume", value = market.volume?.let { String.format("%.0f", it) } ?: "—")
                DataTile(modifier = Modifier.weight(1f), label = "Open Interest", value = market.openInterest?.let { String.format("%.0f", it) } ?: "—")
            }
            Spacer(Modifier.height(8.dp))
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                DataTile(modifier = Modifier.weight(1f), label = "24h High/Low", value = "${market.high ?: "—"} / ${market.low ?: "—"}")
                DataTile(modifier = Modifier.weight(1f), label = "Spot Price", value = fmtMoney(market.spotPrice))
            }
        }
    }
}
