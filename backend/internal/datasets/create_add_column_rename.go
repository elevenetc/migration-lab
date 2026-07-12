package datasets

var createAddColumnRenameMigrations = []migration{
	{"v1", createUsersCACR},
	{"v2", addNameColumnCACR},
	{"v3", renameUsersToCustomersCACR},
}

const createUsersCACR = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE
);`

const addNameColumnCACR = `ALTER TABLE users ADD COLUMN name VARCHAR(255);`

const renameUsersToCustomersCACR = `ALTER TABLE users RENAME TO customers;`
