CREATE TABLE merchants (
	id UUID PRIMARY KEY,
	name VARCHAR(32) NOT NULL,
	phone VARCHAR(16) NOT NULL,
	pin VARCHAR(255) NOT NULL,
	timezone VARCHAR(64) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO merchants (id, name, phone, pin, timezone) VALUES ('14045670-dd76-416d-b798-436757cef4b6', 'Alfan', '628123', '$2a$10$AVpkJDbaVra8swA5Mm0q/eibAJe/zm1jeRcNCP5nuyI7Hi3r1kd1m', 'Asia/Jakarta');