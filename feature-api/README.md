# Feature Flag API

High-performance Go service for feature flag management and evaluation. Supports multi-tier caching (L1 LRU, L2 Redis) and advanced rollout rules.

## Configuration

Set via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| "PORT" | API Port | "8080" |
| "MONGO_URI" | MongoDB Connection String | "mongodb://localhost:27017" |
| "MONGO_DB_NAME" | Database Name | "feature_flags" |
| "MONGO_COLLECTION_NAME" | Collection Name | "flags" |
| "REDIS_ADDR" | Redis Address | "localhost:6379" |
| "REDIS_PASSWORD" | Redis Password | "" |
| "REDIS_CACHE_PREFIX" | Redis Key Prefix | "flags:id:" |
| "API_KEY" | Auth Key for X-API-KEY header | "" |
| "CACHE_TTL_SECONDS" | Redis Cache TTL | "30" |
| "LOG_LEVEL" | Logging level (debug, info, warn, error) | "info" |

---

## API Endpoints

All requests (except "/health") require header: "X-API-KEY: <your-key>"

- "GET /api/flags" - List all flags
- "POST /api/flags" - Create a flag
- "GET /api/flags/{id}" - Get flag details
- "PATCH /api/flags/{id}" - Update a flag
- "DELETE /api/flags/{id}" - Delete a flag
- "POST /api/flags/{id}/evaluate" - Evaluate flag for a user context

---

## Rule Match Strategy

Flags support two strategies for evaluating multiple rules via the "ruleMatchStrategy" field:

- **"any" (Default)**: Short-circuit OR logic. Returns the value of the **first** rule that matches.
- **"all"**: AND logic. Returns the value of the **last** rule only if **every** rule matches. If any rule fails, returns "defaultValue".

### Example (ALL Strategy):
```json
{
  "key": "strict-feature",
  "ruleMatchStrategy": "all",
  "defaultValue": false,
  "rules": [
    { "type": "geography", "config": { "countries": ["US"] }, "value": true },
    { "type": "attribute", "config": { "attributeKey": "beta", "attributeOp": "eq", "attributeValue": "true" }, "value": true }
  ]
}
```
*(Only returns "true" if user is in US **AND** has beta=true attribute)*

---

## Rule Types

### 1. Percentage
Buckets users based on a deterministic hash of their "userId".
- **Config**: "{\"percentage\": float}" (0-100, supports 0.01 precision)
- **Evaluation Request**:
```json
{ "userId": "user-123" }
```
- **Example Create JSON**:
```json
{
  "key": "new-ui",
  "name": "New UI Layout",
  "enabled": true,
  "defaultValue": false,
  "ruleMatchStrategy": "any",
  "rules": [
    { "type": "percentage", "config": { "percentage": 10.5 }, "value": true }
  ]
}
```

### 2. User List
Matches specific User IDs.
- **Config**: "{\"userIds\": [\"id1\", \"id2\"]}"
- **Evaluation Request**:
```json
{ "userId": "admin-1" }
```

### 3. Attribute
Matches custom user attributes. Supports "eq", "neq", "contains" (supports comma-separated lists), "gt", "lt".
- **Config**: "{\"attributeKey\": \"string\", \"attributeOp\": \"string\", \"attributeValue\": \"string\"}"
- **Evaluation Request**:
```json
{ "attributes": { "plan": "premium" } }
```

### 4. Schedule
Matches based on time windows (UTC).
- **Config**: "{\"enableAt\": \"ISO8601\", \"disableAt\": \"ISO8601\"}"
- **Evaluation Request**:
```json
{} 
```
*(Uses server time)*

### 5. Gradual Rollout
Increases rollout percentage over time.
- **Config**: "{\"startAt\": \"ISO8\", \"endAt\": \"...\", \"startPercent\": 0, \"endPercent\": 100}"
- **Evaluation Request**:
```json
{ "userId": "user-456" }
```

### 6. Geography
Matches based on location fields.
- **Config**: "{\"countries\": [], \"states\": [], \"cities\": [], \"zipCodes\": []}"
- **Evaluation Request**:
```json
{ "country": "US", "city": "New York" }
```

---

## Evaluation

Send a "POST" to "/api/flags/{id}/evaluate" with the user's context.

### Evaluation Request JSON
```json
{
  "userId": "user-123",
  "country": "US",
  "state": "NY",
  "city": "New York",
  "zipCode": "10001",
  "attributes": {
    "plan": "premium",
    "version": "2.0"
  }
}
```

### Evaluation Response JSON
```json
{
  "enabled": true,
  "reason": "matched rule: attribute"
}
```
