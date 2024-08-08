-- +goose Up
-- +goose StatementBegin

-- 既存のfilesテーブルを変更
ALTER TABLE files ADD COLUMN taken_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- filesテーブルを元の状態に戻す
ALTER TABLE files
DROP COLUMN taken_at;

-- +goose StatementEnd

