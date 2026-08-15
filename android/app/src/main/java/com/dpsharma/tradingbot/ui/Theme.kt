package com.dpsharma.tradingbot.ui

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

val BgDark = Color(0xFF0A0E14)
val CardDark = Color(0xFF12161F)
val BorderDark = Color(0xFF232838)
val AccentTeal = Color(0xFF2DD4BF)
val AccentBlue = Color(0xFF3B82F6)
val TextPrimary = Color(0xFFF1F5F9)
val TextSecondary = Color(0xFF8B93A7)
val PositiveGreen = Color(0xFF34D399)
val NegativeRed = Color(0xFFF87171)

private val DpSharmaColors = darkColorScheme(
    background = BgDark,
    surface = CardDark,
    primary = AccentBlue,
    secondary = AccentTeal,
    onBackground = TextPrimary,
    onSurface = TextPrimary,
)

@Composable
fun TradingBotTheme(content: @Composable () -> Unit) {
    MaterialTheme(colorScheme = DpSharmaColors, content = content)
}
