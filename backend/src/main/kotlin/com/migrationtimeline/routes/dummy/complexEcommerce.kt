package com.migrationtimeline.routes.dummy

import com.migrationtimeline.models.MigrationTimelineResponse
import com.migrationtimeline.routes.migrationsToResponse

private val createUsers = """
    CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) NOT NULL UNIQUE,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val createOrganizations = """
    CREATE TABLE organizations (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val createProducts = """
    CREATE TABLE products (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        price DECIMAL(10, 2) NOT NULL,
        created_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val createCategories = """
    CREATE TABLE categories (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        parent_id INTEGER REFERENCES categories(id)
    );
""".trimIndent()

private val createOrders = """
    CREATE TABLE orders (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id),
        status VARCHAR(50) DEFAULT 'pending',
        created_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val createOrderItems = """
    CREATE TABLE order_items (
        id SERIAL PRIMARY KEY,
        order_id INTEGER NOT NULL REFERENCES orders(id),
        product_id INTEGER NOT NULL REFERENCES products(id),
        quantity INTEGER NOT NULL,
        price DECIMAL(10, 2) NOT NULL
    );
""".trimIndent()

private val addUserProfile = """
    ALTER TABLE users ADD COLUMN first_name VARCHAR(100);
""".trimIndent()

private val createAddresses = """
    CREATE TABLE addresses (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id),
        street VARCHAR(255) NOT NULL,
        city VARCHAR(100) NOT NULL,
        country VARCHAR(100) NOT NULL,
        postal_code VARCHAR(20)
    );
""".trimIndent()

private val createPayments = """
    CREATE TABLE payments (
        id SERIAL PRIMARY KEY,
        order_id INTEGER NOT NULL REFERENCES orders(id),
        amount DECIMAL(10, 2) NOT NULL,
        method VARCHAR(50) NOT NULL,
        status VARCHAR(50) DEFAULT 'pending',
        created_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val addProductDetails = """
    ALTER TABLE products ADD COLUMN description TEXT;
""".trimIndent()

private val createInventory = """
    CREATE TABLE inventory (
        id SERIAL PRIMARY KEY,
        product_id INTEGER NOT NULL REFERENCES products(id) UNIQUE,
        quantity INTEGER NOT NULL DEFAULT 0,
        reserved INTEGER NOT NULL DEFAULT 0,
        updated_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val createReviews = """
    CREATE TABLE reviews (
        id SERIAL PRIMARY KEY,
        product_id INTEGER NOT NULL REFERENCES products(id),
        user_id INTEGER NOT NULL REFERENCES users(id),
        rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
        comment TEXT,
        created_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val addOrderTracking = """
    ALTER TABLE orders ADD COLUMN tracking_number VARCHAR(100);
""".trimIndent()

private val createSubscriptions = """
    CREATE TABLE subscriptions (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id),
        plan VARCHAR(50) NOT NULL,
        status VARCHAR(50) DEFAULT 'active',
        started_at TIMESTAMP DEFAULT NOW(),
        expires_at TIMESTAMP
    );
""".trimIndent()

private val createNotifications = """
    CREATE TABLE notifications (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id),
        type VARCHAR(50) NOT NULL,
        message TEXT NOT NULL,
        read BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val addUserOrganization = """
    ALTER TABLE users ADD COLUMN organization_id INTEGER REFERENCES organizations(id);
""".trimIndent()

private val addPaymentDetails = """
    ALTER TABLE payments ADD COLUMN transaction_id VARCHAR(255);
""".trimIndent()

private val createAuditLog = """
    CREATE TABLE audit_log (
        id SERIAL PRIMARY KEY,
        user_id INTEGER REFERENCES users(id),
        action VARCHAR(100) NOT NULL,
        entity_type VARCHAR(100) NOT NULL,
        entity_id INTEGER NOT NULL,
        changes JSONB,
        created_at TIMESTAMP DEFAULT NOW()
    );
""".trimIndent()

private val dropUserPhone = """
    ALTER TABLE users DROP COLUMN phone;
""".trimIndent()

// ALTER COLUMN TYPE
private val changeProductPrice = """
    ALTER TABLE products ALTER COLUMN price TYPE NUMERIC(12, 4);
""".trimIndent()

// SET NOT NULL
private val setDescriptionNotNull = """
    ALTER TABLE products ALTER COLUMN description SET NOT NULL;
""".trimIndent()

// DROP NOT NULL
private val dropDescriptionNotNull = """
    ALTER TABLE products ALTER COLUMN description DROP NOT NULL;
""".trimIndent()

// SET DEFAULT
private val setOrderStatusDefault = """
    ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'new';
""".trimIndent()

// DROP DEFAULT
private val dropOrderStatusDefault = """
    ALTER TABLE orders ALTER COLUMN status DROP DEFAULT;
""".trimIndent()

// RENAME TABLE
private val renameAuditLog = """
    ALTER TABLE audit_log RENAME TO activity_log;
""".trimIndent()

// RENAME COLUMN
private val renameUserEmail = """
    ALTER TABLE users RENAME COLUMN email TO email_address;
""".trimIndent()

// ADD CONSTRAINT - various types
private val addConstraints = """
    ALTER TABLE products ADD CONSTRAINT uk_products_sku UNIQUE (sku);
    ALTER TABLE orders ADD CONSTRAINT chk_orders_status CHECK (status IN ('new', 'pending', 'shipped', 'delivered', 'cancelled'));
    ALTER TABLE reviews ADD CONSTRAINT fk_reviews_products FOREIGN KEY (product_id) REFERENCES products(id);
""".trimIndent()

// DROP CONSTRAINT
private val dropSkuConstraint = """
    ALTER TABLE products DROP CONSTRAINT uk_products_sku;
""".trimIndent()

// DROP TABLE
private val dropNotifications = """
    DROP TABLE notifications;
""".trimIndent()

// PARTITION BY - partitioned parent table
private val createOrdersPartitioned = """
    CREATE TABLE orders_history (
        id SERIAL,
        order_id INTEGER NOT NULL,
        status VARCHAR(50) NOT NULL,
        changed_at TIMESTAMP NOT NULL
    ) PARTITION BY RANGE (changed_at);
""".trimIndent()

// PARTITION OF - child partition
private val createOrdersHistory2024 = """
    CREATE TABLE orders_history_2024 PARTITION OF orders_history
        FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
""".trimIndent()

private val createOrdersHistory2025 = """
    CREATE TABLE orders_history_2025 PARTITION OF orders_history
        FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
""".trimIndent()

fun complexEcommerce(): MigrationTimelineResponse {
    return migrationsToResponse(
        mapOf(
            "v1" to createUsers,
            "v2" to createOrganizations,
            "v3" to "$createProducts\n$createCategories",
            "v4" to createOrders,
            "v5" to createOrderItems,
            "v6" to addUserProfile,
            "v7" to createAddresses,
            "v8" to createPayments,
            "v9" to addProductDetails,
            "v10" to createInventory,
            "v11" to createReviews,
            "v12" to addOrderTracking,
            "v13" to "$createSubscriptions\n$createNotifications",
            "v14" to addUserOrganization,
            "v15" to addPaymentDetails,
            "v16" to createAuditLog,
            "v17" to dropUserPhone,
            "v18" to changeProductPrice,
            "v19" to setDescriptionNotNull,
            "v20" to setOrderStatusDefault,
            "v21" to addConstraints,
            "v22" to renameUserEmail,
            "v23" to renameAuditLog,
            "v24" to dropSkuConstraint,
            "v25" to dropDescriptionNotNull,
            "v26" to dropOrderStatusDefault,
            "v27" to dropNotifications,
            "v28" to createOrdersPartitioned,
            "v29" to "$createOrdersHistory2024\n$createOrdersHistory2025"
        )
    )
}
