# Production Deployment Guide

Guide for deploying go-boilerplate to production environments.

---

## Overview

This guide covers:
- Deployment platforms (Coolify, Docker, Kubernetes)
- Production configuration
- Security hardening
- Monitoring and alerting
- Backup and disaster recovery

---

## Deployment Options

### Option 1: Coolify (Recommended)

Coolify provides a simple, self-hosted PaaS experience.

#### Setup

1. **Create Application in Coolify**
   - Go to your Coolify dashboard
   - Click "New Application"
   - Select "Docker" deployment type

2. **Configure Build Settings**
   ```
   Dockerfile: apps/backend/Dockerfile
   Build Context: .
   Port: 8080
   ```

3. **Set Environment Variables**
   
   Add via Coolify Secrets UI:
   ```bash
   PRIMARY_ENV=production
   SERVER_PORT=8080
   SERVER_CORS_ALLOWED_ORIGINS=https://yourdomain.com
   
   DATABASE_HOST=your-postgres-host
   DATABASE_PORT=5432
   DATABASE_USER=your-user
   DATABASE_PASSWORD=<secret>
   DATABASE_NAME=your-db
   DATABASE_SSL_MODE=require
   
   REDIS_ADDRESS=your-redis:6379
   REDIS_PASSWORD=<secret>
   
   AUTH_SECRET_KEY=<generate-with-openssl-rand-base64-32>
   
   INTEGRATION_RESEND_API_KEY=<secret>
   
   OBSERVABILITY_NEWRELIC_LICENSE_KEY=<secret>
   OBSERVABILITY_NEWRELIC_APP_NAME=your-app-production
   
   S3_ENDPOINT=https://account-id.r2.cloudflarestorage.com
   S3_BUCKET=your-bucket
   S3_ACCESS_KEY_ID=<secret>
   S3_SECRET_ACCESS_KEY=<secret>
   ```

4. **Optional: Configure Traefik**
   
   Uncomment Traefik labels in `docker-compose.yml`:
   ```yaml
   labels:
     - "traefik.enable=true"
     - "traefik.http.routers.backend.rule=Host(`api.yourdomain.com`)"
     - "traefik.http.routers.backend.entrypoints=websecure"
     - "traefik.http.routers.backend.tls=true"
     - "traefik.http.routers.backend.tls.certresolver=letsencrypt"
   ```

5. **Deploy**
   - Push to your git repository
   - Coolify will automatically build and deploy

### Option 2: Docker Compose

For VPS or dedicated server deployment.

#### Prerequisites

- Docker and Docker Compose installed
- SSL/TLS certificate (Let's Encrypt recommended)
- Reverse proxy (Nginx, Traefik, or Caddy)

#### Steps

1. **Clone Repository**
   ```bash
   git clone https://github.com/your-org/go-boilerplate.git
   cd go-boilerplate
   ```

2. **Configure Environment**
   ```bash
   cp apps/backend/.env.example apps/backend/.env
   # Edit .env with production values
   ```

3. **Build and Start**
   ```bash
   docker compose -f docker-compose.yml up -d --build
   ```

4. **Run Migrations**
   ```bash
   docker compose exec backend task migrations:up
   ```

5. **Verify Health**
   ```bash
   curl http://localhost:8080/health
   ```

### Option 3: Kubernetes

For scalable, enterprise deployments.

#### Manifests

Create Kubernetes manifests:

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-boilerplate
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-boilerplate
  template:
    metadata:
      labels:
        app: go-boilerplate
    spec:
      containers:
      - name: backend
        image: your-registry/go-boilerplate:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secrets
              key: password
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: go-boilerplate
spec:
  selector:
    app: go-boilerplate
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-boilerplate
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.yourdomain.com
    secretName: api-tls
  rules:
  - host: api.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: go-boilerplate
            port:
              number: 80
```

Deploy:
```bash
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml
```

---

## Production Configuration

### Security Checklist

- [ ] Use strong, random `AUTH_SECRET_KEY`
- [ ] Enable SSL/TLS for database (`DATABASE_SSL_MODE=require`)
- [ ] Set specific CORS origins (never `*`)
- [ ] Use secrets manager for sensitive values
- [ ] Enable Redis authentication
- [ ] Use non-root user in containers (already configured)
- [ ] Keep dependencies updated
- [ ] Enable rate limiting
- [ ] Use distroless images (already configured)
- [ ] Configure firewall rules
- [ ] Enable DDoS protection

### Environment Variables

See [Configuration Reference](../reference/CONFIGURATION.md) for complete list.

**Critical Production Settings**:

```bash
PRIMARY_ENV=production
SERVER_CORS_ALLOWED_ORIGINS=https://yourdomain.com
DATABASE_SSL_MODE=require
DATABASE_MAX_OPEN_CONNS=50
LOG_LEVEL=info
```

### Resource Limits

Recommended resource allocation:

**Minimum**:
- CPU: 0.5 core
- Memory: 512 MB
- Disk: 10 GB

**Recommended**:
- CPU: 2 cores
- Memory: 2 GB
- Disk: 50 GB

**For High Traffic**:
- CPU: 4+ cores
- Memory: 4+ GB
- Disk: 100+ GB
- Consider horizontal scaling

---

## Database Setup

### PostgreSQL

#### Managed Services (Recommended)

Use managed PostgreSQL services:
- AWS RDS
- Google Cloud SQL
- Azure Database for PostgreSQL
- DigitalOcean Managed Databases
- Supabase

#### Self-Hosted

```bash
# Create production database
createdb -h your-host -U postgres production_db

# Create user
psql -h your-host -U postgres -c "CREATE USER app_user WITH PASSWORD 'strong_password';"

# Grant permissions
psql -h your-host -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE production_db TO app_user;"
```

#### Connection Pooling

Consider using PgBouncer for connection pooling:

```ini
[databases]
production_db = host=postgres port=5432 dbname=production_db

[pgbouncer]
pool_mode = transaction
max_client_conn = 100
default_pool_size = 20
```

### Redis

#### Managed Services (Recommended)

- AWS ElastiCache
- Redis Cloud
- DigitalOcean Managed Redis
- Upstash

#### Security

```bash
# Enable authentication
requirepass your_strong_password

# Disable dangerous commands
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG ""
```

---

## Monitoring & Observability

### New Relic APM

1. **Sign up** at [newrelic.com](https://newrelic.com)

2. **Get License Key**
   - Go to Account Settings
   - Copy license key

3. **Configure Application**
   ```bash
   OBSERVABILITY_NEWRELIC_LICENSE_KEY=your_license_key
   OBSERVABILITY_NEWRELIC_APP_NAME=your-app-production
   ```

4. **View Metrics**
   - Response times
   - Error rates
   - Database queries
   - External services
   - Custom transactions

### Logs

**Centralized Logging**:

Options:
- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **CloudWatch Logs** (AWS)
- **Google Cloud Logging**

**Log Aggregation**:

```bash
# Docker logs to stdout (already configured)
docker logs -f container_id

# Or use log driver
docker run --log-driver=syslog ...
```

### Health Checks

Monitor health endpoint:

```bash
# Simple monitoring
while true; do
  curl -f http://your-api/health || alert
  sleep 60
done

# Or use monitoring service
# - UptimeRobot
# - Pingdom
# - StatusCake
```

### Metrics

Key metrics to monitor:

- **Response Time**: p50, p95, p99
- **Error Rate**: 4xx, 5xx errors
- **Throughput**: Requests per second
- **Database**:
  - Connection pool usage
  - Query performance
  - Slow queries
- **Redis**:
  - Memory usage
  - Hit rate
  - Connection count
- **System**:
  - CPU usage
  - Memory usage
  - Disk I/O
  - Network I/O

---

## Backups

### Automated Backups

Already configured via `db-backup` service:

```yaml
# docker-compose.yml
db-backup:
  image: postgres:18-alpine
  environment:
    BACKUP_CRON: "0 */6 * * *"  # Every 6 hours
    BACKUP_RETENTION_DAYS: "14"
```

### Manual Backup

```bash
# Using Makefile
make backup-run

# Or using task
cd apps/backend
task backup:run
```

### Restore

```bash
cd apps/backend
task backup:restore URI=s3://bucket/path/to/backup.sql.zst
```

### Backup Strategy

**3-2-1 Rule**:
- **3** copies of data
- **2** different storage types
- **1** copy offsite

**Our Implementation**:
1. Live database (production)
2. Local backup (Docker volume)
3. S3 backup (Cloudflare R2)

### Verify Backups

Regularly test restore process:

```bash
# 1. Download backup
# 2. Restore to test environment
# 3. Verify data integrity
# 4. Document process
```

---

## Disaster Recovery

### Recovery Time Objective (RTO)

Target: < 1 hour

**Steps**:
1. Spin up new infrastructure
2. Restore database from latest backup
3. Update DNS records
4. Verify application health

### Recovery Point Objective (RPO)

Target: < 6 hours (backup frequency)

To reduce RPO:
- Increase backup frequency
- Enable database replication
- Use point-in-time recovery (PITR)

### DR Checklist

- [ ] Document recovery procedures
- [ ] Store backups in multiple regions
- [ ] Test recovery process quarterly
- [ ] Maintain infrastructure as code
- [ ] Keep DNS TTL low for faster failover
- [ ] Have standby database replica

---

## Scaling

### Vertical Scaling

Increase resources for single instance:

```bash
# Docker Compose
resources:
  limits:
    cpus: '4'
    memory: 4G
```

### Horizontal Scaling

Run multiple instances:

**Docker Compose**:
```bash
docker compose up -d --scale backend=3
```

**Kubernetes**:
```bash
kubectl scale deployment go-boilerplate --replicas=5
```

### Load Balancing

Use load balancer:
- **Cloud**: AWS ALB, GCP Load Balancer
- **Self-hosted**: Nginx, HAProxy, Traefik

**Nginx Example**:
```nginx
upstream backend {
    least_conn;
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}

server {
    listen 80;
    server_name api.yourdomain.com;
    
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Database Scaling

- **Read Replicas**: Distribute read traffic
- **Connection Pooling**: PgBouncer, PgPool
- **Sharding**: For very large datasets

### Caching

- **Application-level**: Redis cache
- **CDN**: CloudFlare, Fastly
- **Reverse Proxy**: Nginx, Varnish

---

## SSL/TLS

### Let's Encrypt with Traefik

Already configured in `docker-compose.yml`:

```yaml
traefik:
  command:
    - "--certificatesresolvers.letsencrypt.acme.email=your@email.com"
    - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
```

### Let's Encrypt with Certbot

```bash
# Install certbot
apt-get install certbot

# Get certificate
certbot certonly --standalone -d api.yourdomain.com

# Auto-renew
certbot renew --dry-run
```

### Nginx SSL Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    
    location / {
        proxy_pass http://localhost:8080;
    }
}
```

---

## Deployment Checklist

### Pre-Deployment

- [ ] Run all tests locally
- [ ] Check linting passes
- [ ] Review code changes
- [ ] Update documentation
- [ ] Test in staging environment
- [ ] Backup production database
- [ ] Notify team about deployment

### Deployment

- [ ] Put maintenance page (if needed)
- [ ] Deploy new version
- [ ] Run database migrations
- [ ] Verify health checks pass
- [ ] Check logs for errors
- [ ] Test critical user flows
- [ ] Remove maintenance page
- [ ] Monitor metrics for 30 minutes

### Post-Deployment

- [ ] Verify backup completed successfully
- [ ] Check error rates in APM
- [ ] Monitor response times
- [ ] Review logs for anomalies
- [ ] Update deployment documentation
- [ ] Notify team of successful deployment

### Rollback Plan

If issues occur:

1. **Quick Rollback**:
   ```bash
   # Kubernetes
   kubectl rollout undo deployment/go-boilerplate
   
   # Docker
   docker compose down
   docker compose up -d previous-tag
   ```

2. **Database Rollback**:
   ```bash
   task migrations:down
   ```

3. **Verify**:
   - Check health endpoint
   - Test critical features
   - Review logs

---

## Troubleshooting

### Application Won't Start

1. Check logs:
   ```bash
   docker logs container_name
   ```

2. Verify environment variables
3. Check database connectivity
4. Ensure migrations are applied

### High Memory Usage

1. Check connection pool settings
2. Review goroutine leaks
3. Use pprof for profiling:
   ```bash
   go tool pprof http://localhost:8080/debug/pprof/heap
   ```

### Slow Queries

1. Enable query logging
2. Use `EXPLAIN ANALYZE` in PostgreSQL
3. Add database indexes
4. Consider caching

### Connection Pool Exhaustion

1. Increase `DATABASE_MAX_OPEN_CONNS`
2. Use connection pooler (PgBouncer)
3. Optimize query performance
4. Scale horizontally

---

## What's Next?

- **Configure monitoring**: [CI/CD Guide](./CI_CD.md)
- **Security hardening**: [Best Practices](../development/BEST_PRACTICES.md#security)
- **Performance tuning**: [Architecture Guide](../reference/ARCHITECTURE.md)
