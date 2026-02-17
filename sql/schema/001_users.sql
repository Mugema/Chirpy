
-- +goose Up
Create table users (
    id UUID primary key,
    created_at TimeStamp not null,
    updated_at TimeStamp not null,
    email text not null unique,
    password text not null
);

-- +goose Down
Drop table users;
