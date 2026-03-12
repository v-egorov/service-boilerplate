-- Add created_by and updated_by columns to object_types
ALTER TABLE objects_service.object_types 
ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) DEFAULT 'system',
ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255) DEFAULT 'system';

-- Create index for created_by for auditing
CREATE INDEX IF NOT EXISTS idx_object_types_created_by ON objects_service.object_types(created_by);
CREATE INDEX IF NOT EXISTS idx_object_types_updated_by ON objects_service.object_types(updated_by);
