package main

import (
    "database/sql"
    "fmt"
    "image"
    "image/jpeg"
    _ "image/gif"
    _ "image/png"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
    "encoding/json"
    "context"
    "bytes"
    "io"

    // "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
    "github.com/nfnt/resize"
    "github.com/xfrr/goffmpeg/transcoder"
    "gopkg.in/vansante/go-ffprobe.v2"
)


var db *sql.DB

func init() {
    var err error
    db, err = sql.Open("mysql", os.Getenv("DB_CONNECTION"))
    if err != nil {
        log.Fatal(err)
    }
}

func getFileType(filename string) string {
    ext := filepath.Ext(filename)
    switch strings.ToLower(ext) {
    case ".jpg", ".jpeg", ".png", ".gif":
        return "image"
    case ".mp4", ".avi", ".mov":
        return "video"
    default:
        return "unknown"
    }
}

func uploadFile(c *gin.Context) {
    form, err := c.MultipartForm()
    if err != nil {
        c.String(http.StatusBadRequest, "get form err: %s", err.Error())
        return
    }
    files := form.File["files"]

    // var successFileIDs = []uint;
    for _, file := range files {
        fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
        filePath := filepath.Join("static/files", fileName)

        if err := c.SaveUploadedFile(file, filePath); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        fileType := getFileType(file.Filename)
        // var duration int
        // var format string

        if fileType == "video" {
            _, _, err = getVideoMetadata(filePath)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        }

        // ファイル情報をデータベースに保存
        // result, err := db.Exec("INSERT INTO files (title, original_filename, path, file_type, duration, format) VALUES (?, ?, ?, ?, ?, ?)",
        result, err := db.Exec("INSERT INTO files (title, original_filename, path, file_type) VALUES (?, ?, ?, ?)",
            // file.Filename, file.Filename, filePath, fileType, duration, format)
            file.Filename, file.Filename, filePath, fileType)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        fileID, _ := result.LastInsertId()
        // successFileIDs.push(fileID)

        // サムネイルを生成して保存
        thumbnailSizes := []struct{ width, height uint }{
            {100, 100},
            {200, 200},
            {300, 300},
        }

        for _, size := range thumbnailSizes {
            thumbnailPath, err := createThumbnail(fileName, fileType, size.width, size.height)
            if err != nil {
                log.Printf("Failed to create thumbnail: %v", err)
                continue
            }

            if err := c.SaveUploadedFile(file, thumbnailPath); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }


            _, err = db.Exec("INSERT INTO thumbnails (file_id, path, width, height) VALUES (?, ?, ?, ?)",
                fileID, thumbnailPath, size.width, size.height)
            if err != nil {
                log.Printf("Failed to save thumbnail info: %v", err)
            }
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})//, "file_ids": fileID})
}

func createThumbnail(fileName, fileType string, width, height uint) (string, error) {
    filePath := filepath.Join("static/files", fileName)
    path := filepath.Join("static/thumbnails", fileName)
    thumbnailPath := fmt.Sprintf("%s_%dx%d.jpg", strings.TrimSuffix(path, filepath.Ext(path)), width, height)

    if fileType == "image" {
        return createImageThumbnail(filePath, thumbnailPath, width, height)
    } else if fileType == "video" {
        return createVideoThumbnail(filePath, thumbnailPath)
    }

    return "", fmt.Errorf("unsupported file type")
}

func createImageThumbnail(imagePath, thumbnailPath string, width, height uint) (string, error) {
    file, err := os.Open(imagePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    img, _, err := image.Decode(file)
    if err != nil {
        return "", fmt.Errorf("failed to decode image: %w", err)
    }

    fmt.Printf("2")
    thumbnail := resize.Thumbnail(width, height, img, resize.Lanczos3)

    fmt.Printf("3")
    out, err := os.Create(thumbnailPath)
    if err != nil {
        return "", err
    }
    defer out.Close()

    fmt.Printf("4")
    jpeg.Encode(out, thumbnail, nil)

    fmt.Printf("5")
    return thumbnailPath, nil
}

func createVideoThumbnail(videoPath, thumbnailPath string) (string, error) {
    trans := new(transcoder.Transcoder)
    err := trans.Initialize(videoPath, thumbnailPath)
    if err != nil {
        return "", err
    }
    done := trans.Run(false)
    err = <-done // チャンネルからエラーを受け取る
    if err != nil {
        return "", err
    }
    return thumbnailPath, err
}

func getVideoMetadata(videoPath string) (int, string, error) {
    ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancelFn()

    data, err := ffprobe.ProbeURL(ctx,videoPath)

    if err != nil {
        log.Fatalf("Error getting media info: %v", err)
    }

    duration := data.Format.Duration()

    format := data.Format.FormatName

    return int(duration.Seconds()), format, nil
}

func getThumbnails(c *gin.Context) {
    fileID := c.Param("id")
    width := c.Query("width")
    height := c.Query("height")

    query := "SELECT path FROM thumbnails WHERE file_id = ?"
    args := []interface{}{fileID}

    if width != "" && height != "" {
        query += " AND width = ? AND height = ?"
        args = append(args, width, height)
    }

    rows, err := db.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var thumbnails []string
    for rows.Next() {
        var path string
        if err := rows.Scan(&path); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }
        thumbnails = append(thumbnails, path)
    }

    c.JSON(http.StatusOK, gin.H{"thumbnails": thumbnails})
}

func deleteFile(c *gin.Context) {
    fileID := c.Param("id")

    // ファイル情報を取得
    var filePath string
    err := db.QueryRow("SELECT path FROM files WHERE id = ?", fileID).Scan(&filePath)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    // サムネイルのパスを取得
    rows, err := db.Query("SELECT path FROM thumbnails WHERE file_id = ?", fileID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var thumbnailPaths []string
    for rows.Next() {
        var path string
        if err := rows.Scan(&path); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }
        thumbnailPaths = append(thumbnailPaths, path)
    }

    // データベースからファイル情報を削除（thumbnailsは外部キー制約でカスケード削除される）
    _, err = db.Exec("DELETE FROM files WHERE id = ?", fileID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file from database"})
        return
    }

    // 実際のファイルを削除
    if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
        log.Printf("Failed to delete file: %v", err)
    }

    // サムネイルを削除
    for _, thumbnailPath := range thumbnailPaths {
        if err := os.Remove(thumbnailPath); err != nil && !os.IsNotExist(err) {
            log.Printf("Failed to delete thumbnail: %v", err)
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "File and its thumbnails deleted successfully"})
}

type ThumbnailRecord struct {
    ID            uint    `json:"id"`
    FileId        uint    `json:"fileId"`
    Path          string  `json:"path"`
    width         uint    `json:"width"`
    height        uint    `json:"height"`
    CreatedAt     []uint8 `json:"created_at"`
    UpdatedAt     []uint8 `json:"updated_at"`
}
func getImages(c *gin.Context) {
    files, err := db.Query("SELECT * FROM thumbnails WHERE width = 100")
    if err != nil {
        c.String(http.StatusInternalServerError, "database insert err: %s", err.Error())
        return
    }
    // defer rows.Close()

    var results []ThumbnailRecord
    for files.Next() {
        var r ThumbnailRecord
        if err := files.Scan(&r.ID, &r.FileId, &r.Path, &r.width, &r.height, &r.CreatedAt, &r.UpdatedAt); err != nil {
            log.Fatal(err)
        }
        results = append(results, r)
    }
    // 結果をJSONに変換
    jsonData, err := json.Marshal(results)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{ "data": string(jsonData) })
}

type FileRecord struct {
    ID            uint      `json:"id"`
    Title         string    `json:"title"`
    FileType      string    `json:"fileType"`
    FileName      string    `json:"fileName"`
    Path          string    `json:"path"`
    CreatedAt     []uint8 `json:"created_at"`
    UpdatedAt     []uint8 `json:"updated_at"`
}
func getFile(c *gin.Context) {
    fileID := c.Param("id")

    var r FileRecord
    err := db.QueryRow("SELECT * FROM files WHERE id = ?", fileID).
        Scan(&r.ID, &r.Title, &r.FileType, &r.FileName, &r.Path, &r.CreatedAt, &r.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
        }
        return
    }
    // 結果をJSONに変換
    
    c.JSON(http.StatusOK, gin.H{"data": r})
}

func uploadImage(c *gin.Context) {
    err := c.Request.ParseMultipartForm(10 << 20)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
        return
    }

    form, err := c.MultipartForm()
    if err != nil {
        c.String(http.StatusBadRequest, "get form err: %s", err.Error())
        return
    }
    files := form.File["files"]

    for _, file := range files {
        src, err := file.Open()
        if err != nil {
            c.String(http.StatusInternalServerError, "file open err: %s", err.Error())
            return
        }
        defer src.Close()

        img, _, err := image.Decode(src)
        if err != nil {
            fmt.Printf("image decode err: %s", err.Error())
            c.String(http.StatusInternalServerError, "image decode err: %s", err.Error())
            return
        }

        thumbnail := resize.Resize(100, 100, img, resize.Lanczos3)
        var buf bytes.Buffer
        if err := jpeg.Encode(&buf, thumbnail, nil); err != nil {
            c.String(http.StatusInternalServerError, "thumbnail encode err: %s", err.Error())
            return
        }

        fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
        savePath := filepath.Join("static/images", fileName)
        saveThumbnailPath := filepath.Join("static/thumbnails", fileName)

        if err := c.SaveUploadedFile(file, savePath); err != nil {
            c.String(http.StatusInternalServerError, "save file err: %s", err.Error())
            return
        }

        out, err := os.Create(saveThumbnailPath)
        if err != nil {
            c.String(http.StatusInternalServerError, "create thumbnail file err: %s", err.Error())
            return
        }
        defer out.Close()

        _, err = io.Copy(out, &buf)
        if err != nil {
            c.String(http.StatusInternalServerError, "save thumbnail err: %s", err.Error())
            return
        }


        // MySQLにデータを保存
        _, err = db.Exec("INSERT INTO files (title, path, thumbnail_path) VALUES (?, ?, ?)",
            file.Filename, savePath, saveThumbnailPath)
        if err != nil {
            c.String(http.StatusInternalServerError, "database insert err: %s", err.Error())
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Upload complete"})
}

func main() {
    router := gin.Default()
    // config := cors.DefaultConfig()
    // config.AllowOrigins = []string{"http://react_container:5173"} // Viteのデフォルトポート
    // router.Use(cors.New(config))

    router.Static("/static", "./static")
    router.POST("/api/upload", uploadFile)
    router.GET("/api/thumbnails", getImages)
    router.GET("/api/thumbnails/:id", getThumbnails)
    router.DELETE("/api/files/:id", deleteFile)
    router.GET("/api/file/:id", getFile)

    if err := router.Run(":8080"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
