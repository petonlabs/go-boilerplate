---
title: "Quick start"
chapter: 1
description: "Get the project up and running locally in five minutes."
---

Prerequisites: Docker & Docker Compose, Go 1.24+, (optional) bun/npm for frontend.

Steps:

1. Copy example env into `apps/backend/.env` and edit secrets (do not commit secrets):

```bash
cp apps/backend/.env.example apps/backend/.env
# or run the helper that appends safe dev defaults
make check-env
```

2. Start the stack:

```bash
make docker-up
# or: docker compose up -d --build
```

3. Verify the backend is healthy:

```bash
curl -sS http://localhost:8080/health | python3 -m json.tool
```

If you see `{"status":"healthy", ...}` you're good to go.
