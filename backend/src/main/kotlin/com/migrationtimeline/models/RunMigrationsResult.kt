package com.migrationtimeline.models

import kotlinx.serialization.Serializable

@Serializable
data class RunMigrationsResult(
    val success: Boolean,
    val message: String,
    val migrationsApplied: Int
)
