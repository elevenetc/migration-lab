package datasets

var multipleOperationPerMigrationMigrations = []migration{
	{"v1", createUsersAndOrgsMOPM},
	{"v2", alterUsersAndOrgsMOPM},
}

const createUsersAndOrgsMOPM = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255)
);

CREATE TABLE orgs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255)
);`

const alterUsersAndOrgsMOPM = `ALTER TABLE users RENAME COLUMN name TO first_name;
ALTER TABLE users ADD COLUMN last_name VARCHAR(255);
ALTER TABLE users ADD COLUMN address VARCHAR(255);
ALTER TABLE orgs ADD COLUMN address VARCHAR(255);`
