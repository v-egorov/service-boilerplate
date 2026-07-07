# Proposal: Fix Pagination Inconsistency in objects-service

## Problem

objects-service has two competing pagination conventions across its 5 list endpoints, making the API confusing for clients and error-prone for developers.

### Current State (broken)

| Endpoint | Params Used | Default | Response Fields |
|----------|------------|---------|-----------------|
| `/api/v1/objects` | `?limit=N&offset=M` | N/A (must specify) | count, limit, offset, total |
| `/api/v1/object-types` | `?limit=N&offset=M` | limit=50, offset=0 | limit, offset, count |
| `/api/v1/relationships` | `?page=P&page_size=S` | page=1, page_size=20 | limit, offset, total |
| `/api/v1/relationship-types` | `?page=P&page_size=S` | page=1, page_size=20 | **none** (raw gin.H) |

### Impact

- Clients must remember which endpoint uses which param format
- Relationship-types returns no pagination metadata at all — a silent bug
- The `(page-1)*size` arithmetic is scattered across handlers and repos instead of being centralized
- Dev mode means backward compatibility is not a concern

## What Changes

- **Standardize on `limit/offset`** as the canonical pagination format for all list endpoints
- **Default limit=50** everywhere (relationships was 20, will be bumped)
- **All handlers return consistent pagination fields**: count, limit, offset, total
- Replace `page/page_size` structs with `limit/offset` in filter models
- Remove `(page-1)*size` arithmetic — repos receive Offset directly

## What Does NOT Change

- Pagination behavior (results per page, cursor logic) — only the interface changes
- objects-service handler layer for `/objects` and `/object-types` — already using limit/offset ✅
- Other services (user-service, auth-service) — no list endpoints today
- MCP tool surface area or mcp-server client code

## Scope

**objects-service only.** 4 affected files in models, 3 in handlers, 1 in service, 2 in repository, 1 interface file. ~50 lines of changes across 8 files (including tests).

## Risks

- **None significant.** This is dev mode — no external consumers depend on the current param format. Relationship-types endpoint currently has zero clients beyond manual testing.
