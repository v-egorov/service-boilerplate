SELECT
    schemaname as schema,
    relname as table,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname)) as size,
    n_tup_ins - n_tup_del as row_count
FROM pg_stat_user_tables
WHERE schemaname = 'user_service'
ORDER BY n_tup_ins - n_tup_del DESC;