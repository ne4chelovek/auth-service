-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS role_endpoints;
CREATE TABLE IF NOT EXISTS role_endpoints (
role TEXT NOT NULL,
endpoint TEXT NOT NULL,
PRIMARY KEY (role, endpoint)
);

INSERT INTO role_endpoints (role, endpoint) VALUES
('admin', '/chat_v1.Chat/Create'),
('admin', '/chat_v1.Chat/Get'),
('admin', '/chat_v1.Chat/Delete'),
('admin', '/chat_v1.Chat/SendMessage'),
('admin', '/chat_v1.Chat/Connect'),
('user', '/chat_v1.Chat/Get'),
('user', '/chat_v1.Chat/Create'),
('user', '/chat_v1.Chat/SendMessage'),
('user', '/chat_v1.Chat/Delete'),
('user', '/chat_v1.Chat/Connect');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS role_endpoints;
-- +goose StatementEnd