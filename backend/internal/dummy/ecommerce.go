package dummy

// complexEcommerceMigrations models a realistic, long-lived ecommerce schema.
// CREATE and ALTER statements are interleaved across tables (not batched per table),
// and the partitioned table (order_events) has its parent created early with child
// partitions added much later, producing a large horizontal gap on the timeline.
var complexEcommerceMigrations = []migration{
	{"v1", createUsers},
	{"v2", createOrganizations},
	{"v3", createCategories},
	{"v4", createProducts},
	{"v5", addUserProfile},
	{"v6", createOrders},
	{"v7", addProductDetails},
	{"v8", createOrderEventsPartitioned},
	{"v9", createOrderItems},
	{"v10", addUserOrganization},
	{"v11", createAddresses},
	{"v12", createPayments},
	{"v13", addOrderTracking},
	{"v14", createInventory},
	{"v15", createCarts},
	{"v16", addProductWeight},
	{"v17", createCartItems},
	{"v18", createReviews},
	{"v19", setProductDescriptionNotNull},
	{"v20", createShippingMethods},
	{"v21", addOrderShippingMethod},
	{"v22", createShipments},
	{"v23", createSubscriptions},
	{"v24", createNotifications},
	{"v25", addPaymentDetails},
	{"v26", createCoupons},
	{"v27", createWishlists},
	{"v28", changeProductPrice},
	{"v29", createSuppliers},
	{"v30", addProductSupplier},
	{"v31", createWarehouses},
	{"v32", addInventoryWarehouse},
	{"v33", createAuditLog},
	{"v34", setOrderStatusDefault},
	{"v35", addConstraints},
	{"v36", createRefunds},
	{"v37", renameUserEmail},
	{"v38", addCouponUsageLimit},
	{"v39", renameAuditLog},
	{"v40", addOrderEventActor},
	{"v41", createProductImages},
	{"v42", dropUserPhone},
	{"v43", addCartExpiry},
	{"v44", createOrderEvents2023},
	{"v45", createOrderEvents2024},
	{"v46", createOrderEvents2025},
	{"v47", dropSkuConstraint},
	{"v48", dropDescriptionNotNull},
	{"v49", dropOrderStatusDefault},
	{"v50", dropNotifications},
	{"v51", renameShipmentTracking},
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

const createCategories = `CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    parent_id INTEGER REFERENCES categories(id)
);`

const createProducts = `CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);`

const addUserProfile = `ALTER TABLE users ADD COLUMN first_name VARCHAR(100);
ALTER TABLE users ADD COLUMN last_name VARCHAR(100);
ALTER TABLE users ADD COLUMN phone VARCHAR(20);`

const createOrders = `CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);`

const addProductDetails = `ALTER TABLE products ADD COLUMN description TEXT;
ALTER TABLE products ADD COLUMN sku VARCHAR(50);`

const createOrderEventsPartitioned = `CREATE TABLE order_events (
    id SERIAL,
    order_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    occurred_at TIMESTAMP NOT NULL
) PARTITION BY RANGE (occurred_at);`

const createOrderItems = `CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);`

const addUserOrganization = `ALTER TABLE users ADD COLUMN organization_id INTEGER REFERENCES organizations(id);`

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

const addOrderTracking = `ALTER TABLE orders ADD COLUMN tracking_number VARCHAR(100);`

const createInventory = `CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) UNIQUE,
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW()
);`

const createCarts = `CREATE TABLE carts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    created_at TIMESTAMP DEFAULT NOW()
);`

const addProductWeight = `ALTER TABLE products ADD COLUMN weight_grams INTEGER;`

const createCartItems = `CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INTEGER NOT NULL REFERENCES carts(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL DEFAULT 1
);`

const createReviews = `CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);`

const setProductDescriptionNotNull = `ALTER TABLE products ALTER COLUMN description SET NOT NULL;`

const createShippingMethods = `CREATE TABLE shipping_methods (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    carrier VARCHAR(100) NOT NULL,
    base_cost DECIMAL(10, 2) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);`

const addOrderShippingMethod = `ALTER TABLE orders ADD COLUMN shipping_method_id INTEGER REFERENCES shipping_methods(id);`

const createShipments = `CREATE TABLE shipments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    shipping_method_id INTEGER REFERENCES shipping_methods(id),
    tracking_code VARCHAR(100),
    shipped_at TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'preparing'
);`

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

const addPaymentDetails = `ALTER TABLE payments ADD COLUMN transaction_id VARCHAR(255);`

const createCoupons = `CREATE TABLE coupons (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    discount_percent INTEGER NOT NULL CHECK (discount_percent > 0 AND discount_percent <= 100),
    expires_at TIMESTAMP
);`

const createWishlists = `CREATE TABLE wishlists (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    created_at TIMESTAMP DEFAULT NOW()
);`

const changeProductPrice = `ALTER TABLE products ALTER COLUMN price TYPE NUMERIC(12, 4);`

const createSuppliers = `CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);`

const addProductSupplier = `ALTER TABLE products ADD COLUMN supplier_id INTEGER REFERENCES suppliers(id);`

const createWarehouses = `CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL
);`

const addInventoryWarehouse = `ALTER TABLE inventory ADD COLUMN warehouse_id INTEGER REFERENCES warehouses(id);`

const createAuditLog = `CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id INTEGER NOT NULL,
    changes JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);`

const setOrderStatusDefault = `ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'new';`

const addConstraints = `ALTER TABLE products ADD CONSTRAINT uk_products_sku UNIQUE (sku);
ALTER TABLE orders ADD CONSTRAINT chk_orders_status CHECK (status IN ('new', 'pending', 'shipped', 'delivered', 'cancelled'));
ALTER TABLE reviews ADD CONSTRAINT fk_reviews_products FOREIGN KEY (product_id) REFERENCES products(id);`

const createRefunds = `CREATE TABLE refunds (
    id SERIAL PRIMARY KEY,
    payment_id INTEGER NOT NULL REFERENCES payments(id),
    amount DECIMAL(10, 2) NOT NULL,
    reason VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'requested',
    created_at TIMESTAMP DEFAULT NOW()
);`

const renameUserEmail = `ALTER TABLE users RENAME COLUMN email TO email_address;`

const addCouponUsageLimit = `ALTER TABLE coupons ADD COLUMN usage_limit INTEGER;`

const renameAuditLog = `ALTER TABLE audit_log RENAME TO activity_log;`

// ALTER on the partitioned parent table -> AccessExclusiveLock warning.
const addOrderEventActor = `ALTER TABLE order_events ADD COLUMN actor_id INTEGER;`

const createProductImages = `CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    url VARCHAR(500) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0
);`

const dropUserPhone = `ALTER TABLE users DROP COLUMN phone;`

const addCartExpiry = `ALTER TABLE carts ADD COLUMN expires_at TIMESTAMP;`

// Child partitions added long after the parent (v8) -> large horizontal gap.
const createOrderEvents2023 = `CREATE TABLE order_events_2023 PARTITION OF order_events
    FOR VALUES FROM ('2023-01-01') TO ('2024-01-01');`

const createOrderEvents2024 = `CREATE TABLE order_events_2024 PARTITION OF order_events
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');`

const createOrderEvents2025 = `CREATE TABLE order_events_2025 PARTITION OF order_events
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');`

const dropSkuConstraint = `ALTER TABLE products DROP CONSTRAINT uk_products_sku;`

const dropDescriptionNotNull = `ALTER TABLE products ALTER COLUMN description DROP NOT NULL;`

const dropOrderStatusDefault = `ALTER TABLE orders ALTER COLUMN status DROP DEFAULT;`

const dropNotifications = `DROP TABLE notifications;`

const renameShipmentTracking = `ALTER TABLE shipments RENAME COLUMN tracking_code TO tracking_number;`
