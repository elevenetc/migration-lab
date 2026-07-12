package datasets

var renameWithNewTableInBetweenMigrations = []migration{
	{"v1", createUsersRWNT},
	{"v2", createSettingsRWNT},
	{"v3", renameUsersToClientsRWNT},
}

const createUsersRWNT = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE
);`

const createSettingsRWNT = `CREATE TABLE settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    value TEXT
);`

const renameUsersToClientsRWNT = `ALTER TABLE users RENAME TO clients;`
