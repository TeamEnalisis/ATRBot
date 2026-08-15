package com.dpsharma.tradingbot.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.dpsharma.tradingbot.BotConfig
import com.dpsharma.tradingbot.Settings

@Composable
fun ConfigScreen(
    settings: Settings,
    config: BotConfig?,
    onSaveConnection: (backendUrl: String, appToken: String) -> Unit,
    onSaveConfig: (BotConfig) -> Unit,
) {
    var backendUrl by remember { mutableStateOf(settings.backendUrl) }
    var appToken by remember { mutableStateOf(settings.appToken) }

    var form by remember(config) { mutableStateOf(config ?: BotConfig()) }

    Column(
        Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(bottom = 40.dp)
    ) {
        SectionHeader("Configuration")

        ConfigGroup("Backend connection") {
            ConfigField("Backend URL", backendUrl, { backendUrl = it }) 
            ConfigField("App Token (X-App-Token)", appToken, { appToken = it })
            Button(onClick = { onSaveConnection(backendUrl, appToken) }, modifier = Modifier.fillMaxWidth()) {
                Text("Save connection")
            }
        }

        ConfigGroup("Delta Exchange API") {
            ConfigField("API Key", form.deltaApiKey, { form = form.copy(deltaApiKey = it) })
            ConfigField("API Secret (leave blank to keep saved value)", form.deltaApiSecret, { form = form.copy(deltaApiSecret = it) })
            ConfigField("Base URL", form.baseUrl, { form = form.copy(baseUrl = it) })
        }

        ConfigGroup("Instrument") {
            ConfigField("Product ID", form.productId.toString(), { form = form.copy(productId = it.toIntOrNull() ?: form.productId) }, KeyboardType.Number)
            ConfigField("Product Symbol", form.productSymbol, { form = form.copy(productSymbol = it) })
            ConfigField("Candle resolution (1m/5m/15m/1h)", form.resolution, { form = form.copy(resolution = it) })
        }

        ConfigGroup("Strategy") {
            ConfigField("Key Value / ATR multiplier (a)", form.keyValue.toString(), { form = form.copy(keyValue = it.toDoubleOrNull() ?: form.keyValue) }, KeyboardType.Decimal)
            ConfigField("ATR Period (c)", form.atrPeriod.toString(), { form = form.copy(atrPeriod = it.toIntOrNull() ?: form.atrPeriod) }, KeyboardType.Number)
            ToggleRow("Use Heikin Ashi source", form.useHeikinAshi) { form = form.copy(useHeikinAshi = it) }
            ConfigField("Poll interval (seconds)", form.pollIntervalSeconds.toString(), { form = form.copy(pollIntervalSeconds = it.toIntOrNull() ?: form.pollIntervalSeconds) }, KeyboardType.Number)
        }

        ConfigGroup("Position sizing") {
            ConfigField("Lot size (contracts/lot)", form.lotSize.toString(), { form = form.copy(lotSize = it.toIntOrNull() ?: form.lotSize) }, KeyboardType.Number)
            ConfigField("Number of lots", form.numLots.toString(), { form = form.copy(numLots = it.toIntOrNull() ?: form.numLots) }, KeyboardType.Number)
        }

        ConfigGroup("Expenses & profit target") {
            ConfigField("Fee amount", form.feeAmount.toString(), { form = form.copy(feeAmount = it.toDoubleOrNull() ?: form.feeAmount) }, KeyboardType.Decimal)
            ConfigField("Funding amount", form.fundingAmount.toString(), { form = form.copy(fundingAmount = it.toDoubleOrNull() ?: form.fundingAmount) }, KeyboardType.Decimal)
            ConfigField("Brokerage amount", form.brokerageAmount.toString(), { form = form.copy(brokerageAmount = it.toDoubleOrNull() ?: form.brokerageAmount) }, KeyboardType.Decimal)
            ConfigField("Beta (profit = expenses × beta on top of entry)", form.beta.toString(), { form = form.copy(beta = it.toDoubleOrNull() ?: form.beta) }, KeyboardType.Decimal)
        }

        ConfigGroup("Stop loss (Support/Resistance)") {
            ConfigField("Support level (0 = auto-detect)", form.supportLevel.toString(), { form = form.copy(supportLevel = it.toDoubleOrNull() ?: form.supportLevel) }, KeyboardType.Decimal)
            ConfigField("Resistance level (0 = auto-detect)", form.resistanceLevel.toString(), { form = form.copy(resistanceLevel = it.toDoubleOrNull() ?: form.resistanceLevel) }, KeyboardType.Decimal)
            ConfigField("Auto S/R lookback (candles)", form.srLookback.toString(), { form = form.copy(srLookback = it.toIntOrNull() ?: form.srLookback) }, KeyboardType.Number)
        }

        Button(
            onClick = { onSaveConfig(form) },
            modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
            colors = ButtonDefaults.buttonColors(containerColor = AccentTeal, contentColor = BgDark)
        ) {
            Text("Save configuration to backend", fontWeight = FontWeight.Bold)
        }

        Spacer(Modifier.height(8.dp))
        Text(
            "Made & Manage with ❤️ by D.P Sharma",
            color = TextSecondary,
            modifier = Modifier.padding(top = 4.dp)
        )
    }
}

@Composable
private fun ConfigGroup(title: String, content: @Composable ColumnScope.() -> Unit) {
    Text(title, color = TextPrimary, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(top = 18.dp, bottom = 8.dp))
    Column(verticalArrangement = Arrangement.spacedBy(10.dp), content = content)
}

@Composable
private fun ConfigField(
    label: String,
    value: String,
    onChange: (String) -> Unit,
    keyboardType: KeyboardType = KeyboardType.Text,
) {
    OutlinedTextField(
        value = value,
        onValueChange = onChange,
        label = { Text(label) },
        singleLine = true,
        keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = OutlinedTextFieldDefaults.colors(
            focusedContainerColor = CardDark,
            unfocusedContainerColor = CardDark,
            focusedTextColor = TextPrimary,
            unfocusedTextColor = TextPrimary,
            focusedLabelColor = AccentTeal,
            unfocusedLabelColor = TextSecondary,
        )
    )
}

@Composable
private fun ToggleRow(label: String, value: Boolean, onChange: (Boolean) -> Unit) {
    Row(
        Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(label, color = TextPrimary)
        Switch(checked = value, onCheckedChange = onChange)
    }
}
