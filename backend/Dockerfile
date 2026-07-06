FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

RUN apk add --no-cache docker-cli

COPY backend/build/libs/*-all.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
