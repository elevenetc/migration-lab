package datasets

// complexEcommerceMigrations models a realistic, long-lived ecommerce schema.
// One migration is one product release: it creates the feature's tables with
// their full column set already in place, and alters the existing tables the
// feature plugs into, so a single migration fills several rows of one timeline
// column. The partitioned table (order_events) has its parent created early
// with child partitions added much later, producing a large horizontal gap.
var complexEcommerceMigrations = []migration{
	{"v1", initialAccounts},
	{"v2", catalogFoundation},
	{"v3", ordersFoundation},
	{"v4", paymentsFeature},
	{"v5", addressesFeature},
	{"v6", warehousingFoundation},
	{"v7", cartFeature},
	{"v8", orderEventsPartitionedParent},
	{"v9", reviewsFeature},
	{"v10", shippingFeature},
	{"v11", couponsFeature},
	{"v12", supplierRestocking},
	{"v13", subscriptionsFeature},
	{"v14", notificationsFeature},
	{"v15", wishlistFeature},
	{"v16", auditTrail},
	{"v17", taxCalculation},
	{"v18", giftCardsFeature},
	{"v19", catalogDataQuality},
	{"v20", moneyPrecision},
	{"v21", orderStatusRules},
	{"v22", refundsFeature},
	{"v23", identityCleanup},
	{"v24", cartLifecycle},
	{"v25", activityLogRename},
	{"v26", orderEventsActor},
	{"v27", notificationOutboxFeature},
	{"v28", contactConsolidation},
	{"v29", marketplaceSkus},
	{"v30", orderEventsBackfillPartitions},
	{"v31", orderEventsCurrentPartition},
	{"v32", trackingConsolidation},
}

const initialAccounts = `CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    billing_email VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);`

const catalogFoundation = `CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    parent_id INTEGER REFERENCES categories(id)
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    sku VARCHAR(50) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    weight_grams INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    url VARCHAR(500) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0
);`

const ordersFoundation = `CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    tracking_number VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    price DECIMAL(10, 2) NOT NULL
);`

const paymentsFeature = `CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    amount DECIMAL(10, 2) NOT NULL,
    method VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE orders ADD COLUMN paid_at TIMESTAMP;`

const addressesFeature = `CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    label VARCHAR(50) NOT NULL DEFAULT 'home',
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20)
);

ALTER TABLE users ADD COLUMN organization_id INTEGER REFERENCES organizations(id);`

const warehousingFoundation = `CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);`

const cartFeature = `CREATE TABLE carts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INTEGER NOT NULL REFERENCES carts(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL DEFAULT 1
);`

// The partitioned parent lands early; its child partitions arrive near the end
// of the timeline (v30/v31), which is what draws the long horizontal gap.
const orderEventsPartitionedParent = `CREATE TABLE order_events (
    id SERIAL,
    order_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    payload JSONB,
    occurred_at TIMESTAMP NOT NULL
) PARTITION BY RANGE (occurred_at);`

const reviewsFeature = `CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(255),
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE products ADD COLUMN rating_avg NUMERIC(3, 2), ADD COLUMN rating_count INTEGER NOT NULL DEFAULT 0;`

const shippingFeature = `CREATE TABLE shipping_methods (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    carrier VARCHAR(100) NOT NULL,
    base_cost DECIMAL(10, 2) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE shipments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    shipping_method_id INTEGER NOT NULL REFERENCES shipping_methods(id),
    tracking_code VARCHAR(100),
    shipped_at TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'preparing'
);

ALTER TABLE orders ADD COLUMN shipping_method_id INTEGER REFERENCES shipping_methods(id);
ALTER TABLE orders ADD COLUMN shipping_address_id INTEGER REFERENCES addresses(id);`

const couponsFeature = `CREATE TABLE coupons (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    discount_percent INTEGER NOT NULL CHECK (discount_percent > 0 AND discount_percent <= 100),
    usage_limit INTEGER,
    expires_at TIMESTAMP
);

CREATE TABLE coupon_redemptions (
    id SERIAL PRIMARY KEY,
    coupon_id INTEGER NOT NULL REFERENCES coupons(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    order_id INTEGER NOT NULL REFERENCES orders(id),
    redeemed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE orders ADD COLUMN coupon_id INTEGER REFERENCES coupons(id), ADD COLUMN discount_amount DECIMAL(10, 2) NOT NULL DEFAULT 0;`

const supplierRestocking = `CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255),
    lead_time_days INTEGER NOT NULL DEFAULT 7,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE products ADD COLUMN supplier_id INTEGER REFERENCES suppliers(id);
ALTER TABLE inventory ADD COLUMN reorder_point INTEGER NOT NULL DEFAULT 0;`

const subscriptionsFeature = `CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    price_monthly DECIMAL(10, 2) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    plan_id INTEGER NOT NULL REFERENCES plans(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP
);`

const notificationsFeature = `CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE notification_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    channel VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE
);`

const wishlistFeature = `CREATE TABLE wishlists (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL DEFAULT 'default',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE wishlist_items (
    id SERIAL PRIMARY KEY,
    wishlist_id INTEGER NOT NULL REFERENCES wishlists(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    added_at TIMESTAMP NOT NULL DEFAULT NOW()
);`

const auditTrail = `CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id INTEGER NOT NULL,
    changes JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);`

const taxCalculation = `CREATE TABLE tax_rates (
    id SERIAL PRIMARY KEY,
    country VARCHAR(100) NOT NULL,
    region VARCHAR(100),
    rate NUMERIC(5, 4) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

ALTER TABLE order_items ADD COLUMN tax_rate_id INTEGER REFERENCES tax_rates(id), ADD COLUMN tax_amount DECIMAL(10, 2) NOT NULL DEFAULT 0;`

const giftCardsFeature = `CREATE TABLE gift_cards (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    initial_amount DECIMAL(10, 2) NOT NULL,
    balance DECIMAL(10, 2) NOT NULL,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE gift_card_transactions (
    id SERIAL PRIMARY KEY,
    gift_card_id INTEGER NOT NULL REFERENCES gift_cards(id),
    order_id INTEGER REFERENCES orders(id),
    amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);`

const catalogDataQuality = `ALTER TABLE products ADD CONSTRAINT uk_products_sku UNIQUE (sku);
ALTER TABLE products ALTER COLUMN description SET NOT NULL;`

const moneyPrecision = `ALTER TABLE products ALTER COLUMN price TYPE NUMERIC(12, 4);
ALTER TABLE order_items ALTER COLUMN price TYPE NUMERIC(12, 4);
ALTER TABLE payments ALTER COLUMN amount TYPE NUMERIC(12, 4);`

const orderStatusRules = `ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'new';
ALTER TABLE orders ADD CONSTRAINT chk_orders_status CHECK (status IN ('new', 'pending', 'shipped', 'delivered', 'cancelled'));`

const refundsFeature = `CREATE TABLE refunds (
    id SERIAL PRIMARY KEY,
    payment_id INTEGER NOT NULL REFERENCES payments(id),
    amount NUMERIC(12, 4) NOT NULL,
    reason VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'requested',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE payments ADD COLUMN refunded_amount NUMERIC(12, 4) NOT NULL DEFAULT 0, ADD COLUMN transaction_id VARCHAR(255);
ALTER TABLE payments ALTER COLUMN status DROP DEFAULT;`

const identityCleanup = `ALTER TABLE users RENAME COLUMN email TO email_address;
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP;`

const cartLifecycle = `ALTER TABLE carts ADD COLUMN expires_at TIMESTAMP;
ALTER TABLE cart_items ADD COLUMN added_at TIMESTAMP NOT NULL DEFAULT NOW();`

const activityLogRename = `ALTER TABLE audit_log RENAME TO activity_log;
ALTER TABLE activity_log ADD COLUMN request_id VARCHAR(64);`

// ALTER on the partitioned parent table -> AccessExclusiveLock warning.
const orderEventsActor = `ALTER TABLE order_events ADD COLUMN actor_id INTEGER;`

const notificationOutboxFeature = `CREATE TABLE notification_outbox (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    channel VARCHAR(50) NOT NULL,
    template VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

DROP TABLE notifications;`

const contactConsolidation = `ALTER TABLE addresses ADD COLUMN phone VARCHAR(20);
ALTER TABLE users DROP COLUMN phone;`

const marketplaceSkus = `ALTER TABLE products DROP CONSTRAINT uk_products_sku;
ALTER TABLE products ALTER COLUMN description DROP NOT NULL;`

// Child partitions added long after the parent (v8) -> large horizontal gap.
const orderEventsBackfillPartitions = `CREATE TABLE order_events_2023 PARTITION OF order_events
    FOR VALUES FROM ('2023-01-01') TO ('2024-01-01');

CREATE TABLE order_events_2024 PARTITION OF order_events
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');`

const orderEventsCurrentPartition = `CREATE TABLE order_events_2025 PARTITION OF order_events
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

ALTER TABLE order_events ADD COLUMN channel VARCHAR(50);`

const trackingConsolidation = `ALTER TABLE shipments RENAME COLUMN tracking_code TO tracking_number;
ALTER TABLE orders DROP COLUMN tracking_number;`
