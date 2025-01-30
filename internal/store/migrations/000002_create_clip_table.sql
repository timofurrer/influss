CREATE TABLE IF NOT EXISTS clip (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    title TEXT,
    author TEXT,
    published_at TIMESTAMP WITH TIME ZONE,
    modified_at TIMESTAMP WITH TIME ZONE,
    excerpt TEXT,
    html_content TEXT,
    plain_text_content TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
