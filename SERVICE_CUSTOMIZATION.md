# Service Customization Guide

This guide provides step-by-step instructions for customizing a newly created service to use meaningful entity names instead of the generic "entity" placeholders.

## Quick Start

1. **Create the service** (uses generic "entity" naming):

   ```bash
   ./scripts/create-service.sh my-service 8085
   ```

2. **Customize entity names** (follow this guide)

3. **Test the customized service**

## Entity Name Customization

### Step 1: Choose Your Names

Decide on your entity names following these conventions:

| Component         | Convention | Example                                           |
| ----------------- | ---------- | ------------------------------------------------- |
| **Entity Name**   | PascalCase | `ResearchObject`, `BlogPost`, `UserProfile`       |
| **Plural Name**   | kebab-case | `research-objects`, `blog-posts`, `user-profiles` |
| **Variable Name** | camelCase  | `researchObject`, `blogPost`, `userProfile`       |

### Step 2: File Renaming

Rename the entity-related files:

```bash
cd services/my-service

# Handlers
mv internal/handlers/entity_handler.go internal/handlers/research_object_handler.go

# Repository
mv internal/repository/entity_repository.go internal/repository/research_object_repository.go

# Services
mv internal/services/entity_service.go internal/services/research_object_service.go

# Models
mv internal/models/entity.go internal/models/research_object.go
mv internal/models/entity_test.go internal/models/research_object_test.go
```

### Step 3: Update Import Paths

Update all import statements in `.go` files:

```bash
# Update import paths
find . -name "*.go" -exec sed -i 's|entity|research_object|g' {} \;

# Update package names in go.mod (if needed)
sed -i 's|entity|research_object|g' go.mod
```

### Step 4: Update Code References

Update all code references from "entity" to your chosen names:

```bash
# Update struct names, function names, variable names
find . -name "*.go" -exec sed -i 's/Entity/ResearchObject/g' {} \;
find . -name "*.go" -exec sed -i 's/entity/researchObject/g' {} \;

# Update plural references
find . -name "*.go" -exec sed -i 's/entities/research-objects/g' {} \;
```

### Step 5: Update Route Paths

Update the API routes in `cmd/main.go`:

```bash
# Change route group
sed -i 's|"/entities"|"/research-objects"|g' cmd/main.go

# Update variable name (if needed)
sed -i 's/entities :=/researchObjects :=/g' cmd/main.go
```

### Step 6: Update Database Schema

Update database migration files and schema references:

```bash
# Update table names in SQL files
find migrations/ -name "*.sql" -exec sed -i 's/entities/research_objects/g' {} \;

# Update schema references
find . -name "*.sql" -exec sed -i 's/idx_entities_/idx_research_objects_/g' {} \;
find . -name "*.sql" -exec sed -i 's/update_entities_/update_research_objects_/g' {} \;
```

### Step 7: Update Configuration and Documentation

```bash
# Update any hardcoded references in config files
find . -name "*.yaml" -o -name "*.toml" -o -name "*.md" | xargs grep -l "entity" | xargs sed -i 's/entity/research_object/g'

# Update README and documentation
sed -i 's/entity/research object/g' README.md
```

## Verification Steps

### 1. Check Compilation

```bash
go mod tidy
go build ./cmd/main.go
```

### 2. Run Tests

```bash
go test ./...
```

### 3. Check API Endpoints

```bash
# Start the service
make run-my-service

# Test endpoints (should now use /research-objects instead of /entities)
curl http://localhost:8085/api/v1/research-objects
```

### 4. Verify Database Schema

```bash
# Check that tables use correct names
# research_objects instead of entities
```

## Common Issues and Solutions

### Issue: Import Path Errors

**Symptom**: `go build` fails with import path errors
**Solution**: Ensure all import statements are updated consistently

### Issue: Variable Name Conflicts

**Symptom**: Go compilation errors about variable names
**Solution**: Ensure variable names follow Go conventions (no hyphens)

### Issue: Database Migration Errors

**Symptom**: Migration fails due to table name conflicts
**Solution**: Drop existing tables before running new migrations

### Issue: Route Conflicts

**Symptom**: API routes don't work
**Solution**: Verify route paths in `cmd/main.go` match your plural name

## Advanced Customization

### Custom Field Names

Update the entity model fields in `internal/models/research_object.go`:

```go
type ResearchObject struct {
    ID          int64     `json:"id" db:"id"`
    Title       string    `json:"title" db:"title"`           // Changed from Name
    Description string    `json:"description" db:"description"`
    Status      string    `json:"status" db:"status"`         // Added custom field
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
```

### Custom Validation Rules

Update validation in `internal/services/research_object_service.go`:

```go
func (s *ResearchObjectService) validateCreateRequest(req CreateResearchObjectRequest) error {
    if req.Title == "" {  // Changed from Name
        return models.NewValidationError("title", "title is required")
    }
    // Add custom validation for Status field
    if req.Status != "" && req.Status != "draft" && req.Status != "published" {
        return models.NewValidationError("status", "status must be 'draft' or 'published'")
    }
    return nil
}
```

### Custom Database Indexes

Update migration files to add custom indexes:

```sql
-- Add custom indexes for research objects
CREATE INDEX IF NOT EXISTS idx_research_objects_status ON SCHEMA_NAME.research_objects(status);
CREATE INDEX IF NOT EXISTS idx_research_objects_title ON SCHEMA_NAME.research_objects(title);
```

## Automation Script

For convenience, you can create a simple automation script:

```bash
#!/bin/bash
# customize-service.sh

SERVICE_DIR=$1
ENTITY_NAME=$2
PLURAL_NAME=$3

cd "$SERVICE_DIR"

# File renaming
mv internal/handlers/entity_handler.go "internal/handlers/${PLURAL_NAME%?}_handler.go"
# ... add other file renames

# Code updates
find . -name "*.go" -exec sed -i "s/Entity/$ENTITY_NAME/g" {} \;
find . -name "*.go" -exec sed -i "s/entity/${PLURAL_NAME//-/_}/g" {} \;
# ... add other replacements

echo "Service customized successfully!"
```

## Best Practices

1. **Test Thoroughly**: Run all tests after customization
2. **Backup First**: Create a backup before major changes
3. **Incremental Changes**: Make changes in small steps and test each one
4. **Consistent Naming**: Use the same naming conventions throughout
5. **Documentation**: Update README and API docs to reflect changes

## Need Help?

If you encounter issues during customization:

1. Check the service logs for error messages
2. Verify all file renames were completed
3. Ensure import paths are consistent
4. Test compilation at each step
5. Refer to the original template for reference

## Examples

### Example 1: Blog Service

```bash
# Create service
./scripts/create-service.sh blog-service 8085

# Customize
cd services/blog-service
# Follow steps above with:
# Entity Name: BlogPost
# Plural Name: blog-posts
# Variable Name: blogPost
```

### Example 2: E-commerce Service

```bash
# Create service
./scripts/create-service.sh product-service 8081

# Customize
cd services/product-service
# Entity Name: Product
# Plural Name: products
# Variable Name: product
```

This customization approach gives you full control over your service naming while maintaining the benefits of the automated service creation process.

