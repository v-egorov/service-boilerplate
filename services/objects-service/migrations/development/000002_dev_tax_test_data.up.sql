-- Development test data for objects_service taxonomy system
-- This migration creates sample object types and objects for testing

-- Insert root object types
INSERT INTO objects_service.object_types (name, description, metadata) VALUES
('Category', 'Root category type for organizing content', '{"icon": "folder", "color": "#FF6B6B"}'),
('Product', 'Root product type for e-commerce items', '{"icon": "package", "color": "#4ECDC4"}'),
('Article', 'Root article type for content management', '{"icon": "document", "color": "#45B7D1"}'),
('Location', 'Root location type for geographical data', '{"icon": "map-pin", "color": "#96CEB4"}');

-- Insert child object types
INSERT INTO objects_service.object_types (name, parent_type_id, description, metadata) VALUES
('Electronics', 2, 'Electronic products and devices', '{"icon": "laptop", "requires_warranty": true}'),
('Clothing', 2, 'Clothing and apparel items', '{"icon": "shirt", "has_sizes": true}'),
('Books', 2, 'Books and publications', '{"icon": "book", "has_isbn": true}'),
('News Article', 3, 'News and journalistic content', '{"icon": "newspaper", "requires_fact_check": true}'),
('Blog Post', 3, 'Blog and opinion content', '{"icon": "pen-tool", "allows_comments": true}'),
('Tutorial', 3, 'Educational and how-to content', '{"icon": "graduation-cap", "has_difficulty": true}'),
('Country', 4, 'Country-level geographical data', '{"icon": "globe", "has_iso_code": true}'),
('City', 4, 'City-level geographical data', '{"icon": "building", "has_population": true}');

-- Insert sample objects for Categories
INSERT INTO objects_service.objects (object_type_id, name, description, metadata, tags, status) VALUES
(1, 'Technology', 'Technology related content and discussions', '{"priority": "high", "moderated": true}', ARRAY['tech', 'innovation'], 'active'),
(1, 'Science', 'Scientific content and research', '{"priority": "medium", "peer_reviewed": true}', ARRAY['science', 'research'], 'active'),
(1, 'Business', 'Business and entrepreneurship content', '{"priority": "medium", "verified": true}', ARRAY['business', 'startup'], 'active');

-- Insert sample objects for Products
INSERT INTO objects_service.objects (object_type_id, name, description, metadata, tags, status) VALUES
(2, 'Laptop Pro', 'High-performance laptop for professionals', '{"price": 1299.99, "brand": "TechCorp", "model": "LP-2024"}', ARRAY['laptop', 'professional', 'tech'], 'active'),
(5, 'Smartphone X', 'Latest smartphone with advanced features', '{"price": 899.99, "brand": "MobileTech", "screen_size": "6.5"}', ARRAY['smartphone', 'mobile', 'tech'], 'active'),
(6, 'Cotton T-Shirt', 'Comfortable cotton t-shirt in various colors', '{"price": 19.99, "brand": "EcoWear", "material": "cotton"}', ARRAY['clothing', 'casual', 'eco'], 'active'),
(7, 'Programming Guide', 'Comprehensive guide to modern programming', '{"price": 45.99, "author": "John Doe", "pages": 450}', ARRAY['book', 'programming', 'education'], 'active');

-- Insert sample objects for Articles
INSERT INTO objects_service.objects (object_type_id, name, description, metadata, tags, status) VALUES
(3, 'Breaking: AI Advances', 'Latest breakthrough in artificial intelligence research', '{"author": "Jane Smith", "word_count": 1200, "published_date": "2024-01-15"}', ARRAY['ai', 'research', 'breaking'], 'active'),
(9, 'Tech Trends 2024', 'Analysis of technology trends for the coming year', '{"author": "Mike Johnson", "word_count": 2500, "allows_comments": true}', ARRAY['tech', 'trends', 'analysis'], 'active'),
(10, 'Getting Started with Go', 'Comprehensive tutorial for Go programming beginners', '{"author": "Sarah Wilson", "word_count": 3000, "difficulty": "beginner"}', ARRAY['go', 'programming', 'tutorial'], 'active');

-- Insert sample objects for Locations
INSERT INTO objects_service.objects (object_type_id, name, description, metadata, tags, status) VALUES
(4, 'United States', 'United States of America', '{"iso_code": "US", "population": 331000000, "continent": "North America"}', ARRAY['country', 'usa', 'north-america'], 'active'),
(11, 'Canada', 'Canada', '{"iso_code": "CA", "population": 38000000, "continent": "North America"}', ARRAY['country', 'canada', 'north-america'], 'active'),
(12, 'New York City', 'New York City, USA', '{"population": 8300000, "country": "US", "timezone": "EST"}', ARRAY['city', 'usa', 'metropolitan'], 'active'),
(12, 'Toronto', 'Toronto, Canada', '{"population": 2800000, "country": "CA", "timezone": "EST"}', ARRAY['city', 'canada', 'metropolitan'], 'active');

-- Insert hierarchical objects (child objects)
INSERT INTO objects_service.objects (object_type_id, parent_object_id, name, description, metadata, tags, status) VALUES
(1, 1, 'Web Development', 'Web development technologies and frameworks', '{"priority": "high", "sub_category": true}', ARRAY['web', 'dev', 'frontend', 'backend'], 'active'),
(1, 1, 'Machine Learning', 'Machine learning algorithms and applications', '{"priority": "high", "sub_category": true}', ARRAY['ml', 'ai', 'data-science'], 'active'),
(2, 2, 'Gaming Laptop', 'High-performance laptop for gaming', '{"price": 1599.99, "brand": "GameCorp", "gpu": "RTX-4070"}', ARRAY['laptop', 'gaming', 'high-performance'], 'active'),
(3, 9, 'Climate Change Impact', 'How climate change affects technology trends', '{"author": "Jane Smith", "word_count": 1800, "followup": true}', ARRAY['climate', 'tech', 'environment'], 'active');