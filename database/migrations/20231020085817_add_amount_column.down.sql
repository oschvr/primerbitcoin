ALTER TABLE orders RENAME TO orders_old;

CREATE TABLE IF NOT EXISTS orders (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  exchange TEXT,
  order_id INTEGER,
  symbol TEXT,
  quantity TEXT,
  price TEXT,
  success INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO orders (id, exchange, order_id, symbol, quantity, price, success, created_at) SELECT id, exchange, order_id, symbol, quantity, price, success, created_at FROM orders_old;