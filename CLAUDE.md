# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Migration Timeline is a tool for analyzing and visualizing database migrations. It parses Flyway PostgreSQL migrations,
builds AST representations, and renders interactive timelines.

## Architecture

Two-module monorepo:

- **backend/**: Kotlin + Ktor + Gradle service for migration parsing and AST generation
- **frontend/**: React + TypeScript + Zustand web client for visualization

## Tech Stack

### Backend

- Kotlin with Ktor framework
- Gradle build system
- JSQLParser and pg_query for SQL parsing
- Flyway for migration analysis

### Frontend

- React with TypeScript
- Zustand for state management
- @xyflow/react for graph visualization

#### Structure

- `src/api/` - API types and fetch functions (mirrors backend models)
- `src/store/` - Zustand store with async actions
- `src/components/` - React components
- `src/contracts/` - API contract validation

#### Data Flow

1. App mounts → calls `loadMigrations()`
2. Store fetches `/api/migrations`
3. Timeline reads store, groups by table, renders graph

#### Visualization

- Tables as rows, operations as nodes positioned by timestamp
- X-axis: migrations with same timestamp align vertically in same column
- Edges connect operations on same table chronologically
- CREATE_TABLE (blue) vs ALTER_TABLE (darker blue)

## Verification

After making changes, run tests:
```
just test
```

## API Contract Testing

Backend and frontend types are kept in sync via compile-time contract validation:

1. `ApiContractFixtureGenerator.kt` generates JSON fixtures from backend models to `api-contracts/fixtures/`
2. `validate-api-contracts.ts` imports fixtures and validates against TypeScript types
3. Exhaustive switch ensures all Operation variants are handled

When adding new Operation types:
1. Add to `MigrationAst.kt` with `@SerialName`
2. Add to `migrationApi.ts` and update `Operation` union
3. Add case to `validate-api-contracts.ts` switch
4. Update `ApiContractFixtureGenerator.kt` to include example

`just test` fails if types drift.

## Constraints

- PostgreSQL and Flyway migrations only
- Frontend simplicity prioritized over polish for MVP
- See [supported-sql.md](supported-sql.md) for supported SQL statements
