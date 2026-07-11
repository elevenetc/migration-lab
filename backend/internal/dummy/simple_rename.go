package dummy

var simpleRenameMigrations = []migration{
	{"v1", createUsersSimple},
	{"v2", renameUsersToClients},
}

const createUsersSimple = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE
);`

const renameUsersToClients = `ALTER TABLE users RENAME TO clients;`
