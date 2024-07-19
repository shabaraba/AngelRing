-- +goose Up
-- +goose StatementBegin

-- 既存のfilesテーブルを変更
ALTER TABLE files
ADD COLUMN file_type VARCHAR(128) NOT NULL AFTER title;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- filesテーブルを元の状態に戻す
ALTER TABLE files
DROP COLUMN file_type;

-- +goose StatementEnd

