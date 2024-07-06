package main

import (
    "bytes"
    "database/sql"
    "encoding/json"
    "fmt"
    "image"
    "image/jpeg"
    _ "image/png"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/nfnt/resize"
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
    var err error
    // MySQLの接続情報を適切に設定してください
    db, err = sql.Open("mysql", os.Getenv("DB_CONNECTION"))
    if err != nil {
        log.Fatal(err)
    }
}

type FileRecord struct {
    ID   int    `json:"id"`
    Title string `json:"title"`
    Path string `json:"path"`
    ThumbnailPath string `json:"thumbnailPath"`
    CreatedAt string `json:"createdAt"`
    UpdatedAt string `json:"updatedAt"`
}

func getImages(c *gin.Context) {
    // MySQLにデータを保存
    files, err := db.Query("SELECT * FROM files")
    if err != nil {
        c.String(http.StatusInternalServerError, "database insert err: %s", err.Error())
        return
    }
    // defer rows.Close()

    var results []FileRecord
    for files.Next() {
        var r FileRecord
        if err := files.Scan(&r.ID, &r.Title, &r.Path, &r.ThumbnailPath, &r.CreatedAt, &r.UpdatedAt); err != nil {
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
    router.Static("/static", "./static")
    router.GET("/api/images", getImages)
    router.POST("/api/images", uploadImage)

    if err := router.Run(":8080"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
