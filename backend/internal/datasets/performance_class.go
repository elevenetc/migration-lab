package datasets

// One migration per performance class, in ascending predicted cost: five
// metadata-only ones — including the two PostgreSQL 11+ fast paths that look
// expensive and are not — then a data-scanning pair, then two real rewrites.
var performanceClassMigrations = []migration{
	{"v1", createAccountsTable},
	{"v2", alterAccountsAddNote},
	{"v3", alterAccountsAddStatusDefault},
	{"v4", alterAccountsAddAgeCheckNotValid},
	{"v5", alterAccountsAddCreatedAtNow},
	{"v6", alterAccountsEmailType},
	{"v7", alterAccountsSetNoteNotNull},
	{"v8", alterAccountsAddEmailUnique},
	{"v9", alterAccountsAddTokenRandom},
	{"v10", alterAccountsAgeType},
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

// now() is STABLE, so PostgreSQL 11+ stores the value in the catalog instead of
// writing it to every row.
const alterAccountsAddCreatedAtNow = `ALTER TABLE accounts ADD COLUMN created_at TIMESTAMP DEFAULT now();`

// varchar(255) to text is binary-coercible and drops the length limit, so there
// is nothing to verify and nothing to rewrite.
const alterAccountsEmailType = `ALTER TABLE accounts ALTER COLUMN email TYPE TEXT;`

const alterAccountsSetNoteNotNull = `ALTER TABLE accounts ALTER COLUMN note SET NOT NULL;`

const alterAccountsAddEmailUnique = `ALTER TABLE accounts ADD CONSTRAINT accounts_email_unique UNIQUE (email);`

// The cast hides the call from anything matching on the deparsed default, while
// random() stays volatile and rewrites the table.
const alterAccountsAddTokenRandom = `ALTER TABLE accounts ADD COLUMN token INT DEFAULT random()::int;`

const alterAccountsAgeType = `ALTER TABLE accounts ALTER COLUMN age TYPE BIGINT;`
