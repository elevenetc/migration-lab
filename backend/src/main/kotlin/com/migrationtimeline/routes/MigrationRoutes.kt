package com.migrationtimeline.routes

import com.migrationtimeline.routes.dummy.complexEcommerce
import com.migrationtimeline.routes.dummy.simpleUserAndOrg
import io.ktor.server.application.*
import io.ktor.server.response.*
import io.ktor.server.routing.*


fun Application.migrationRoutes() {
    routing {
        get("/api/migrations") {
            //call.respond(simpleUserAndOrg())
            call.respond(complexEcommerce())
        }
    }
}
