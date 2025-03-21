package postrge

const (
	MigrationQuery = `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		login TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS wallets (
		id SERIAL PRIMARY KEY,
		user_id INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		balance NUMERIC(10,2) DEFAULT 0 CHECK (balance >= 0),
		total_withdrawn NUMERIC(10,2) DEFAULT 0 CHECK (total_withdrawn >= 0),
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		order_number TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW(),
		status TEXT NOT NULL,
		bonus_amount NUMERIC(10,2) CHECK (bonus_amount >= 0),
		UNIQUE(user_id, order_number)
	);

	CREATE TABLE IF NOT EXISTS withdrawals (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		order_number TEXT NOT NULL,
		amount NUMERIC(10,2) NOT NULL CHECK (amount >= 0),
		created_at TIMESTAMP DEFAULT NOW(),
		status BOOLEAN NOT NULL DEFAULT TRUE
	);

	CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
	CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);
	`
)
