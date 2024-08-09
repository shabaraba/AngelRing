# Angel Ring

家族専用のファイルストレージシステム
今のところwebを使ってやり取りする想定

Cloud→クラウド→雲→空に浮いているもの→自分達だけの上に浮いているもの→天使の輪→Angel Ring

## KANBAN

https://github.com/users/shabaraba/projects/10


## 使用技術

- frontend
  - typescript / react
- backend
  - golang / gin

## 起動
```sh
docker compose up
```

## DB設定
```sh
docker exec -it go_container bash
cd migrations
goose status
goose up
```

