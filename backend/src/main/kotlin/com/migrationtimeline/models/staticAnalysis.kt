package com.migrationtimeline.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class AnalysisResult(
    val warnings: List<Warning>
)

@Serializable
sealed class Warning {
    abstract val operationId: OperationId
}

@Serializable
@SerialName("ACCESS_EXCLUSIVE_LOCK")
data class AccessExclusiveLock(
    override val operationId: OperationId,
    val tableName: String,
    val message: String
) : Warning()

@Serializable
data class OperationId(
    val migrationId: String,
    val tableName: String
)
