package com.migrationtimeline

import com.migrationtimeline.routes.migrationRoutes
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.*
import io.ktor.server.application.*
import kotlinx.serialization.json.Json
import io.ktor.server.netty.*
import io.ktor.server.plugins.contentnegotiation.*
import io.ktor.server.plugins.cors.routing.*
import io.ktor.server.response.*
import io.ktor.server.routing.*

fun main(args: Array<String>): Unit = EngineMain.main(args)

fun Application.module() {
    install(ContentNegotiation) {
        json(Json { encodeDefaults = true })
    }

    install(CORS) {
        allowHost("localhost:3000")
        allowHost("frontend:3000")
        allowHeader(HttpHeaders.ContentType)
    }

    routing {
        get("/health") {
            call.respond(HttpStatusCode.OK, mapOf("status" to "ok"))
        }
    }

    migrationRoutes()
}
