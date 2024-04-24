package main

import (
  // "fmt"
  // "io"
  "log"
  "net/http"
  // "os"
  // "path/filepath"
  // "sync"

  "github.com/gin-gonic/gin"
  // "github.com/disintegration/imaging"
)

func uploadImage(c *gin.Context) {
  // // マルチパートフォームデータを解析
  // err := c.Request.ParseMultipartForm(10 << 20) // 10 MBのメモリを使用
  //
  // var wg sync.WaitGroup
  //
  // // アップロードされたファイルを処理
  // files := c.Request.MultipartForm.File["files"]
  // for _, fileHeader := range files {
  //   wg.Add(1)
  //   go func(fileHeader *multipart.FileHeader) {
  //     defer wg.Done()
  //
  //     file, err := fileHeader.Open()
  //     if err != nil {
  //       log.Printf("Failed to open file: %s", err.Error())
  //       return
  //     }
  //     defer file.Close()
  //
  //     // アップロード先のファイルパスを生成
  //     savePath := filepath.Join("./static/images", fileHeader.Filename)
  //
  //     // ファイルをサーバーに保存
  //     outputFile, err := os.Create(savePath)
  //     if err != nil {
  //       log.Printf("Failed to create file on server: %s", err.Error())
  //       return
  //     }
  //     defer outputFile.Close()
  //
  //     // アップロードされたファイルを保存先にコピー
  //     _, err = io.Copy(outputFile, file)
  //     if err != nil {
  //       log.Printf("Failed to save file on server: %s", err.Error())
  //       return
  //     }
  //
  //     // サムネイルを作成
  //     thumbnailPath := createThumbnail(savePath)
  //     if thumbnailPath == "" {
  //       log.Printf("Failed to create thumbnail")
  //       return
  //     }
  //
  //     log.Printf("Thumbnail created: %s", thumbnailPath)
  //   }(fileHeader)
  // }
  //
  // wg.Wait()

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

