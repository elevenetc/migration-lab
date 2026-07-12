package datasets

var simpleTableMigrations = []migration{
	{"v1", createUsersTable},
}

const createUsersTable = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE
);`
