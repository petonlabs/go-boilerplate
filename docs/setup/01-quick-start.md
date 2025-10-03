---
title: "Quick start"
order: 1
writebook: true
description: "Get the project running locally in ~5 minutes."
---

Quick start
---------------

Get the project running locally in ~5 minutes.

Prerequisites

1. Docker & Docker Compose
1. Go 1.24+ (for local builds/tests)

Steps

1. Prepare env (do not commit secrets):

```bash
cp apps/backend/.env.example apps/backend/.env
# or run helper to append safe dev defaults
make check-env
```

1. Start services:

```bash
make docker-up
```

1. Confirm health:

```bash
curl -sS http://localhost:8080/health | python3 -m json.tool
```

If you see `"status": "healthy"` the backend is ready.
