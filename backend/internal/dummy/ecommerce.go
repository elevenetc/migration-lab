package dummy

var complexEcommerceMigrations = []migration{
	{"v1", createUsers},
	{"v2", createOrganizations},
	{"v3", createProducts + "\n" + createCategories},
	{"v4", createOrders},
	{"v5", createOrderItems},
	{"v6", addUserProfile},
	{"v7", createAddresses},
	{"v8", createPayments},
	{"v9", addProductDetails},
	{"v10", createInventory},
	{"v11", createReviews},
	{"v12", addOrderTracking},
	{"v13", createSubscriptions + "\n" + createNotifications},
	{"v14", addUserOrganization},
	{"v15", addPaymentDetails},
	{"v16", createAuditLog},
	{"v17", dropUserPhone},
	{"v18", changeProductPrice},
	{"v19", setDescriptionNotNull},
	{"v20", setOrderStatusDefault},
	{"v21", addConstraints},
	{"v22", renameUserEmail},
	{"v23", renameAuditLog},
	{"v24", dropSkuConstraint},
	{"v25", dropDescriptionNotNull},
	{"v26", dropOrderStatusDefault},
	{"v27", dropNotifications},
	{"v28", createOrdersPartitioned},
	{"v29", createOrdersHistory2024 + "\n" + createOrdersHistory2025},
}

const createUsers = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);`

const createOrganizations = `CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);`

const createProducts = `CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);`

const createCategories = `CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    parent_id INTEGER REFERENCES categories(id)
);`

const createOrders = `CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);`

const createOrderItems = `CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);`

const addUserProfile = `ALTER TABLE users ADD COLUMN first_name VARCHAR(100);
ALTER TABLE users ADD COLUMN phone VARCHAR(20);`

const createAddresses = `CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20)
);`

const createPayments = `CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    amount DECIMAL(10, 2) NOT NULL,
    method VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);`

const addProductDetails = `ALTER TABLE products ADD COLUMN description TEXT;
ALTER TABLE products ADD COLUMN sku VARCHAR(50);`

const createInventory = `CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) UNIQUE,
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW()
);`

const createReviews = `CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);`

const addOrderTracking = `ALTER TABLE orders ADD COLUMN tracking_number VARCHAR(100);`

const createSubscriptions = `CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    plan VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    started_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP
);`

const createNotifications = `CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);`

const addUserOrganization = `ALTER TABLE users ADD COLUMN organization_id INTEGER REFERENCES organizations(id);`

const addPaymentDetails = `ALTER TABLE payments ADD COLUMN transaction_id VARCHAR(255);`

const createAuditLog = `CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id INTEGER NOT NULL,
    changes JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);`

const dropUserPhone = `ALTER TABLE users DROP COLUMN phone;`

const changeProductPrice = `ALTER TABLE products ALTER COLUMN price TYPE NUMERIC(12, 4);`

const setDescriptionNotNull = `ALTER TABLE products ALTER COLUMN description SET NOT NULL;`

const dropDescriptionNotNull = `ALTER TABLE products ALTER COLUMN description DROP NOT NULL;`

const setOrderStatusDefault = `ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'new';`

const dropOrderStatusDefault = `ALTER TABLE orders ALTER COLUMN status DROP DEFAULT;`

const renameAuditLog = `ALTER TABLE audit_log RENAME TO activity_log;`

const renameUserEmail = `ALTER TABLE users RENAME COLUMN email TO email_address;`

const addConstraints = `ALTER TABLE products ADD CONSTRAINT uk_products_sku UNIQUE (sku);
ALTER TABLE orders ADD CONSTRAINT chk_orders_status CHECK (status IN ('new', 'pending', 'shipped', 'delivered', 'cancelled'));
ALTER TABLE reviews ADD CONSTRAINT fk_reviews_products FOREIGN KEY (product_id) REFERENCES products(id);`

const dropSkuConstraint = `ALTER TABLE products DROP CONSTRAINT uk_products_sku;`

const dropNotifications = `DROP TABLE notifications;`

const createOrdersPartitioned = `CREATE TABLE orders_history (
    id SERIAL,
    order_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    changed_at TIMESTAMP NOT NULL
) PARTITION BY RANGE (changed_at);`

const createOrdersHistory2024 = `CREATE TABLE orders_history_2024 PARTITION OF orders_history
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');`

const createOrdersHistory2025 = `CREATE TABLE orders_history_2025 PARTITION OF orders_history
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');`
