package com.vexil.vexil_app

import android.os.Handler
import android.os.Looper
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import io.flutter.plugin.common.EventChannel
import bridge.Bridge_
import bridge.EventListener
import android.content.Intent
import android.os.Build

class MainActivity : FlutterActivity() {
    private val METHOD_CHANNEL = "com.vexil/vexil"
    private val EVENT_CHANNEL = "com.vexil/vexil_events"

    private var bridge: Bridge_? = null
    private var eventSink: EventChannel.EventSink? = null
    private val mainHandler = Handler(Looper.getMainLooper())

    private var eventListener: EventListener? = null

    private val STORAGE_PERMISSION_CODE = 1002
    private var storagePermissionResult: MethodChannel.Result? = null

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        val filesDir = getFilesDir()
        val tzOffset = java.util.TimeZone.getDefault().rawOffset / 1000
        bridge = Bridge_(filesDir.absolutePath, tzOffset.toLong())

        eventListener = object : EventListener {
            override fun onProgress(
                taskID: String?,
                state: String?,
                percent: Double,
                speedMBps: Double,
                sent: Long,
                total: Long,
                eta: Long
            ) {
                mainHandler.post {
                    val map = mapOf(
                        "type" to "progress",
                        "taskId" to (taskID ?: ""),
                        "state" to (state ?: ""),
                        "percent" to percent,
                        "speedMBps" to speedMBps,
                        "sent" to sent,
                        "total" to total,
                        "eta" to eta
                    )
                    eventSink?.success(map)
                }
            }

            override fun onComplete(taskID: String?) {
                mainHandler.post {
                    val map = mapOf(
                        "type" to "complete",
                        "taskId" to (taskID ?: "")
                    )
                    eventSink?.success(map)
                }
            }

            override fun onError(taskID: String?, err: String?) {
                mainHandler.post {
                    val map = mapOf(
                        "type" to "error",
                        "taskId" to (taskID ?: ""),
                        "error" to (err ?: "")
                    )
                    eventSink?.success(map)
                }
            }
        }
        bridge?.setEventListener(eventListener)

        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, METHOD_CHANNEL)
            .setMethodCallHandler { call, result ->
                when (call.method) {
                    "getLocalIP" -> {
                        try {
                            val wifiManager = applicationContext.getSystemService(WIFI_SERVICE) as? android.net.wifi.WifiManager
                            val ip = wifiManager?.connectionInfo?.ipAddress ?: 0
                            val ipStr = String.format("%d.%d.%d.%d",
                                ip and 0xff,
                                ip shr 8 and 0xff,
                                ip shr 16 and 0xff,
                                ip shr 24 and 0xff
                            )
                            result.success(ipStr)
                        } catch (e: Exception) {
                            result.success("")
                        }
                    }
                    "discoverDevices" -> {
                        val timeout = call.argument<Int>("timeout") ?: 3
                        Thread {
                            try {
                                val json = bridge?.discoverDevices(timeout.toLong()) ?: "[]"
                                mainHandler.post { result.success(json) }
                            } catch (e: Exception) {
                                mainHandler.post { result.error("DISCOVER_ERROR", e.message, null) }
                            }
                        }.start()
                    }
                    "sendFiles" -> {
                        val host = call.argument<String>("host") ?: ""
                        val port = call.argument<Int>("port") ?: 0
                        val pathsJson = call.argument<String>("paths") ?: "[]"
                        val peerName = call.argument<String>("peerName") ?: ""
                        val numConns = call.argument<Int>("numConns") ?: 4
                        val maxChunkMB = call.argument<Int>("maxChunkMB") ?: 16
                        val tlsEnabled = call.argument<Boolean>("tlsEnabled") ?: true
                        Thread {
                            try {
                                val taskId = bridge?.startSend(host, port.toLong(), pathsJson, peerName, numConns.toLong(), maxChunkMB.toLong(), tlsEnabled) ?: ""
                                mainHandler.post { result.success(taskId) }
                            } catch (e: Exception) {
                                mainHandler.post { result.error("SEND_ERROR", e.message, null) }
                            }
                        }.start()
                    }
                    "startReceive" -> {
                        val port = call.argument<Int>("port") ?: 0
                        val saveDir = call.argument<String>("saveDir") ?: ""
                        val deviceName = call.argument<String>("deviceName") ?: ""
                        val numConns = call.argument<Int>("numConns") ?: 4
                        val maxChunkMB = call.argument<Int>("maxChunkMB") ?: 16
                        val tlsEnabled = call.argument<Boolean>("tlsEnabled") ?: true
                        Thread {
                            try {
                                val taskId = bridge?.startReceive(port.toLong(), saveDir, deviceName, numConns.toLong(), maxChunkMB.toLong(), tlsEnabled) ?: ""
                                mainHandler.post { result.success(taskId) }
                            } catch (e: Exception) {
                                mainHandler.post { result.error("RECEIVE_ERROR", e.message, null) }
                            }
                        }.start()
                    }
                    "cancelTransfer" -> {
                        val taskId = call.argument<String>("taskId") ?: ""
                        Thread {
                            try {
                                bridge?.cancelTransfer(taskId)
                                mainHandler.post { result.success(null) }
                            } catch (e: Exception) {
                                mainHandler.post { result.error("CANCEL_ERROR", e.message, null) }
                            }
                        }.start()
                    }
                    "getHistory" -> {
                        val limit = call.argument<Int>("limit") ?: 20
                        Thread {
                            try {
                                val json = bridge?.getHistory(limit.toLong()) ?: "[]"
                                android.util.Log.d("Vexil", "getHistory result: $json")
                                mainHandler.post { result.success(json) }
                            } catch (e: Exception) {
                                mainHandler.post { result.error("HISTORY_ERROR", e.message, null) }
                            }
                        }.start()
                    }
                    "clearHistory" -> {
                        Thread {
                            try {
                                bridge?.clearHistory()
                                mainHandler.post { result.success(null) }
                            } catch (e: Exception) {
                                mainHandler.post { result.error("CLEAR_ERROR", e.message, null) }
                            }
                        }.start()
                    }
                    "deleteHistory" -> {
                        val index = call.argument<Int>("index") ?: 0
                        Thread {
                            try {
                                bridge?.deleteHistory(index.toLong())
                                mainHandler.post { result.success(null) }
                            } catch (e: Exception) {
                                mainHandler.post { result.error("DELETE_ERROR", e.message, null) }
                            }
                        }.start()
                    }

                    "checkWritePermission" -> {
                        val path = call.argument<String>("path") ?: ""
                        val file = java.io.File(path)
                        result.success(file.exists() && file.canWrite())
                    }
                    
                    "requestStoragePermission" -> {
                        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
                            val intent = Intent(android.provider.Settings.ACTION_MANAGE_APP_ALL_FILES_ACCESS_PERMISSION).apply {
                                data = android.net.Uri.parse("package:${applicationContext.packageName}")
                            }
                            startActivity(intent)
                            result.success(null)
                        } else {
                            storagePermissionResult = result
                            requestPermissions(
                                arrayOf(
                                    android.Manifest.permission.READ_EXTERNAL_STORAGE,
                                    android.Manifest.permission.WRITE_EXTERNAL_STORAGE
                                ),
                                STORAGE_PERMISSION_CODE
                            )
                        }
                    }

                    "openFile" -> {
                        val path = call.argument<String>("path") ?: ""
                        try {
                            val file = java.io.File(path)
                            if (!file.exists()) {
                                result.error("FILE_NOT_FOUND", "文件不存在: $path", null)
                            } else {
                                val uri = androidx.core.content.FileProvider.getUriForFile(
                                    this@MainActivity,
                                    "${applicationContext.packageName}.fileprovider",
                                    file
                                )
                                val intent = android.content.Intent(android.content.Intent.ACTION_VIEW).apply {
                                    setDataAndType(uri, "*/*")
                                    addFlags(android.content.Intent.FLAG_GRANT_READ_URI_PERMISSION)
                                    addFlags(android.content.Intent.FLAG_ACTIVITY_NEW_TASK)
                                }
                                startActivity(intent)
                                result.success(null)
                            }
                        } catch (e: Exception) {
                            result.error("OPEN_ERROR", e.message, null)
                        }
                    }
                    "deleteFile" -> {
                        val savePath = call.argument<String>("savePath") ?: ""
                        val fileNames = call.argument<List<String>>("fileNames") ?: emptyList()
                        Thread {
                            try {
                                for (name in fileNames) {
                                    val file = java.io.File(savePath, name)
                                    if (file.exists()) file.delete()
                                }
                                mainHandler.post { result.success(null) }
                            } catch (e: Exception) {
                                mainHandler.post { result.error("DELETE_FILE_ERROR", e.message, null) }
                            }
                        }.start()
                    }
                    "deleteFiles" -> {
                        val pathsStr = call.argument<String>("paths") ?: ""
                        android.util.Log.d("Vexil", "deleteFiles paths: $pathsStr")
                        Thread {
                            try {
                                val paths = pathsStr.split(", ")
                                for (path in paths) {
                                    if (path.isNotEmpty()) {
                                        val file = java.io.File(path)
                                        val deleted = file.delete()
                                        android.util.Log.d("Vexil", "delete $path -> $deleted")
                                    }
                                }
                                mainHandler.post { result.success(null) }
                            } catch (e: Exception) {
                                android.util.Log.e("Vexil", "deleteFiles error: ${e.message}")
                                mainHandler.post { result.error("DELETE_FILE_ERROR", e.message, null) }
                            }
                        }.start()
                    }
                    "startForegroundService" -> {
                        val intent = Intent(this@MainActivity, ForegroundService::class.java)
                        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                            startForegroundService(intent)
                        } else {
                            startService(intent)
                        }
                        result.success(null)
                    }

                    "stopForegroundService" -> {
                        val intent = Intent(this@MainActivity, ForegroundService::class.java)
                        stopService(intent)
                        result.success(null)
                    }

                    "updateNotification" -> {
                        val title = call.argument<String>("title") ?: ""
                        val content = call.argument<String>("content") ?: ""
                        val stopAfter = call.argument<Boolean>("stopAfter") ?: false
                        ForegroundService.instance?.updateNotification(title, content)
                        if (stopAfter) {
                            val intent = Intent(this@MainActivity, ForegroundService::class.java)
                            stopService(intent)
                        }
                        result.success(null)
                    }
                    else -> result.notImplemented()
                }
            }

        EventChannel(flutterEngine.dartExecutor.binaryMessenger, EVENT_CHANNEL)
            .setStreamHandler(object : EventChannel.StreamHandler {
                override fun onListen(arguments: Any?, events: EventChannel.EventSink?) {
                    eventSink = events
                }

                override fun onCancel(arguments: Any?) {
                    eventSink = null
                }
            })
    }

    override fun onRequestPermissionsResult(requestCode: Int, permissions: Array<out String>, grantResults: IntArray) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == STORAGE_PERMISSION_CODE) {
            val granted = grantResults.all { it == android.content.pm.PackageManager.PERMISSION_GRANTED }
            storagePermissionResult?.success(granted)
            storagePermissionResult = null
        }
    }
}