
-- +goose Up
Create table users (
    id UUID primary key,
    created_at TimeStamp not null,
    updated_at TimeStamp not null,
    email text not null unique,
    password text not null
);

Alter table users add column is_chirpy_red boolean not null ;

-- +goose Down
Drop table users;
