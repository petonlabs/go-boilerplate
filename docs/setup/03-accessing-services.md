
---
title: "Accessing services"
order: 3
writebook: true
description: "Practical commands for HTTP, Postgres, Redis and backups."
---

Accessing services
------------------

Practical commands to reach the running services when the compose stack is up.

HTTP (backend)

```bash
# base
http://localhost:8080

# health (pretty JSON)
curl -sS http://localhost:8080/health | python3 -m json.tool

# open OpenAPI UI (macOS)
make open-docs
```

Postgres

```bash
# interactive psql inside the running container
make psql

# single command example
docker compose exec postgres psql -U app -d app -c "SELECT now();"
```

Redis

```bash
make redis-cli
```

Backups

```bash
make list-backups
```

If you prefer to connect a GUI to Postgres/Redis on the host, enable `docker-compose.override.yml` (opt-in) which maps ports 5432 and 6379 to localhost.
