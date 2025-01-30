CREATE INDEX IF NOT EXISTS idx_clips_url ON clip(url);
CREATE INDEX IF NOT EXISTS idx_clips_published_at ON clip(published_at);
CREATE INDEX IF NOT EXISTS idx_clips_created_at ON clip(created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_clips_url_unique ON clip(url);
