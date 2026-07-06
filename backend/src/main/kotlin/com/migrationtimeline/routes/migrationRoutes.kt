package com.migrationtimeline.routes

import com.migrationtimeline.routes.dummy.*
import com.migrationtimeline.runner.runMigrations
import io.ktor.server.application.*
import io.ktor.server.response.*
import io.ktor.server.routing.*

//private val currentMigrations: Map<String, String> = partitionedTableWithAlterMigrations
private val currentMigrations: Map<String, String> = complexEcommerceMigrations

fun Application.migrationRoutes() {
    routing {
        get("/api/migrations") {
            call.respond(migrationsToResponse(currentMigrations))
        }

        post("/api/migrations/run") {
            val result = runMigrations(currentMigrations)
            call.respond(result)
        }
    }
}
