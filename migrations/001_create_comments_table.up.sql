-- Create extension for UUID generation if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create extension for trigram matching (required for gist_trgm_ops)
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create comments table with hierarchical support
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    root_id VARCHAR(255) NOT NULL,  -- External entity ID (post, product, etc.)
    parent_id UUID REFERENCES comments(id),  -- Self-referencing for hierarchy
    user_id VARCHAR(255) NOT NULL,  -- External user ID
    content TEXT NOT NULL CHECK (length(content) > 0 AND length(content) <= 10000),
    media_url TEXT,
    link_url TEXT,
    upvotes BIGINT DEFAULT 0 CHECK (upvotes >= 0),
    downvotes BIGINT DEFAULT 0 CHECK (downvotes >= 0),
    score BIGINT DEFAULT 0,  -- Calculated field: upvotes - downvotes
    depth INTEGER DEFAULT 0 CHECK (depth >= 0),
    path TEXT NOT NULL,  -- Materialized path for efficient tree queries (e.g., "1.2.3")
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create votes table
CREATE TABLE votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    vote_type SMALLINT NOT NULL CHECK (vote_type IN (-1, 1)),  -- -1 for downvote, 1 for upvote
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure one vote per user per comment
    UNIQUE(comment_id, user_id)
);

-- Indexes for optimal query performance

-- Primary access patterns
CREATE INDEX idx_comments_root_id ON comments(root_id) WHERE NOT is_deleted;
CREATE INDEX idx_comments_user_id ON comments(user_id) WHERE NOT is_deleted;
CREATE INDEX idx_comments_parent_id ON comments(parent_id) WHERE NOT is_deleted;

-- Hierarchical queries
CREATE INDEX idx_comments_path ON comments USING GIST(path gist_trgm_ops) WHERE NOT is_deleted;
CREATE INDEX idx_comments_root_depth ON comments(root_id, depth) WHERE NOT is_deleted;
CREATE INDEX idx_comments_parent_depth ON comments(parent_id, depth) WHERE NOT is_deleted;

-- Sorting and pagination
CREATE INDEX idx_comments_root_score_created ON comments(root_id, score DESC, created_at DESC) WHERE NOT is_deleted;
CREATE INDEX idx_comments_root_created ON comments(root_id, created_at DESC) WHERE NOT is_deleted;
CREATE INDEX idx_comments_root_updated ON comments(root_id, updated_at DESC) WHERE NOT is_deleted;

-- Vote queries
CREATE INDEX idx_votes_comment_id ON votes(comment_id);
CREATE INDEX idx_votes_user_id ON votes(user_id);
CREATE INDEX idx_votes_comment_user ON votes(comment_id, user_id);

-- Statistics queries
CREATE INDEX idx_comments_created_at ON comments(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_comments_root_created_stats ON comments(root_id, created_at) WHERE NOT is_deleted;

-- Function to update comment scores after vote changes
CREATE OR REPLACE FUNCTION update_comment_score()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the comment's vote counts and score
    UPDATE comments 
    SET 
        upvotes = (SELECT COUNT(*) FROM votes WHERE comment_id = NEW.comment_id AND vote_type = 1),
        downvotes = (SELECT COUNT(*) FROM votes WHERE comment_id = NEW.comment_id AND vote_type = -1),
        updated_at = NOW()
    WHERE id = NEW.comment_id;
    
    -- Update the calculated score
    UPDATE comments 
    SET score = upvotes - downvotes 
    WHERE id = NEW.comment_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to handle vote deletions
CREATE OR REPLACE FUNCTION update_comment_score_on_delete()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the comment's vote counts and score
    UPDATE comments 
    SET 
        upvotes = (SELECT COUNT(*) FROM votes WHERE comment_id = OLD.comment_id AND vote_type = 1),
        downvotes = (SELECT COUNT(*) FROM votes WHERE comment_id = OLD.comment_id AND vote_type = -1),
        updated_at = NOW()
    WHERE id = OLD.comment_id;
    
    -- Update the calculated score
    UPDATE comments 
    SET score = upvotes - downvotes 
    WHERE id = OLD.comment_id;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Triggers to automatically update scores when votes change
CREATE TRIGGER trigger_update_comment_score_insert
    AFTER INSERT ON votes
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_score();

CREATE TRIGGER trigger_update_comment_score_update
    AFTER UPDATE ON votes
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_score();

CREATE TRIGGER trigger_update_comment_score_delete
    AFTER DELETE ON votes
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_score_on_delete();

-- Function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at on comments
CREATE TRIGGER trigger_comments_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to automatically update updated_at on votes
CREATE TRIGGER trigger_votes_updated_at
    BEFORE UPDATE ON votes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();