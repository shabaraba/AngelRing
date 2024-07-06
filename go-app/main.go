package main

import (
    "bytes"
    "fmt"
    "image"
    "image/jpeg"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/nfnt/resize"
)

func uploadImage(c *gin.Context) {
    err := c.Request.ParseMultipartForm(10 << 20) // 10 MBのメモリを使用
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

        // 画像をデコード
        img, _, err := image.Decode(src)
        if err != nil {
            c.String(http.StatusInternalServerError, "image decode err: %s", err.Error())
            return
        }

        thumbnail := resize.Resize(100, 100, img, resize.Lanczos3)
        // サムネイルをJPEGとしてエンコード
        var buf bytes.Buffer
        if err := jpeg.Encode(&buf, thumbnail, nil); err != nil {
            c.String(http.StatusInternalServerError, "thumbnail encode err: %s", err.Error())
            return
        }

        fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
        saveThumbnailPath := filepath.Join("static/thumbnails", fileName)

        if err := c.SaveUploadedFile(file, filepath.Join("static/images", file.Filename)); err != nil {
            c.String(http.StatusInternalServerError, "save file err: %s", err.Error())
            return
        }

        // サムネイルを保存
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
    }

    c.JSON(http.StatusOK, gin.H{"message": "Upload complete"})
}

func main() {
    router := gin.Default()

    // 静的ファイルの提供
    router.Static("/static", "./static")

    // 画像アップロードエンドポイント
    router.POST("/api/images", uploadImage)

    if err := router.Run(":8080"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
