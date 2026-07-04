# Build stage
FROM eclipse-temurin:21-jdk-alpine AS build
WORKDIR /app

# Copy only build files first (rarely change)
COPY gradle gradle
COPY gradlew settings.gradle.kts gradle.properties ./
COPY backend/build.gradle.kts backend/

# Download dependencies (cached until build.gradle.kts changes)
RUN --mount=type=cache,target=/root/.gradle ./gradlew :backend:dependencies --no-daemon

# Now copy source (changes frequently)
COPY backend/src backend/src

# Build (only recompiles, deps already cached)
RUN --mount=type=cache,target=/root/.gradle ./gradlew :backend:buildFatJar --no-daemon

# Runtime stage
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /app/backend/build/libs/*-all.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
