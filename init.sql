-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create database if it doesn't exist (this will be run by docker-entrypoint)
-- The database is already created by POSTGRES_DB environment variable
