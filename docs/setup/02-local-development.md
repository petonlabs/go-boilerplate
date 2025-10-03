
---
title: "Local development"
order: 2
writebook: true
description: "How to build, run and iterate on the backend locally."
---

Local development
-----------------

Useful commands to iterate on the backend and services.

Build and run

```bash
# build binary
make build

# run backend locally (reads apps/backend/.env)
make backend-run
```

Docker-based development

```bash
# bring up DB, redis, backup runner, and backend container
make docker-up

# follow backend logs
make logs
```

Notes

- The project avoids publishing DB/Redis ports to the host by default. Use the optional `docker-compose.override.yml` if you need host ports for GUI tools.
