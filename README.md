<img src="https://raw.githubusercontent.com/TrevorPhippard/ouroboros/main/app/web/public/mini.svg" alt="Sample Image" width="150" />

# Ouroboros

> A social network without algorithms — just people.

![Status](https://img.shields.io/badge/status-in%20development-yellow)
![Backend](https://img.shields.io/badge/backend-Go-blue)
![Frontend](https://img.shields.io/badge/frontend-Next.js-black)
![API](https://img.shields.io/badge/API-GraphQL%20%2B%20gRPC-purple)
![License](https://img.shields.io/badge/license-MIT-green)

---

## Table of Contents

- [Overview](#overview)
- [Why This Exists](#why-this-exists)
- [What Makes It Different](#what-makes-it-different)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Core Services](#core-services)
- [Development](#development)
- [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)
- [Changelog](#changelog)
- [Future Work](#future-work)

---

## Overview

Ouroboros is a modern social platform designed to remove algorithmic manipulation and bring back intentional, human-centered interaction.

---

## Why This Exists

Most platforms optimize for engagement—not well-being.

This leads to:

- Addictive scrolling loops
- Echo chambers
- Amplified misinformation
- Content shaped by algorithms instead of people

---

## What Makes It Different

- Chronological feed only (no ranking algorithms)
- No engagement optimization loops
- Connection-first design
- Minimalist feature set

---

## Features

- Chronological feed
- User profiles
- Posts and likes
- Notifications
- Connection system (follow/friend)
- Authentication

---

## Prerequisites

Docker & Docker Compose: Latest stable
GraphQL endpoint access: Ensure the GraphQL gateway is accessible
Environment variables configured: Set all required environment variables

---

## Quick Start

```bash
docker-compose up --build
```

### Local Services

| Service            | URL                                     |
| ------------------ | --------------------------------------- |
| Frontend           | <http://localhost:3000>                 |
| GraphQL Playground | <http://localhost:4000/graphql>         |
| pgAdmin            | <http://localhost:8083/browser>         |
| Consul             | <http://localhost:8500/ui/dc1/services> |
| Prometheus         | <http://localhost:9090/query>           |
| Kafdrop            | <http://localhost:9000>                 |

---

## Tech Stack

### Frontend

- Next.js
- TanStack Query
- Zustand
- Tailwind CSS
- shadcn/ui

### Backend

- Go (microservices)
- GraphQL Gateway
- gRPC

### Infrastructure

- PostgreSQL
- Neo4j
- Kafka
- RabbitMQ
- Consul
- Prometheus
- Docker

---

## Architecture

Ouroboros uses a microservices architecture with GraphQL as the entry point.

### High-Level Flow

```text
Client (Next.js)
        ↓
GraphQL Gateway
        ↓
-----------------------------
|   Go Microservices        |
|---------------------------|
| Auth Service              |
| Feed Service              |
| Post Service              |
| Profile Service           |
| Connection Service        |
| Notification Service      |
-----------------------------
        ↓
Data Layer
(Postgres + Neo4j)
        ↓
Event Layer
(Kafka / RabbitMQ)
```

---

## Why GraphQL + gRPC?

- **GraphQL** → flexible client queries
- **gRPC** → fast internal communication

Benefits:

- Clean separation of concerns
- Strong typing across services
- High performance

[API Reference](https://github.com/TrevorPhippard/ouroboros/blob/main/documentation/GraphQL_API.md)

---

## Project Structure

```text
/apps
  /frontend

/services
  /auth
  /feed
  /post
  /profile
  /connection
  /notification

/packages
  /graphql-gateway
  /proto
```

---

## Core Services

| Service      | Responsibility                |
| ------------ | ----------------------------- |
| Auth         | Authentication & sessions     |
| Feed         | Chronological feed generation |
| Post         | Post creation & interactions  |
| Profile      | User data management          |
| Connection   | Social graph relationships    |
| Notification | User notifications            |

---

## Observability

Ouroboros includes basic system observability:

- Prometheus for metrics
- Structured logging across Go services
- Health check endpoints on all services
- Consul for service discovery

---

## Development

### Frontend

```bash
npm install
npm run dev
```

### Go Service

```bash
go test ./...
go run ./cmd/<service-name>
```

---

## Environment Variables

Create:

- `.env.local` (frontend)
- Service-specific `.env` files

Typical variables:

- GraphQL endpoint
- gRPC service addresses
- Database URLs
- Auth secrets
- Message broker configs

---

## Troubleshooting

- Regenerate code after schema changes
- Ensure GraphQL and gRPC contracts match
- Verify services are registered in Consul

---

## Changelog

```text
## [1.0.0] - 2026-04-01
### Added
- Initial microservices architecture
- GraphQL gateway
- Feed + Auth services
```

---

## License

MIT

### citations

accessed 17 April 2026.

'Social Media Addiction and Poor Mental Health: Examining the Mediating Roles of Internet Addiction and Phubbing', PubMed (2023) <https://pubmed.ncbi.nlm.nih.gov/36972903/>
accessed 17 April 2026.

'The echo chamber effect on social media', Proceedings of the National Academy of Sciences (2021) <https://pmc.ncbi.nlm.nih.gov/articles/PMC7936330/>
accessed 17 April 2026.

'Social Media Addiction and Poor Mental Health: Examining the Mediating Roles of Internet Addiction and Phubbing', PubMed (2023) <https://pubmed.ncbi.nlm.nih.gov/36972903/>
accessed 17 April 2026.

'The echo chamber effect on social media', Proceedings of the National Academy of Sciences (2021) <https://pmc.ncbi.nlm.nih.gov/articles/PMC7936330/>
accessed 17 April 2026.
