package datasets

// wideMigrations stresses the horizontal layout: only 5 tables evolved over
// ~30 migrations, so the timeline is much wider than it is tall.
var wideMigrations = []migration{
	{"v1", wideCreateAuthors},
	{"v2", wideCreateArticles},
	{"v3", wideCreateComments},
	{"v4", wideCreateTags},
	{"v5", wideCreateMedia},
	{"v6", `ALTER TABLE authors ADD COLUMN bio TEXT;`},
	{"v7", `ALTER TABLE articles ADD COLUMN published_at TIMESTAMP;`},
	{"v8", `ALTER TABLE comments ADD COLUMN approved BOOLEAN DEFAULT FALSE;`},
	{"v9", `ALTER TABLE articles ADD COLUMN slug VARCHAR(255);`},
	{"v10", `ALTER TABLE tags ADD COLUMN description TEXT;`},
	{"v11", `ALTER TABLE media ADD COLUMN alt_text VARCHAR(255);`},
	{"v12", `ALTER TABLE authors ADD COLUMN avatar_url VARCHAR(500);`},
	{"v13", `ALTER TABLE articles ALTER COLUMN slug SET NOT NULL;`},
	{"v14", `ALTER TABLE comments ADD COLUMN edited_at TIMESTAMP;`},
	{"v15", `ALTER TABLE articles ADD CONSTRAINT uk_articles_slug UNIQUE (slug);`},
	{"v16", `ALTER TABLE media ADD COLUMN size_bytes BIGINT;`},
	{"v17", `ALTER TABLE authors RENAME COLUMN bio TO biography;`},
	{"v18", `ALTER TABLE articles ADD COLUMN view_count INTEGER DEFAULT 0;`},
	{"v19", `ALTER TABLE tags ADD COLUMN color CHAR(7);`},
	{"v20", `ALTER TABLE comments ALTER COLUMN approved SET DEFAULT TRUE;`},
	{"v21", `ALTER TABLE media ALTER COLUMN size_bytes SET NOT NULL;`},
	{"v22", `ALTER TABLE articles ADD COLUMN summary TEXT;`},
	{"v23", `ALTER TABLE authors ADD COLUMN verified BOOLEAN DEFAULT FALSE;`},
	{"v24", `ALTER TABLE comments ADD COLUMN parent_id INTEGER REFERENCES comments(id);`},
	{"v25", `ALTER TABLE tags RENAME COLUMN description TO summary;`},
	{"v26", `ALTER TABLE media ADD COLUMN mime_type VARCHAR(100);`},
	{"v27", `ALTER TABLE articles ALTER COLUMN view_count TYPE BIGINT;`},
	{"v28", `ALTER TABLE authors DROP COLUMN avatar_url;`},
	{"v29", `ALTER TABLE comments ADD CONSTRAINT chk_comments_body CHECK (LENGTH(body) > 0);`},
	{"v30", `ALTER TABLE articles ALTER COLUMN summary SET NOT NULL;`},
}

const wideCreateAuthors = `CREATE TABLE authors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE
);`

const wideCreateArticles = `CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    author_id INTEGER NOT NULL REFERENCES authors(id),
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL
);`

const wideCreateComments = `CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    article_id INTEGER NOT NULL REFERENCES articles(id),
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);`

const wideCreateTags = `CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);`

const wideCreateMedia = `CREATE TABLE media (
    id SERIAL PRIMARY KEY,
    article_id INTEGER NOT NULL REFERENCES articles(id),
    url VARCHAR(500) NOT NULL
);`
