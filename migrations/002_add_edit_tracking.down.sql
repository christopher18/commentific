-- Remove the enhanced edit tracking trigger
DROP TRIGGER IF EXISTS trigger_comments_edit_tracking ON comments;

-- Drop the edit tracking function
DROP FUNCTION IF EXISTS update_comment_edit_tracking();

-- Recreate the original simple updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_comments_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Drop the edit tracking indexes
DROP INDEX IF EXISTS idx_comments_edit_count;
DROP INDEX IF EXISTS idx_comments_is_edited;

-- Remove the enhanced edit tracking columns
ALTER TABLE comments DROP COLUMN IF EXISTS original_content;
ALTER TABLE comments DROP COLUMN IF EXISTS edit_count;
ALTER TABLE comments DROP COLUMN IF EXISTS content_updated_at;
ALTER TABLE comments DROP COLUMN IF EXISTS is_edited; 