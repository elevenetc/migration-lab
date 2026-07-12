package datasets

// tallMigrations stresses the vertical layout: ~60 tables created in only
// 5 migrations, so the timeline is much taller than it is wide.
var tallMigrations = []migration{
	{"v1", tallCoreEntities},
	{"v2", tallProductDomain},
	{"v3", tallSalesDomain},
	{"v4", tallPurchasingDomain},
	{"v5", tallAccessDomain},
}

const tallCoreEntities = `CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
CREATE TABLE vendors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL
);
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE offices (
    id SERIAL PRIMARY KEY,
    city VARCHAR(100) NOT NULL
);
CREATE TABLE currencies (
    code CHAR(3) PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE countries (
    code CHAR(2) PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id),
    email VARCHAR(255)
);
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE leads (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
CREATE TABLE campaigns (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    subject VARCHAR(255) NOT NULL
);
CREATE TABLE notes (
    id SERIAL PRIMARY KEY,
    body TEXT NOT NULL
);
CREATE TABLE attachments (
    id SERIAL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL
);`

const tallProductDomain = `CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
CREATE TABLE item_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE item_variants (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES items(id)
);
CREATE TABLE price_lists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE stock_levels (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES items(id),
    quantity INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE stock_locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE units_of_measure (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);
CREATE TABLE brands (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE item_images (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES items(id),
    url VARCHAR(500) NOT NULL
);
CREATE TABLE item_attributes (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES items(id),
    name VARCHAR(100) NOT NULL
);
CREATE TABLE price_list_entries (
    id SERIAL PRIMARY KEY,
    price_list_id INTEGER NOT NULL REFERENCES price_lists(id),
    item_id INTEGER NOT NULL REFERENCES items(id)
);
CREATE TABLE stock_movements (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES items(id),
    quantity INTEGER NOT NULL
);
CREATE TABLE reorder_rules (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES items(id),
    min_quantity INTEGER NOT NULL
);
ALTER TABLE customers ADD COLUMN segment VARCHAR(50);`

const tallSalesDomain = `CREATE TABLE sales_orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id)
);
CREATE TABLE sales_order_lines (
    id SERIAL PRIMARY KEY,
    sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id)
);
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id)
);
CREATE TABLE invoice_lines (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER NOT NULL REFERENCES invoices(id)
);
CREATE TABLE credit_notes (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER NOT NULL REFERENCES invoices(id)
);
CREATE TABLE quotes (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id)
);
CREATE TABLE deliveries (
    id SERIAL PRIMARY KEY,
    sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id)
);
CREATE TABLE delivery_lines (
    id SERIAL PRIMARY KEY,
    delivery_id INTEGER NOT NULL REFERENCES deliveries(id)
);
CREATE TABLE returns (
    id SERIAL PRIMARY KEY,
    sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id)
);
CREATE TABLE return_lines (
    id SERIAL PRIMARY KEY,
    return_id INTEGER NOT NULL REFERENCES returns(id)
);
CREATE TABLE sales_targets (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employees(id),
    amount DECIMAL(12, 2) NOT NULL
);
CREATE TABLE discounts (
    id SERIAL PRIMARY KEY,
    percent DECIMAL(5, 2) NOT NULL
);
ALTER TABLE items ADD COLUMN barcode VARCHAR(64);`

const tallPurchasingDomain = `CREATE TABLE purchase_orders (
    id SERIAL PRIMARY KEY,
    vendor_id INTEGER NOT NULL REFERENCES vendors(id)
);
CREATE TABLE purchase_order_lines (
    id SERIAL PRIMARY KEY,
    purchase_order_id INTEGER NOT NULL REFERENCES purchase_orders(id)
);
CREATE TABLE goods_receipts (
    id SERIAL PRIMARY KEY,
    purchase_order_id INTEGER NOT NULL REFERENCES purchase_orders(id)
);
CREATE TABLE vendor_bills (
    id SERIAL PRIMARY KEY,
    vendor_id INTEGER NOT NULL REFERENCES vendors(id)
);
CREATE TABLE payment_runs (
    id SERIAL PRIMARY KEY,
    executed_at TIMESTAMP
);
CREATE TABLE bank_accounts (
    id SERIAL PRIMARY KEY,
    iban VARCHAR(34) NOT NULL
);
CREATE TABLE vendor_contacts (
    id SERIAL PRIMARY KEY,
    vendor_id INTEGER NOT NULL REFERENCES vendors(id),
    email VARCHAR(255)
);
CREATE TABLE vendor_ratings (
    id SERIAL PRIMARY KEY,
    vendor_id INTEGER NOT NULL REFERENCES vendors(id),
    score INTEGER NOT NULL
);
CREATE TABLE purchase_requisitions (
    id SERIAL PRIMARY KEY,
    requested_by INTEGER NOT NULL REFERENCES employees(id)
);
CREATE TABLE requisition_lines (
    id SERIAL PRIMARY KEY,
    requisition_id INTEGER NOT NULL REFERENCES purchase_requisitions(id)
);
CREATE TABLE landed_costs (
    id SERIAL PRIMARY KEY,
    purchase_order_id INTEGER NOT NULL REFERENCES purchase_orders(id),
    amount DECIMAL(12, 2) NOT NULL
);
ALTER TABLE employees ADD COLUMN department_id INTEGER REFERENCES departments(id);`

const tallAccessDomain = `CREATE TABLE tax_rates (
    id SERIAL PRIMARY KEY,
    percent DECIMAL(5, 2) NOT NULL
);
CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    currency_code CHAR(3) NOT NULL REFERENCES currencies(code),
    rate DECIMAL(12, 6) NOT NULL
);
CREATE TABLE user_accounts (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employees(id)
);
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id),
    permission_id INTEGER NOT NULL REFERENCES permissions(id)
);
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_account_id INTEGER NOT NULL REFERENCES user_accounts(id)
);
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    user_account_id INTEGER NOT NULL REFERENCES user_accounts(id)
);
CREATE TABLE login_attempts (
    id SERIAL PRIMARY KEY,
    user_account_id INTEGER REFERENCES user_accounts(id),
    succeeded BOOLEAN NOT NULL
);
CREATE TABLE password_resets (
    id SERIAL PRIMARY KEY,
    user_account_id INTEGER NOT NULL REFERENCES user_accounts(id)
);
ALTER TABLE vendors ADD COLUMN rating INTEGER;
ALTER TABLE sales_orders ADD COLUMN status VARCHAR(50) DEFAULT 'draft';`
