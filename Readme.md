# Blog API - Hệ thống quản lý bài viết

## Giới thiệu

Dự án này là hệ thống API quản lý bài viết sử dụng Go, PostgreSQL, Redis, Elasticsearch. Hỗ trợ các chức năng: tạo bài viết, tìm kiếm, cache, log hoạt động, và tìm bài viết liên quan.

---

## Yêu cầu hệ thống

- Docker & Docker Compose
- Git

---

## Hướng dẫn chạy dự án

1. **Clone repository về máy**
   ```sh
   git clone https://github.com/Huy-Nguyen-Seta/go_bvn
   cd go_bvn
   ```

2. **Khởi động các service bằng Docker Compose**
   ```sh
   docker-compose up --build
   ```
   - Service API sẽ chạy ở cổng `8080`
   - PostgreSQL ở cổng `5432`
   - Redis ở cổng `6379`
   - Elasticsearch ở cổng `9200`

3. **Kiểm tra database đã có dữ liệu mẫu**
   - File `init.sql` sẽ tự động được thực thi nếu bạn cấu hình volume hoặc script cho container PostgreSQL.

---

## Hướng dẫn test các API

### 1. Tạo bài viết mới
```sh
curl -X POST http://localhost:8080/api/v1/posts \
-H "Content-Type: application/json" \
-d '{"title":"Một Con Vịt Remix","content":"Phiên bản remix của bài hát Một Con Vịt...","tags":["remix","một con vịt","âm nhạc"]}'
```

### 2. Lấy chi tiết bài viết theo ID
```sh
curl http://localhost:8080/api/v1/posts/1
```
- Kết quả trả về sẽ có trường `related_posts` (nếu có bài viết liên quan theo tags).

### 3. Tìm kiếm bài viết theo tag
```sh
curl "http://localhost:8080/api/v1/posts/search-by-tag?tag=Tophats"
```

### 4. Tìm kiếm full-text (title, content)
```sh
curl "http://localhost:8080/api/v1/posts/search?q=remix"
```

### 5. Cập nhật bài viết
```sh
curl -X PUT http://localhost:8080/api/v1/posts/1 \
-H "Content-Type: application/json" \
-d '{"title":"Một Con Vịt phiên bản mới","content":"Nội dung mới...","tags":["một con vịt","âm nhạc"]}'
```

---

## Một số lưu ý

- Nếu bạn thay đổi code Go, hãy build lại image API:
  ```sh
  docker-compose build api
  docker-compose up --force-recreate api
  ```
- Nếu cần reset dữ liệu, xóa volume của PostgreSQL và khởi động lại.

---

## Liên hệ