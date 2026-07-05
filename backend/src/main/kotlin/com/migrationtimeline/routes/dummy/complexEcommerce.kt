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
            "V1__create_users" to createUsers,
            "V2__create_organizations" to createOrganizations,
            "V3__create_products_and_categories" to "$createProducts\n$createCategories",
            "V4__create_orders" to createOrders,
            "V5__create_order_items" to createOrderItems,
            "V6__add_user_profile" to addUserProfile,
            "V7__create_addresses" to createAddresses,
            "V8__create_payments" to createPayments,
            "V9__add_product_details" to addProductDetails,
            "V10__create_inventory" to createInventory,
            "V11__create_reviews" to createReviews,
            "V12__add_order_tracking" to addOrderTracking,
            "V13__create_subscriptions_and_notifications" to "$createSubscriptions\n$createNotifications",
            "V14__add_user_organization" to addUserOrganization,
            "V15__add_payment_details" to addPaymentDetails,
            "V16__create_audit_log" to createAuditLog,
            "V17__drop_user_phone" to dropUserPhone,
            "V18__change_product_price_precision" to changeProductPrice,
            "V19__set_description_not_null" to setDescriptionNotNull,
            "V20__set_order_status_default" to setOrderStatusDefault,
            "V21__add_constraints" to addConstraints,
            "V22__rename_user_email" to renameUserEmail,
            "V23__rename_audit_log" to renameAuditLog,
            "V24__drop_sku_constraint" to dropSkuConstraint,
            "V25__drop_description_not_null" to dropDescriptionNotNull,
            "V26__drop_order_status_default" to dropOrderStatusDefault,
            "V27__drop_notifications" to dropNotifications,
            "V28__create_orders_history_partitioned" to createOrdersPartitioned,
            "V29__create_orders_history_partitions" to "$createOrdersHistory2024\n$createOrdersHistory2025"
        )
    )
}
