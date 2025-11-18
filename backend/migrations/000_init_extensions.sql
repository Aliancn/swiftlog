-- Migration: Initialize PostgreSQL extensions
-- Description: Enable required extensions for UUID generation

-- Enable UUID generation extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Comments
COMMENT ON EXTENSION pgcrypto IS 'Cryptographic functions including gen_random_uuid()';
