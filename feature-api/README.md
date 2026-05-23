# Feature Flag API

High-performance Go service for feature flag management and evaluation. Supports multi-tier caching (L1 LRU, L2 Redis) and advanced rollout rules.

## Configuration

Set via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | API Port | `8080` |
| `MONGO_URI` | MongoDB Connection String | `mongodb://localhost:27017` |
| `MONGO_DB_NAME` | Database Name | `feature_flags` |
| `REDIS_ADDR` | Redis Address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis Password | `""` |
| `API_KEY` | Auth Key for X-API-KEY header | `""` |
| `CACHE_TTL_SECONDS` | Redis Cache TTL | `30` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |

---

## API Endpoints

All requests (except `/health`) require header: `X-API-KEY: <your-key>`

- `GET /api/flags` - List all flags
- `POST /api/flags` - Create a flag
- `GET /api/flags/{id}` - Get flag details
- `PATCH /api/flags/{id}` - Update a flag
- `DELETE /api/flags/{id}` - Delete a flag
- `POST /api/flags/{id}/evaluate` - Evaluate flag for a user context

---

## Rule Types

Flags use a list of rules evaluated in order. The first rule that matches determines the value. If no rules match, the `defaultValue` is used.

### 1. Percentage
Buckets users based on a deterministic hash of their `userId`.
- **Config**: `{"percentage": float}` (0-100)
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
  "rules": [
    { "type": "percentage", "config": { "percentage": 10.5 }, "value": true }
  ]
}
```

### 2. User List
Matches specific User IDs.
- **Config**: `{"userIds": ["id1", "id2"]}`
- **Evaluation Request**:
```json
{ "userId": "admin-1" }
```
- **Example Update JSON**:
```json
{
  "rules": [
    { "type": "user_list", "config": { "userIds": ["admin-1", "tester-2"] }, "value": true }
  ]
}
```

### 3. Attribute
Matches custom user attributes. Supports `eq`, `neq`, `contains`, `gt`, `lt`.
- **Config**: `{"attributeKey": "string", "attributeOp": "string", "attributeValue": "string"}`
- **Evaluation Request**:
```json
{ "attributes": { "plan": "premium" } }
```
- **Example**:
```json
{
  "type": "attribute",
  "config": {
    "attributeKey": "plan",
    "attributeOp": "eq",
    "attributeValue": "premium"
  },
  "value": true
}
```

### 4. Schedule
Matches based on time windows (UTC).
- **Config**: `{"enableAt": "ISO8601", "disableAt": "ISO8601"}`
- **Evaluation Request**:
```json
{} 
```
*(Uses server time)*

- **Example**:
```json
{
  "type": "schedule",
  "config": {
    "enableAt": "2026-06-01T00:00:00Z",
    "disableAt": "2026-06-02T00:00:00Z"
  },
  "value": true
}
```

### 5. Gradual Rollout
Increases rollout percentage over time.
- **Config**: `{"startAt": "ISO8", "endAt": "...", "startPercent": 0, "endPercent": 100}`
- **Evaluation Request**:
```json
{ "userId": "user-456" }
```
- **Example**:
```json
{
  "type": "gradual",
  "config": {
    "startAt": "2026-05-23T00:00:00Z",
    "endAt": "2026-05-30T00:00:00Z",
    "startPercent": 0,
    "endPercent": 100
  },
  "value": true
}
```

### 6. Geography
Matches based on location fields.
- **Config**: `{"countries": [], "states": [], "cities": [], "zipCodes": []}`
- **Evaluation Request**:
```json
{ "country": "US", "city": "New York" }
```
- **Example**:
```json
{
  "type": "geography",
  "config": { "countries": ["US", "CA"], "cities": ["New York"] },
  "value": true
}
```

---

## Evaluation

Send a `POST` to `/api/flags/{id}/evaluate` with the user's context.

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
