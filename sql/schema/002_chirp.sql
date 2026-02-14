-- +goose Up
create table chirp(
    id UUID primary key,
    created_at Timestamp not null,
    updated_at Timestamp not null,
    user_id UUID not null ,
    body text not null,
    foreign key(user_id) references users(id) on delete cascade
);

-- +goose Down
Drop table chirp;