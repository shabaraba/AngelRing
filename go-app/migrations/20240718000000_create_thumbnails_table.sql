-- +goose Up
-- +goose StatementBegin

-- 既存のfilesテーブルを変更
ALTER TABLE files
DROP COLUMN thumbnail_path,
ADD COLUMN original_filename VARCHAR(255) NOT NULL AFTER title;
-- +goose StatementEnd

-- +goose StatementBegin

-- サムネイル用の新しいテーブルを作成
CREATE TABLE thumbnails (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_id INT NOT NULL,
    path VARCHAR(255) NOT NULL,
    width INT NOT NULL,
    height INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin

-- サムネイルのサイズでインデックスを作成
CREATE INDEX idx_thumbnails_size ON thumbnails(width, height);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- サムネイルテーブルを削除
DROP TABLE IF EXISTS thumbnails;
-- +goose StatementEnd

-- +goose StatementBegin
-- filesテーブルを元の状態に戻す
ALTER TABLE files
ADD COLUMN thumbnail_path VARCHAR(255),
DROP COLUMN original_filename;

-- +goose StatementEnd

