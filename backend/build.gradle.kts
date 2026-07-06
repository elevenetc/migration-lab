plugins {
    kotlin("jvm") version "2.3.21"
    kotlin("plugin.serialization") version "2.3.21"
    id("io.ktor.plugin") version "3.2.3"
}

group = "com.migrationtimeline"
version = "0.1.0"

application {
    mainClass.set("com.migrationtimeline.ApplicationKt")
}

repositories {
    mavenCentral()
}

dependencies {
    // Ktor
    implementation("io.ktor:ktor-server-netty:3.2.3")
    implementation("io.ktor:ktor-server-content-negotiation:3.2.3")
    implementation("io.ktor:ktor-serialization-kotlinx-json:3.2.3")
    implementation("io.ktor:ktor-server-cors:3.2.3")

    // SQL Parsing
    implementation("com.github.jsqlparser:jsqlparser:5.0")

    // Logging
    implementation("ch.qos.logback:logback-classic:1.5.25")

    // Database
    implementation("org.testcontainers:testcontainers-postgresql:2.0.5")
    implementation("org.postgresql:postgresql:42.7.7")

    // Testing
    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.11.4")
}

tasks.test {
    useJUnitPlatform()
}

kotlin {
    jvmToolchain(21)
}
