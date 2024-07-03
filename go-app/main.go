package main

import (
  "log"
  "bytes"
  "fmt"
  "net/http"
  // "os"
  "path/filepath"
  "image"
  "image/jpeg"
  // "sync"
  // "mime/multipart"

  "github.com/gin-gonic/gin"
  "github.com/nfnt/resize"
  // "github.com/disintegration/imaging"
)

func uploadImage(c *gin.Context) {
    err := c.Request.ParseMultipartForm(10 << 20) // 10 MBのメモリを使用
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
        return
    }

    // var wg sync.WaitGroup

      // アップロードされたファイルを処理
    form, err := c.MultipartForm()
    if err != nil {
        c.String(http.StatusBadRequest, "get form err: %s", err.Error())
        return
    }
    // file, err := c.FormFile("files")
    files := form.File["files"]

    log.Println(files)
    log.Println("aaa")
    for _, file := range files {
        // 画像をデコード
        img, _, err := image.Decode(file)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        thumbnail := resize.Resize(100, 100, img, resize.Lanczos)
        // サムネイルをJPEGとしてエンコード
        var buf bytes.Buffer
        if err := jpeg.Encode(&buf, thumbnail, nil); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
        saveThumbnailPath := filepath.Join("static/thumbnails", fileName)
        log.Println(saveThumbnailPath)
        err := c.SaveUploadedFile(&buf, saveThumbnailPath)
        if err != nil {
            c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
            return
        }
        savePath := filepath.Join("static/images", file.Filename)
        log.Println(savePath)

        err := c.SaveUploadedFile(file, savePath)
        if err != nil {
            c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
            return
        }
    }

  c.JSON(http.StatusOK, gin.H{"message": "Upload complete"})
}

// func createThumbnail(imagePath string) string {
//   // 画像ファイルを開く
//   file, err := os.Open(imagePath)
//   if err != nil {
//     log.Printf("Failed to open image file: %s", err.Error())
//     return ""
//   }
//   defer file.Close()
//
//   // 画像をデコードしてイメージオブジェクトを取得
//   img, _, err := Imaging.Decode(file)
//   if err != nil {
//     log.Printf("Failed to decode image: %s", err.Error())
//     return ""
//   }
//
//   // サムネイルを作成
//   thumbnail := imaging.Resize(img, 100, 0, imaging.Lanczos)
//
//   // サムネイルを保存
//   thumbnailPath := filepath.Join("./static/images/thumbnails", filepath.Base(imagePath))
//   err = imaging.Save(thumbnail, thumbnailPath)
//   if err != nil {
//     log.Printf("Failed to save thumbnail: %s", err.Error())
//     return ""
//   }
//
//   return thumbnailPath
// }

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

