## Mục lục
1. [Bài toán](#bài-toán)
2. [Cách chạy](#cách-chạy)
3. [Thiết kế & Quyết định kỹ thuật](#thiết-kế--quyết-định-kỹ-thuật)
4. [Trade-offs](#trade-offs)
5. [Challenges & Solutions](#challenges--solutions)
6. [Performance & Scalability](#performance--scalability)
7. [Limitations & Improvements](#limitations--improvements)


## Bài toán
> Em có một 1 URL dài và muốn nó ngắn lại nên em tạo 1 dịch vụ bằng golang giúp độ dài URL ngắn lại, dễ theo dõi và chia sẻ cho người khác
### Why
- URLs thường dài và khó nhớ VD: `https://example.com/products/category/item?utm_source=facebook&utm_campaign=summer2024` (90+ ký tự)
- Ứng dụng giới hạn ký tự: Twitter, SMS
- Khó chia sẻ qua giấy hoặc nói miệng
### Use case
- Chia sẻ link (một vài ứng dụng giới hạn ký tự)
- Theo dõi số lần click vào link rút gọn (analytics cho marketing campaigns)
- Printed materials: QR codes, business cards, posters
- User experience: Dễ nhớ, dễ gõ (go.company.com/docs thay vì URL dài)


## Cách chạy
### Yêu cầu
- Trên máy tính có golang bản 1.24.x
- Database: PostgreSQL
- Có git, docker
### Chạy chương trình
```bash
# 1. Clone repo
git clone https://github.com/giaptai/finan-test-short-url
cd finan-test-short-url

# 2. Cấu hình database 
Vào thư mục chứa code tìm file .env.sample đổi thành .env 
Chỉnh thông số cho phù hợp thông số khai báo trong file docker compose

# 3. Chạy PostgreSQL qua Docker compose
docker-compose up -d

# 4. Chạy ứng dụng
go run main.go
```
Ứng dụng sẽ chạy tại ```http://localhost:8088```
### API Endpoints
| Method | Endpoint              | Mô tả                       |
|--------|-----------------------|-----------------------------|
| POST   | ```/api/urls```             | Tạo short URL               |
| GET    | ```/api/urls/:shortCode```  | Xem thông tin               |
| GET    | ```/api/urls```             | Lấy danh sách URL           |
| GET    | ```/:shortCode```          | Redirect to link URL gốc    |



## Thiết kế & Kỹ thuật

### Kiến trúc: 
> Handle (controller) → BLL_Business logic layer → _DAO_Data access layer -> PostgreSQL

**Lý do:**
- Mỗi tầng mỗi trách nhiệm riêng, 
- Dễ test, dễ thay đổi loại database mà không anh hưởng business logic

### Table schema
Table lưu url: có dạng trong [schema.sql](schema/schema.sql) và sẽ được tạo tự động khi container lần đầu tạo *db_data*

### Thuật toán
- Cho bộ chọn gồm ```0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz``` 62 ký tự
- Độ dài link rút gọn là 6 => một ký tự có 62 cách chọn => có 62^6 cách chọn tạo link rút gọn
- Cho phép thử lại 5 lần nếu bị trùng link rút gọn


## Challenges & Solutions
### Xử lỗi **FATAL:  sorry, too many clients already**
- Vấn đề
- Giải pháp
- Học được
    > Em nhận thấy khi có nhiều kết nối sẽ bị lỗi *too many clients already*, thêm nữa mỗi request thì Gin sẽ tạo 1 goroutines, tuy nhiên hiện tại chỉ có 1 connection tới database dẫn tới nghẽn cổ chai vì thế dựa vào dự án [BankSiM](https://github.com/giaptai/BankSim), em đã dùng:

 - Trong [postgres.go](dao/database/postgres.go) max connection là 100 => cho phép tạo 100 kết nối dồng thời tới postgres


### Limitations & Improvements
- Thuật toán tạo link rút gọn vẫn sẽ có trùng lặp
- Do ngôn ngữ lập trình chính là Java nên em có sử dụng AI với tài liệu để giải quyết bài test
- Chưa deploy ứng dụng: tuy nhiên đã có luồng để chạy
    - Tạo Dockerfile -> Build ứng dụng thành image -> đẩy lên docker hub -> dùng render pull image đó về -> chạy ứng dụng    
