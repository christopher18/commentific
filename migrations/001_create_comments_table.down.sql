-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_votes_updated_at ON votes;
DROP TRIGGER IF EXISTS trigger_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS trigger_update_comment_score_delete ON votes;
DROP TRIGGER IF EXISTS trigger_update_comment_score_update ON votes;
DROP TRIGGER IF EXISTS trigger_update_comment_score_insert ON votes;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS update_comment_score_on_delete();
DROP FUNCTION IF EXISTS update_comment_score();

-- Drop indexes (they will be dropped automatically with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_comments_root_created_stats;
DROP INDEX IF EXISTS idx_comments_created_at;
DROP INDEX IF EXISTS idx_votes_comment_user;
DROP INDEX IF EXISTS idx_votes_user_id;
DROP INDEX IF EXISTS idx_votes_comment_id;
DROP INDEX IF EXISTS idx_comments_root_updated;
DROP INDEX IF EXISTS idx_comments_root_created;
DROP INDEX IF EXISTS idx_comments_root_score_created;
DROP INDEX IF EXISTS idx_comments_parent_depth;
DROP INDEX IF EXISTS idx_comments_root_depth;
DROP INDEX IF EXISTS idx_comments_path;
DROP INDEX IF EXISTS idx_comments_parent_id;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_comments_root_id;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS comments;

-- Note: We don't drop the uuid-ossp extension as it might be used by other parts of the system 