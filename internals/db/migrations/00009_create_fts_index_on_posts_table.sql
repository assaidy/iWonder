-- +goose Up
-- +goose StatementBegin
create index on posts using gin (to_tsvector('english', title || ' ' || content));
-- +goose StatementEnd
