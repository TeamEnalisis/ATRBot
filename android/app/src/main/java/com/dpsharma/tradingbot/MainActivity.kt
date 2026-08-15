package com.dpsharma.tradingbot

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dpsharma.tradingbot.ui.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            TradingBotTheme {
                Surface()
            }
        }
    }
}

@Composable
private fun Surface() {
    val context = androidx.compose.ui.platform.LocalContext.current
    val settings = remember { Settings(context) }
    var api by remember { mutableStateOf(ApiClientFactory.build(settings)) }

    var tab by remember { mutableStateOf(Tab.DASHBOARD) }
    var dashboard by remember { mutableStateOf<DashboardResponse?>(null) }
    var config by remember { mutableStateOf<BotConfig?>(null) }
    var running by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }

    val scope = rememberCoroutineScope()

    suspend fun refresh() {
        try {
            dashboard = api.dashboard()
            running = dashboard?.botRunning ?: running
            error = null
        } catch (e: Exception) {
            error = "Backend unreachable: ${e.message}"
        }
        try {
            config = api.getConfig()
        } catch (_: Exception) { /* keep last known config */ }
    }

    LaunchedEffect(api) {
        while (true) {
            refresh()
            delay(5000)
        }
    }

    Column(
        Modifier
            .fillMaxSize()
            .background(BgDark)
            .padding(16.dp)
    ) {
        Text("D.P.Sharma", color = AccentTeal, fontSize = 30.sp, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(16.dp))
        TopTabBar(selected = tab, onSelect = { tab = it })
        Spacer(Modifier.height(16.dp))

        Column(Modifier.weight(1f)) {
            when (tab) {
                Tab.DASHBOARD -> DashboardScreen(
                    dashboard = dashboard,
                    running = running,
                    config = config,
                    errorMessage = error,
                    onStart = {
                        scope.launch {
                            try { api.start(); refresh() } catch (e: Exception) { error = e.message }
                        }
                    },
                    onStop = {
                        scope.launch {
                            try { api.stop(); refresh() } catch (e: Exception) { error = e.message }
                        }
                    }
                )
                Tab.CONFIG -> ConfigScreen(
                    settings = settings,
                    config = config,
                    onSaveConnection = { url, token ->
                        settings.backendUrl = url
                        settings.appToken = token
                        api = ApiClientFactory.build(settings)
                    },
                    onSaveConfig = { newConfig ->
                        scope.launch {
                            try {
                                api.setConfig(newConfig)
                                config = api.getConfig()
                                error = null
                            } catch (e: Exception) {
                                error = "Save failed: ${e.message}"
                            }
                        }
                    }
                )
            }
        }
    }
}
