package datasets

// One migration per performance class, in ascending cost: the first four are
// metadata-only, then a data-scanning pair, then two table rewrites.
var performanceClassMigrations = []migration{
	{"v1", createAccountsTable},
	{"v2", alterAccountsAddNote},
	{"v3", alterAccountsAddStatusDefault},
	{"v4", alterAccountsAddAgeCheckNotValid},
	{"v5", alterAccountsSetNoteNotNull},
	{"v6", alterAccountsAddEmailUnique},
	{"v7", alterAccountsAddCreatedAtNow},
	{"v8", alterAccountsEmailType},
}

const createAccountsTable = `CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255),
    age INT
);`

const alterAccountsAddNote = `ALTER TABLE accounts ADD COLUMN note TEXT;`

const alterAccountsAddStatusDefault = `ALTER TABLE accounts ADD COLUMN status VARCHAR(50) DEFAULT 'active';`

const alterAccountsAddAgeCheckNotValid = `ALTER TABLE accounts
    ADD CONSTRAINT accounts_age_positive CHECK (age > 0) NOT VALID;`

const alterAccountsSetNoteNotNull = `ALTER TABLE accounts ALTER COLUMN note SET NOT NULL;`

const alterAccountsAddEmailUnique = `ALTER TABLE accounts ADD CONSTRAINT accounts_email_unique UNIQUE (email);`

const alterAccountsAddCreatedAtNow = `ALTER TABLE accounts ADD COLUMN created_at TIMESTAMP DEFAULT now();`

const alterAccountsEmailType = `ALTER TABLE accounts ALTER COLUMN email TYPE TEXT;`
