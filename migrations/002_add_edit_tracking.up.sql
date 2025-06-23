-- Add enhanced edit tracking fields to comments table
ALTER TABLE comments ADD COLUMN is_edited BOOLEAN DEFAULT FALSE;
ALTER TABLE comments ADD COLUMN content_updated_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE comments ADD COLUMN edit_count INTEGER DEFAULT 0;
ALTER TABLE comments ADD COLUMN original_content TEXT; -- Store original content for reference

-- Create indexes for querying edited comments and edit analytics
CREATE INDEX idx_comments_is_edited ON comments(is_edited) WHERE is_edited = TRUE;
CREATE INDEX idx_comments_edit_count ON comments(edit_count) WHERE edit_count > 0;

-- Update trigger to handle enhanced content tracking
-- First, drop the existing trigger that updates updated_at for ALL changes
DROP TRIGGER IF EXISTS trigger_comments_updated_at ON comments;

-- Create an enhanced function that tracks content changes with edit count and original content
CREATE OR REPLACE FUNCTION update_comment_edit_tracking()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if content, media_url, or link_url changed
    IF (OLD.content IS DISTINCT FROM NEW.content) OR 
       (OLD.media_url IS DISTINCT FROM NEW.media_url) OR 
       (OLD.link_url IS DISTINCT FROM NEW.link_url) THEN
        
        -- Store original content if this is the first edit
        IF OLD.is_edited = FALSE THEN
            NEW.original_content = OLD.content;
        END IF;
        
        -- Update edit tracking fields
        NEW.is_edited = TRUE;
        NEW.content_updated_at = NOW();
        NEW.edit_count = OLD.edit_count + 1;
    END IF;
    
    -- Always update the general updated_at timestamp
    NEW.updated_at = NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for enhanced edit tracking
CREATE TRIGGER trigger_comments_edit_tracking
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_edit_tracking(); 