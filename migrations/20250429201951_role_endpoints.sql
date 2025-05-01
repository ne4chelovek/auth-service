-- +goose Up
-- +goose StatementBegin
CREATE TABLE role_endpoints (
    role TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    PRIMARY KEY (role, endpoint)
);

INSERT INTO role_endpoints (role, endpoint) VALUES
    ('admin', '/chat/v1/create'),
    ('admin', '/chat/v1'),
    ('admin', '/chat/v1/delete'),
    ('admin', '/chat/v1/send'),
    ('user', '/chat/v1/create'),
    ('user', '/chat/v1/send');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS role_endpoints;
-- +goose StatementEnd
