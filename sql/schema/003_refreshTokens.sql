-- +goose UP
create table refresh_tokens (
    token text primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    expires_at timestamp not null,
    revoked_at timestamp,
    user_id UUID not null,
    foreign key( user_id) references users(id) on delete cascade
);

-- +goose Down
Drop table refresh_tokens;