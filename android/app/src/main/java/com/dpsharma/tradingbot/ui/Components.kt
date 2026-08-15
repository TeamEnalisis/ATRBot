package com.dpsharma.tradingbot.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

enum class Tab(val label: String) {
    DASHBOARD("Dashboard"),
    CONFIG("Configuration"),
}

@Composable
fun TopTabBar(selected: Tab, onSelect: (Tab) -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(24.dp))
            .background(CardDark)
            .padding(4.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Tab.entries.forEach { tab ->
            val isSelected = tab == selected
            Box(
                modifier = Modifier
                    .weight(1f)
                    .clip(RoundedCornerShape(20.dp))
                    .background(if (isSelected) AccentBlue else CardDark)
                    .clickableNoRipple { onSelect(tab) }
                    .padding(vertical = 10.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = tab.label,
                    color = if (isSelected) TextPrimary else TextSecondary,
                    fontSize = 12.sp,
                    fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                    maxLines = 1
                )
            }
        }
    }
}

@Composable
fun StatusRow(running: Boolean, subtitle: String) {
    Row(verticalAlignment = Alignment.CenterVertically, modifier = Modifier.padding(vertical = 12.dp)) {
        Box(
            modifier = Modifier
                .size(10.dp)
                .clip(CircleShape)
                .background(if (running) PositiveGreen else TextSecondary)
        )
        Spacer(Modifier.width(8.dp))
        Text(if (running) "Running" else "Idle", color = TextPrimary, fontWeight = FontWeight.Medium)
        Spacer(Modifier.width(12.dp))
        Text(subtitle, color = TextSecondary, fontFamily = FontFamily.Monospace)
    }
}

@Composable
fun SectionHeader(title: String, subtitle: String? = null) {
    Text(title, color = TextPrimary, fontSize = 22.sp, fontWeight = FontWeight.Bold)
    if (subtitle != null) {
        Spacer(Modifier.height(4.dp))
        Text(subtitle, color = TextSecondary, fontSize = 13.sp)
    }
    Spacer(Modifier.height(12.dp))
}

@Composable
fun StatCard(
    modifier: Modifier = Modifier,
    label: String,
    value: String,
    subtitle: String,
    valueColor: androidx.compose.ui.graphics.Color = TextPrimary,
) {
    Column(
        modifier = modifier
            .clip(RoundedCornerShape(16.dp))
            .background(CardDark)
            .border(1.dp, BorderDark, RoundedCornerShape(16.dp))
            .padding(16.dp)
    ) {
        Text(label.uppercase(), color = TextSecondary, fontSize = 11.sp, letterSpacing = 1.sp)
        Spacer(Modifier.height(10.dp))
        Text(value, color = valueColor, fontSize = 24.sp, fontFamily = FontFamily.Monospace, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(6.dp))
        Text(subtitle, color = TextSecondary, fontSize = 12.sp)
    }
}

@Composable
fun DataTile(modifier: Modifier = Modifier, label: String, value: String, valueColor: androidx.compose.ui.graphics.Color = TextPrimary) {
    Column(
        modifier = modifier
            .clip(RoundedCornerShape(12.dp))
            .background(BgDark)
            .padding(vertical = 14.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(label.uppercase(), color = TextSecondary, fontSize = 10.sp, letterSpacing = 0.5.sp)
        Spacer(Modifier.height(6.dp))
        Text(value, color = valueColor, fontSize = 17.sp, fontFamily = FontFamily.Monospace)
    }
}

/** Simple no-visual-feedback clickable, kept dependency-free (no extra indication libs). */
@Composable
fun Modifier.clickableNoRipple(onClick: () -> Unit): Modifier = this.then(
    Modifier.clickable(indication = null, interactionSource = androidx.compose.runtime.remember { androidx.compose.foundation.interaction.MutableInteractionSource() }) { onClick() }
)
