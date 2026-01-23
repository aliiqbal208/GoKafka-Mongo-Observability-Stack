# 🎯 System Architecture & Component Documentation

This document provides a comprehensive explanation of the system architecture and how all components work together in this microservice.

## 📑 Table of Contents

- [System Architecture Overview](#-system-architecture-overview)
- [Request Flow Examples](#-request-flow-examples)
  - [CREATE Product (Write Operation)](#1-create-product-write-operation)
  - [GET Product by ID (Read Operation)](#2-get-product-by-id-read-operation)
  - [UPDATE Product (Write Operation)](#3-update-product-write-operation)
  - [LIST Products with Pagination (Read Operation)](#4-list-products-with-pagination-read-operation)
- [Component Breakdown](#-component-breakdown)
- [Observability Stack](#-observability-stack)
- [How Components Connect](#-how-components-connect)
- [Why This Architecture?](#-why-this-architecture)
- [Design Patterns](#-design-patterns)
- [Performance Considerations](#-performance-considerations)
- [Security Considerations](#-security-considerations)
- [Troubleshooting Guide](#-troubleshooting-guide)
- [Deployment Notes](#-deployment-notes)

---

## 🏗️ System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                   CLIENT                                         │
│                        (curl, Postman, Frontend App)                            │
└─────────────────────────────────┬───────────────────────────────────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    ▼                           ▼
            ┌───────────────┐           ┌───────────────┐
            │  HTTP (Fiber) │           │     gRPC      │
            │   Port 5007   │           │   Port 5000   │
            └───────┬───────┘           └───────┬───────┘
                    │                           │
                    └─────────────┬─────────────┘
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           MICROSERVICE (Go)                                      │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                         HANDLERS / DELIVERY                              │    │
│  │   • Validates request (go-playground/validator)                         │    │
│  │   • Routes to UseCase                                                    │    │
│  └─────────────────────────────────┬───────────────────────────────────────┘    │
│                                    ▼                                             │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                            USE CASE                                      │    │
│  │   • Business logic                                                       │    │
│  │   • Decides: Publish to Kafka OR Read from MongoDB/Redis                │    │
│  └──────────┬──────────────────────────────────┬───────────────────────────┘    │
│             │                                  │                                 │
│             ▼                                  ▼                                 │
│  ┌──────────────────────┐           ┌──────────────────────┐                    │
│  │   KAFKA PRODUCER     │           │    REPOSITORIES      │                    │
│  │  (Create/Update)     │           │  (Read Operations)   │                    │
│  └──────────┬───────────┘           └──────────┬───────────┘                    │
│             │                                  │                                 │
└─────────────┼──────────────────────────────────┼─────────────────────────────────┘
              │                                  │
              ▼                                  ▼
┌─────────────────────────┐           ┌──────────────────────────────────────────┐
│      KAFKA CLUSTER      │           │              DATA STORES                  │
│  ┌───────────────────┐  │           │  ┌─────────────┐    ┌─────────────────┐  │
│  │ kafka1 (19091)    │  │           │  │   MongoDB   │    │      Redis      │  │
│  │ kafka2 (19092)    │  │           │  │  (Primary)  │    │    (Cache)      │  │
│  │ kafka3 (19093)    │  │           │  │  Port 27017 │    │   Port 6379     │  │
│  └───────────────────┘  │           │  └─────────────┘    └─────────────────┘  │
│  Topics:                │           └──────────────────────────────────────────┘
│  • create-product       │                        ▲
│  • update-product       │                        │
└──────────┬──────────────┘                        │
           │                                       │
           ▼                                       │
┌─────────────────────────┐                        │
│    KAFKA CONSUMER       │                        │
│  (Consumer Group)       │────────────────────────┘
│  • Reads messages       │     Writes to MongoDB
│  • Processes async      │
└─────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                            OBSERVABILITY                                         │
│  ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐                │
│  │   Prometheus    │   │     Grafana     │   │     Jaeger      │                │
│  │   Port 9090     │   │   Port 3030     │   │   Port 16686    │                │
│  │  (Metrics)      │   │  (Dashboards)   │   │   (Tracing)     │                │
│  └─────────────────┘   └─────────────────┘   └─────────────────┘                │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Request Flow Examples

### 1. CREATE Product (Write Operation)

```
Client                    Fiber Handler              UseCase                 Kafka                  Consumer               MongoDB
  │                            │                        │                      │                        │                     │
  │  POST /api/v1/products     │                        │                      │                        │                     │
  │ ─────────────────────────► │                        │                      │                        │                     │
  │                            │  Validate & Parse      │                      │                        │                     │
  │                            │ ─────────────────────► │                      │                        │                     │
  │                            │                        │  PublishCreate()     │                        │                     │
  │                            │                        │ ───────────────────► │                        │                     │
  │                            │                        │                      │  Message to            │                     │
  │                            │                        │                      │  "create-product"      │                     │
  │                            │                        │                      │ ─────────────────────► │                     │
  │   201 Created              │                        │                      │                        │  Insert Product     │
  │ ◄───────────────────────── │                        │                      │                        │ ──────────────────► │
  │                            │                        │                      │                        │                     │
  │   (Async - happens later)  │                        │                      │                        │   Product Saved     │
  │                            │                        │                      │                        │ ◄────────────────── │
```

**Key Points:**
- Response returns **immediately** after Kafka publish (201 Created)
- MongoDB insert happens **asynchronously** via Kafka Consumer
- This is **eventual consistency** pattern
- If Kafka is down, the request will fail immediately
- Consumer processes messages in order per partition

---

### 2. UPDATE Product (Write Operation)

```
Client                    Fiber Handler              UseCase                 Kafka                  Consumer               MongoDB
  │                            │                        │                      │                        │                     │
  │  PUT /api/v1/products/:id  │                        │                      │                        │                     │
  │ ─────────────────────────► │                        │                      │                        │                     │
  │                            │  Validate & Parse      │                      │                        │                     │
  │                            │ ─────────────────────► │                      │                        │                     │
  │                            │                        │  PublishUpdate()     │                        │                     │
  │                            │                        │ ───────────────────► │                        │                     │
  │                            │                        │                      │  Message to            │                     │
  │                            │                        │                      │  "update-product"      │                     │
  │                            │                        │                      │ ─────────────────────► │                     │
  │   200 OK                   │                        │                      │                        │  Update Product     │
  │ ◄───────────────────────── │                        │                      │                        │ ──────────────────► │
  │                            │                        │                      │                        │                     │
  │                            │                        │                      │                        │  Invalidate Cache   │
  │                            │                        │                      │                        │ ──────────────────► │
```

**Key Points:**
- Similar to CREATE, update is asynchronous
- Consumer invalidates Redis cache after updating MongoDB
- Ensures cache consistency with database

---

### 3. GET Product by ID (Read Operation)

```
Client                    Fiber Handler              UseCase                 Redis                   MongoDB
  │                            │                        │                      │                        │
  │  GET /api/v1/products/:id  │                        │                      │                        │
  │ ─────────────────────────► │                        │                      │                        │
  │                            │  Parse product_id      │                      │                        │
  │                            │ ─────────────────────► │                      │                        │
  │                            │                        │  Check Cache         │                        │
  │                            │                        │ ───────────────────► │                        │
  │                            │                        │                      │                        │
  │                            │                        │  Cache MISS          │                        │
  │                            │                        │ ◄─────────────────── │                        │
  │                            │                        │                      │                        │
  │                            │                        │  Query MongoDB       │                        │
  │                            │                        │ ──────────────────────────────────────────► │
  │                            │                        │                      │                        │
  │                            │                        │  Product Data        │                        │
  │                            │                        │ ◄──────────────────────────────────────────── │
  │                            │                        │                      │                        │
  │                            │                        │  Set Cache           │                        │
  │                            │                        │ ───────────────────► │                        │
  │                            │                        │                      │                        │
  │   200 OK + Product JSON    │                        │                      │                        │
  │ ◄───────────────────────── │                        │                      │                        │
```

**Key Points:**
- First checks **Redis cache** for fast retrieval
- On cache miss, queries **MongoDB**
- Stores result in Redis for future requests (with TTL)
- Cache-aside pattern implementation

---

### 4. LIST Products with Pagination (Read Operation)

```
Client                    Fiber Handler              UseCase                 MongoDB
  │                            │                        │                      │
  │  GET /api/v1/products      │                        │                      │
  │  ?page=1&size=10           │                        │                      │
  │ ─────────────────────────► │                        │                      │
  │                            │  Parse query params    │                      │
  │                            │ ─────────────────────► │                      │
  │                            │                        │  Query with          │
  │                            │                        │  Skip & Limit        │
  │                            │                        │ ───────────────────► │
  │                            │                        │                      │
  │                            │                        │  Products + Count    │
  │                            │                        │ ◄─────────────────── │
  │                            │                        │                      │
  │   200 OK + Paginated Data  │                        │                      │
  │ ◄───────────────────────── │                        │                      │
  │   {                        │                        │                      │
  │     "products": [...],     │                        │                      │
  │     "total": 150,          │                        │                      │
  │     "page": 1,             │                        │                      │
  │     "size": 10             │                        │                      │
  │   }                        │                        │                      │
```

**Key Points:**
- Direct MongoDB query (no Kafka involved)
- Pagination reduces memory usage and improves response time
- Total count helps build pagination UI

---

## 🧩 Component Breakdown

### 1. **Fiber (HTTP Server)** - Port 5007
```go
// internal/server/http.go
s.app.Get("/health", ...)           // Health check
s.app.Get("/api/v1/products", ...)  // REST endpoints
```
- Handles HTTP requests
- Middleware: Logger, CORS, Recovery, RequestID, Compression
- Routes requests to handlers

### 2. **gRPC Server** - Port 5000
```protobuf
// proto/product/product.proto
service ProductsService {
  rpc Create(CreateReq) returns (CreateRes);
  rpc Update(UpdateReq) returns (UpdateRes);
  rpc GetByID(GetByIDReq) returns (GetByIDRes);
}
```
- Binary protocol (faster than REST)
- Used for service-to-service communication

### 3. **Kafka** - Ports 9091, 9092, 9093
```
Topics:
├── create-product   → New product messages
└── update-product   → Product update messages
```
- **Producer**: Publishes messages when create/update called
- **Consumer**: Reads messages and writes to MongoDB
- **3 Brokers**: High availability, replication

### 4. **MongoDB** - Port 27017
```javascript
// Database: products
// Collection: products
{
  "_id": ObjectId("..."),
  "name": "Product Name",
  "description": "...",
  "price": 29.99,
  "quantity": 10,
  ...
}
```
- Primary data store
- Document-based (flexible schema)

### 5. **Redis** - Port 6379
```
Key Pattern: product:{id}
TTL: Configurable (e.g., 5 minutes)
```
- Caching layer
- Reduces MongoDB load
- Faster reads for hot data

### 6. **Zookeeper** - Port 2181
- Kafka cluster coordination
- Broker discovery and health monitoring
- Topic metadata and partition management
- Leader election for partitions
- Configuration management

### 7. **Jaeger** - Port 16686
```go
// Example tracing implementation
span, ctx := opentracing.StartSpanFromContext(ctx, "productHandlers.Create")
defer span.Finish()
```
- Distributed request tracing
- Performance monitoring across services
- Request flow visualization
- Latency analysis and bottleneck detection

---

## 📊 Observability Stack

### Prometheus (Port 9090)
```yaml
# Scrapes metrics from:
- microservice:7070/metrics   # App metrics
- node-exporter:9100         # System metrics
```
Collects:
- HTTP request counts
- Response times
- Kafka producer/consumer metrics

### Grafana (Port 3030)
- Visualizes Prometheus data
- Dashboards for monitoring
- Default login: admin/admin

### Jaeger (Port 16686)
```go
// Tracing spans created for:
span, ctx := opentracing.StartSpanFromContext(ctx, "productHandlers.Create")
```
- Distributed tracing
- Request flow visualization
- Performance bottleneck detection

---

## 🔗 How Components Connect

```yaml
# docker-compose.yml network
networks:
  products_network:
    driver: bridge

# All containers on same network can communicate via container names:
# - microservice → kafka1:19091
# - microservice → mongodb:27017
# - microservice → redis:6379
# - prometheus → microservice:7070
```

| From | To | Purpose |
|------|-----|---------|
| Client | microservice:5007 | HTTP API calls |
| Client | microservice:5000 | gRPC calls |
| microservice | kafka1/2/3:19091-93 | Publish messages |
| microservice | mongodb:27017 | Read/Write data |
| microservice | redis:6379 | Cache operations |
| kafka-consumer | mongodb:27017 | Async writes |
| prometheus | microservice:7070 | Scrape metrics |
| kafdrop | kafka1:19091 | Kafka UI |

---

## 💡 Why This Architecture?

1. **Decoupling**: Kafka separates write operations from actual database writes
   - Services don't need to wait for database operations
   - Producers and consumers evolve independently
   
2. **Scalability**: Can scale consumers independently
   - Add more consumer instances to handle load
   - Kafka partitions enable parallel processing
   
3. **Resilience**: If MongoDB is slow, writes are buffered in Kafka
   - Messages are persisted and replayed if consumer fails
   - No data loss even during service outages
   
4. **Performance**: Redis cache reduces database load
   - Sub-millisecond read latency for cached data
   - Significantly reduced MongoDB query load
   
5. **Observability**: Full visibility into system behavior
   - Metrics, logs, and traces for debugging
   - Real-time monitoring and alerting

---

## 🎨 Design Patterns

### 1. **Event-Driven Architecture**
- Asynchronous communication via Kafka
- Producers publish events without waiting for consumers
- Loose coupling between components

### 2. **Cache-Aside Pattern**
```go
// Pseudocode
func GetProduct(id string) (*Product, error) {
    // 1. Try cache first
    product, err := cache.Get(id)
    if err == nil {
        return product, nil
    }
    
    // 2. Cache miss - query database
    product, err = db.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    // 3. Store in cache for next time
    cache.Set(id, product, ttl)
    return product, nil
}
```

### 3. **Clean Architecture (Hexagonal)**
```
┌─────────────────────────────────────┐
│         Delivery Layer              │  ← HTTP/gRPC handlers
├─────────────────────────────────────┤
│         Use Case Layer              │  ← Business logic
├─────────────────────────────────────┤
│       Repository Layer              │  ← Data access
└─────────────────────────────────────┘
```
- Separation of concerns
- Easy to test and maintain
- Framework independence

### 4. **Repository Pattern**
```go
type ProductRepository interface {
    Create(ctx context.Context, product *Product) error
    GetByID(ctx context.Context, id string) (*Product, error)
    Update(ctx context.Context, product *Product) error
    Delete(ctx context.Context, id string) error
    GetAll(ctx context.Context, pagination *Pagination) ([]*Product, error)
}
```
- Abstracts data access layer
- Enables easy testing with mocks
- Swappable data sources

### 5. **Circuit Breaker Pattern** (Implicit via Retry)
- Retry mechanism for transient failures
- Prevents cascading failures
- Graceful degradation

---

## ⚡ Performance Considerations

### 1. **Redis Caching Strategy**
```yaml
Cache Hit Ratio Target: > 80%
TTL: 5-15 minutes (configurable)
Memory Limit: Set maxmemory policy
Eviction: LRU (Least Recently Used)
```

**Best Practices:**
- Cache frequently accessed products
- Set appropriate TTL based on data update frequency
- Monitor cache hit/miss ratio
- Use Redis pipelining for bulk operations

### 2. **Kafka Optimization**
```yaml
Producer Config:
  acks: 1                    # Balance between durability and performance
  compression.type: snappy   # Reduce network bandwidth
  batch.size: 16384          # Batch messages for efficiency
  linger.ms: 10              # Wait time to batch messages

Consumer Config:
  fetch.min.bytes: 1024      # Minimum data to fetch
  fetch.max.wait.ms: 500     # Max wait time for minimum data
  max.poll.records: 500      # Records per poll
```

**Partitioning Strategy:**
- Use product_id as partition key for ordering
- Number of partitions = expected consumer count
- Ensures related events are ordered

### 3. **MongoDB Indexing**
```javascript
// Essential indexes
db.products.createIndex({ "_id": 1 })          // Primary key (automatic)
db.products.createIndex({ "name": 1 })         // Search by name
db.products.createIndex({ "category": 1 })     // Filter by category
db.products.createIndex({ "created_at": -1 })  // Sort by creation date

// Compound index for common queries
db.products.createIndex({ "category": 1, "price": 1 })
```

**Query Optimization:**
- Use projection to fetch only required fields
- Implement pagination to limit result sets
- Monitor slow queries with MongoDB profiler

### 4. **Connection Pooling**
```go
// MongoDB pool settings
ClientOptions().
    SetMaxPoolSize(100).
    SetMinPoolSize(10).
    SetMaxConnIdleTime(30 * time.Second)

// Redis pool settings
&redis.Options{
    PoolSize:     100,
    MinIdleConns: 10,
    PoolTimeout:  4 * time.Second,
}
```

### 5. **Graceful Shutdown**
- Complete in-flight requests
- Flush Kafka producer buffer
- Close database connections properly
- Deregister from service discovery

---

## 🔒 Security Considerations

### 1. **Authentication & Authorization**
```go
// Add JWT middleware (example)
app.Use(jwtware.New(jwtware.Config{
    SigningKey: []byte(secret),
}))

// Role-based access control
func AdminOnly(c *fiber.Ctx) error {
    role := c.Locals("role").(string)
    if role != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
    }
    return c.Next()
}
```

### 2. **Input Validation**
```go
type CreateProductRequest struct {
    Name        string  `json:"name" validate:"required,min=3,max=100"`
    Description string  `json:"description" validate:"required,max=500"`
    Price       float64 `json:"price" validate:"required,gt=0"`
    Quantity    int     `json:"quantity" validate:"required,gte=0"`
}
```

### 3. **Rate Limiting**
```go
// Fiber rate limiter
app.Use(limiter.New(limiter.Config{
    Max:        100,              // 100 requests
    Expiration: 1 * time.Minute,  // per minute
}))
```

### 4. **MongoDB Security**
```yaml
# Production settings
- Enable authentication (SCRAM-SHA-256)
- Use TLS/SSL for connections
- Implement role-based access control
- Regular security patches
- Audit logging enabled
```

### 5. **Kafka Security**
```yaml
# Production Kafka security
security.protocol: SASL_SSL
sasl.mechanism: SCRAM-SHA-512
sasl.username: ${KAFKA_USER}
sasl.password: ${KAFKA_PASSWORD}
ssl.ca.location: /path/to/ca-cert
```

### 6. **Secrets Management**
```bash
# Use environment variables (never hardcode)
export MONGODB_PASSWORD=$(aws secretsmanager get-secret-value --secret-id prod/db/password)
export REDIS_PASSWORD=$(vault kv get -field=password secret/redis)
export KAFKA_API_KEY=$(kubectl get secret kafka-secret -o jsonpath='{.data.api-key}' | base64 -d)
```

---

## 🔧 Troubleshooting Guide

### Common Issues & Solutions

#### 1. **Kafka Consumer Not Processing Messages**

**Symptoms:**
- Messages accumulate in Kafka topics
- Consumer lag increases
- No database updates

**Debugging Steps:**
```bash
# Check consumer group status
docker exec -it kafka1 kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group products-consumer-group

# Check Kafdrop UI
open http://localhost:9000

# Check container logs
docker logs -f consumer-container
```

**Common Fixes:**
- Restart consumer service
- Check MongoDB connection
- Verify consumer group configuration
- Check for processing errors in logs

#### 2. **Redis Cache Issues**

**Symptoms:**
- Slow read operations
- High MongoDB load
- Cache miss rate > 50%

**Debugging Steps:**
```bash
# Connect to Redis
docker exec -it redis redis-cli

# Check cache stats
INFO stats

# Monitor cache operations
MONITOR

# Check memory usage
INFO memory

# View all keys
KEYS *
```

**Common Fixes:**
- Increase Redis memory limit
- Adjust TTL values
- Implement cache warming
- Use Redis persistent storage

#### 3. **MongoDB Connection Failures**

**Symptoms:**
- Connection timeout errors
- "no reachable servers" error
- Slow queries

**Debugging Steps:**
```bash
# Check MongoDB status
docker exec -it mongodb mongosh

# Check connection pool
db.serverStatus().connections

# Check slow queries
db.setProfilingLevel(2)
db.system.profile.find().pretty()

# Check indexes
db.products.getIndexes()
```

**Common Fixes:**
- Increase connection pool size
- Add missing indexes
- Optimize slow queries
- Scale MongoDB (replica set)

#### 4. **High Latency / Slow Responses**

**Symptoms:**
- Response time > 1 second
- Timeout errors
- High CPU/Memory usage

**Debugging Steps:**
```bash
# Check Jaeger traces
open http://localhost:16686

# Check Prometheus metrics
open http://localhost:9090

# Monitor Grafana dashboards
open http://localhost:3030

# Check container resources
docker stats

# Profile Go application
curl http://localhost:8100/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

**Common Fixes:**
- Optimize database queries
- Increase cache hit ratio
- Scale horizontally (add instances)
- Optimize Kafka consumer batch size

#### 5. **Kafka Broker Down**

**Symptoms:**
- Producer publish failures
- Consumer stops consuming
- Broker unreachable errors

**Debugging Steps:**
```bash
# Check broker status
docker ps | grep kafka

# Check Zookeeper connection
docker exec -it zookeeper zkCli.sh
ls /brokers/ids

# Check broker logs
docker logs kafka1
docker logs kafka2
docker logs kafka3
```

**Common Fixes:**
- Restart failed broker
- Check Zookeeper health
- Verify network connectivity
- Check disk space

#### 6. **Memory Leaks**

**Symptoms:**
- Gradual memory increase
- OOM (Out of Memory) errors
- Container restarts

**Debugging Steps:**
```bash
# Monitor memory usage
docker stats --no-stream

# Get heap dump
curl http://localhost:8100/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Check goroutine leaks
curl http://localhost:8100/debug/pprof/goroutine?debug=2
```

**Common Fixes:**
- Close database connections
- Cancel contexts properly
- Fix goroutine leaks
- Implement proper resource cleanup

---

## 🚀 Deployment Notes

### Local Development
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f microservice

# Rebuild after code changes
docker-compose up -d --build microservice

# Stop all services
docker-compose down

# Clean volumes (reset data)
docker-compose down -v
```

### Production Deployment Checklist

#### Infrastructure
- [ ] Use managed Kafka (AWS MSK, Confluent Cloud)
- [ ] Use MongoDB Atlas or managed MongoDB
- [ ] Use ElastiCache for Redis
- [ ] Set up load balancer (ALB/NLB)
- [ ] Configure auto-scaling groups
- [ ] Set up VPC and security groups
- [ ] Enable encryption at rest and in transit

#### Configuration
- [ ] Use production-grade config management (AWS Secrets Manager, HashiCorp Vault)
- [ ] Set appropriate resource limits (CPU, memory)
- [ ] Configure health checks
- [ ] Set up log aggregation (CloudWatch, ELK)
- [ ] Enable distributed tracing (Jaeger, AWS X-Ray)
- [ ] Configure monitoring and alerting

#### Security
- [ ] Enable TLS/SSL for all services
- [ ] Implement authentication (JWT, OAuth2)
- [ ] Set up API gateway (rate limiting, WAF)
- [ ] Regular security scanning
- [ ] Rotate secrets regularly
- [ ] Follow least privilege principle

#### Performance
- [ ] Configure appropriate database indexes
- [ ] Set up CDN for static assets
- [ ] Implement connection pooling
- [ ] Configure caching strategies
- [ ] Set up read replicas for MongoDB
- [ ] Use Kafka partitioning effectively

#### Monitoring
- [ ] Set up Prometheus/Grafana dashboards
- [ ] Configure alerts (PagerDuty, OpsGenie)
- [ ] Monitor error rates and latencies
- [ ] Track business metrics
- [ ] Set up log-based alerts
- [ ] Regular performance reviews

### Container Orchestration (Kubernetes)

```yaml
# Example deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: products-microservice
spec:
  replicas: 3
  selector:
    matchLabels:
      app: products-microservice
  template:
    metadata:
      labels:
        app: products-microservice
    spec:
      containers:
      - name: microservice
        image: products-microservice:latest
        ports:
        - containerPort: 5007
        - containerPort: 5000
        env:
        - name: MONGODB_URI
          valueFrom:
            secretKeyRef:
              name: db-secrets
              key: mongodb-uri
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 5007
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 5007
          initialDelaySeconds: 10
          periodSeconds: 5
```

### CI/CD Pipeline

```yaml
# Example GitHub Actions workflow
name: CI/CD Pipeline

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    - name: Upload coverage
      uses: codecov/codecov-action@v3

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Build Docker image
      run: docker build -t products-microservice:${{ github.sha }} .
    - name: Push to registry
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker push products-microservice:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/products-microservice \
          microservice=products-microservice:${{ github.sha }}
        kubectl rollout status deployment/products-microservice
```

---

## 📚 Additional Resources

### Documentation Links
- [Fiber Documentation](https://docs.gofiber.io/)
- [Kafka Documentation](https://kafka.apache.org/documentation/)
- [MongoDB Documentation](https://docs.mongodb.com/)
- [Redis Documentation](https://redis.io/documentation)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)

### Useful Commands Reference

#### Docker
```bash
# View running containers
docker ps

# View all containers (including stopped)
docker ps -a

# View logs
docker logs -f <container-name>

# Execute command in container
docker exec -it <container-name> bash

# Remove all stopped containers
docker container prune

# Remove unused images
docker image prune -a

# View disk usage
docker system df
```

#### Kafka
```bash
# List topics
kafka-topics.sh --bootstrap-server localhost:9092 --list

# Describe topic
kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic create-product

# Consume messages
kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic create-product --from-beginning

# Check consumer group lag
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group products-consumer-group
```

#### MongoDB
```bash
# Connect to MongoDB
mongosh mongodb://localhost:27017

# Show databases
show dbs

# Use database
use products

# Show collections
show collections

# Query products
db.products.find().pretty()

# Count documents
db.products.countDocuments()

# Create index
db.products.createIndex({ name: 1 })
```

#### Redis
```bash
# Connect to Redis
redis-cli

# Get all keys
KEYS *

# Get value
GET product:123

# Delete key
DEL product:123

# Check key TTL
TTL product:123

# Flush all data (careful!)
FLUSHALL
```

---

## 🤝 Contributing

When contributing to this project, please ensure:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Write unit tests for new features
   - Maintain test coverage above 80%
   - Use meaningful variable and function names

2. **Documentation**
   - Update this document for architectural changes
   - Add inline comments for complex logic
   - Update Swagger documentation
   - Include examples in commit messages

3. **Testing**
   - Test locally before committing
   - Run `make test` to execute all tests
   - Verify Docker Compose setup works
   - Test both success and error scenarios

4. **Git Workflow**
   - Create feature branches from `main`
   - Write descriptive commit messages
   - Keep commits atomic and focused
   - Rebase before merging to keep history clean

---

Would you like me to dive deeper into any specific component or add more sections?