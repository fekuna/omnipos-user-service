CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE, -- system roles can't be deleted
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(merchant_id, name)
);
CREATE INDEX idx_roles_merchant_id ON roles(merchant_id);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL UNIQUE, -- e.g., 'product.create', 'order.refund'
    name VARCHAR(100) NOT NULL,
    description TEXT,
    module VARCHAR(50) NOT NULL, -- product, order, customer, etc.
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role_id UUID REFERENCES roles(id),
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, suspended
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(merchant_id, username),
    UNIQUE(merchant_id, email)
);
CREATE INDEX idx_users_merchant_id ON users(merchant_id);
CREATE INDEX idx_users_status ON users(status);
