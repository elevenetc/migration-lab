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
    ALTER TABLE users ADD COLUMN last_name VARCHAR(100);
    ALTER TABLE users ADD COLUMN phone VARCHAR(20);
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
    ALTER TABLE products ADD COLUMN sku VARCHAR(100) UNIQUE;
    ALTER TABLE products ADD COLUMN category_id INTEGER REFERENCES categories(id);
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
    ALTER TABLE orders ADD COLUMN shipped_at TIMESTAMP;
    ALTER TABLE orders ADD COLUMN delivered_at TIMESTAMP;
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
    ALTER TABLE payments ADD COLUMN error_message TEXT;
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
            "V16__create_audit_log" to createAuditLog
        )
    )
}
