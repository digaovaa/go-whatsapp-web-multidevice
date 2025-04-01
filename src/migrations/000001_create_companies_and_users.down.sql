-- Drop indexes
DROP INDEX IF EXISTS idx_companies_email;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_company_id;
DROP INDEX IF EXISTS idx_whatsapp_connections_user_id;
DROP INDEX IF EXISTS idx_api_tokens_token;
DROP INDEX IF EXISTS idx_api_tokens_company_id;
DROP INDEX IF EXISTS idx_api_tokens_user_id;

-- Drop tables
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS whatsapp_connections;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS companies; 