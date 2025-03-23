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
		balance REAL DEFAULT 0 CHECK (balance >= 0),
		total_withdrawn REAL DEFAULT 0 CHECK (total_withdrawn >= 0),
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		order_number TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW(),
		status TEXT NOT NULL,
		bonus_amount REAL,
		UNIQUE(user_id, order_number)
	);

	CREATE TABLE IF NOT EXISTS withdrawals (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		order_number TEXT NOT NULL,
		amount REAL,
		created_at TIMESTAMP DEFAULT NOW(),
		status BOOLEAN NOT NULL DEFAULT TRUE
	);

	CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
	CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);
	`

	SearchUserQuery = `SELECT COUNT(*) FROM users WHERE login = $1;`
	InsertUserQuery = `
						INSERT INTO users (login, password_hash, created_at) 
						VALUES ($1, $2, $3);
						`
	GetUserPasswordQuery     = `SELECT password_hash FROM users WHERE login = $1;`
	SearchOrderByNumberQuery = `SELECT user_id from orders WHERE order_number = $1;`
	GetUIDByUserLoginQuery   = `SELECT id FROM users WHERE login = $1;`
	InsertNewOrderQuery      = `
							INSERT INTO orders (user_id, order_number, created_at, status, bonus_amount) 
							VALUES ($1, $2, $3, $4, $5);`

	GetOrdersByUID = `
						SELECT order_number, status, bonus_amount, created_at 
						FROM orders 
						WHERE user_id = $1 
						ORDER BY created_at DESC;`

	GetWalletByUID = `SELECT balance, total_withdrawn from wallets WHERE user_id = $1;`

	CreateUserWalletQuery = `INSERT INTO wallets (user_id, balance, total_withdrawn, created_at) VALUES ($1, $2, $3, $4);`
)

//user_id INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//balance NUMERIC(10,2) DEFAULT 0 CHECK (balance >= 0),
//total_withdrawn NUMERIC(10,2) DEFAULT 0 CHECK (total_withdrawn >= 0),
//created_at TIMESTAMP DEFAULT NOW()
