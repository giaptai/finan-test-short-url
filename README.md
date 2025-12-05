## Mục lục
1. [Bài toán](#bài-toán)
2. [Cách chạy](#cách-chạy)
3. [Thiết kế, Quyết định kỹ thuật & Trade-offs](#thiết-kế-quyết-định-kỹ-thuật--trade-offs)
4. [Trade-offs](#trade-offs)
5. [Challenges, Solutions & Learns](#challenges-solutions--learns)
6. [Performance & Scalability](#performance--scalability)
7. [Limitations & Improvements](#limitations--improvements)
---

## Bài toán
> Em có một 1 URL dài và muốn nó ngắn lại nên em tạo 1 dịch vụ bằng golang giúp độ dài URL ngắn lại, dễ theo dõi và chia sẻ cho người khác

**Why?**
- URLs thường dài và khó nhớ VD: `https://example.com/products/category/item?utm_source=facebook&utm_campaign=summer2024` (90+ ký tự).
- Ứng dụng giới hạn ký tự: Twitter, SMS v.v
- Khó chia sẻ qua giấy hoặc nói miệng.

**Use case:**
- Chia sẻ link (một vài ứng dụng giới hạn ký tự).
- Theo dõi số lần click vào link rút gọn (analytics cho marketing campaigns).
- Printed materials: QR codes, business cards, posters.
- User experience: Dễ nhớ, dễ gõ (go.company.com/docs thay vì URL dài).


## Cách chạy
**Yêu cầu**
- Trên máy tính có golang bản 1.24.x.
- Database: PostgreSQL.
- Có Git, Docker.

**Chạy chương trình**
```bash
# 1. Clone repo
git clone https://github.com/giaptai/finan-test-short-url
cd finan-test-short-url

# 2. Cấu hình database 
Vào thư mục chứa code tìm file .env.sample đổi thành .env 
Chỉnh thông số cho phù hợp thông số khai báo theo file docker-compose

# 3. Chạy PostgreSQL qua Docker compose
docker-compose up -d

# 4. Chạy ứng dụng
go run main.go
```
Ứng dụng sẽ chạy tại ```http://localhost:8088```

**API Endpoints**
| Method | Endpoint              | Mô tả                       |
|--------|-----------------------|-----------------------------|
| POST   | ```/api/urls```             | Tạo short URL               |
| GET    | ```/api/urls/:shortCode```  | Xem thông tin               |
| GET    | ```/api/urls```             | Lấy danh sách URL           |
| GET    | ```/:shortCode```          | Redirect to link URL gốc    |



## Thiết kế, Quyết định kỹ thuật & Trade-offs

**Kiến trúc:**
> Handle (controller) → BLL_Business logic layer → _DAO_Data access layer → PostgreSQL.

**Lý do:**
- Mỗi tầng mỗi trách nhiệm riêng.
    - Handler: Xử lý HTTP requests/responses
    - Service: Business logic, validation, SSRF protection
    - DAO: Database operations (CRUD)
- Dễ test, dễ thay đổi loại database mà không anh hưởng business logic

**Database: PostgreSQL**
| Database | Ưu điểm | Nhược điểm | Khi nào dùng |
|----------|---------|------------|--------------|
| Redis | Siêu nhanh | Data dễ mất, RAM đắt | Caching layer, session storage |
| MongoDB | Flexible schema | Overkill | Complex schemas |
| **PostgreSQL** | ACID, persistent, queries | Chậm hơn Redis | Long-term storage, Complex schemas |

**Table Schema & Index Strategy:** Schema trong [schema.sql](schema\schema.sql) 

**API Design: REST**
| Công nghệ | Đặc điểm chính | Trường hợp sử dụng |
|-----------|----------------|--------------------|
| **REST**  | Phổ biến, được hỗ trợ rộng rãi, dễ test | API phổ biến, yêu cầu tương thích rộng |
| GraphQL | Linh hoạt, tránh over-fetching/under-fetching | Ứng dụng phức tạp, client cần dữ liệu tùy biến |
| gRPC  | Tốc độ cao, hỗ trợ real-time, tối ưu microservices | Giao tiếp nội bộ, hệ thống microservices |


**Thuật toán Base62**

**Cách hoạt động:**

- Cho bộ chọn gồm ```0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz``` 62 ký tự
- Độ dài link rút gọn là 6 → một ký tự có 62 cách chọn → có 62^6 cách chọn tạo link rút gọn
- Cho phép thử lại 5 lần nếu bị trùng link rút gọn.

**So sánh:**
| Approach | Ví dụ | Ưu điểm | Nhược điểm |
|----------|-------|---------|------------|
| Auto-increment | `/1`, `/2` | Không trùng, đơn giản | Dài, predictable, lộ số lượng URLs |
| UUID | `/550e8400-e29b...` | Unique (chưa chắc) | Quá dài (36 chars) |
| Hash (MD5) | `/5d41402abc...` | Cố định | Dài, có thể collision |
| **Base62 Random** | `/QgCHdO` | Ngắn (6 chars), khó đoán, dễ đọc | Có thể collision |


## Challenges, Solutions & Learns
### Xử lỗi **FATAL:  sorry, too many clients already**
- Vấn đề: Em nhận thấy khi có nhiều kết nối sẽ bị lỗi *too many clients already*
- Giải pháp
- Học được
    > mỗi request thì Gin sẽ tạo 1 goroutines, tuy nhiên hiện tại chỉ có 1 connection tới database dẫn tới nghẽn cổ chai vì thế dựa vào dự án [BankSiM](https://github.com/giaptai/BankSim), em đã dùng:

 - Trong [postgres.go](dao/database/postgres.go) max connection là 100 => cho phép tạo 100 kết nối dồng thời tới postgres


## Limitations & Improvements
- Thuật toán tạo link rút gọn vẫn sẽ có trùng lặp
- Do ngôn ngữ lập trình chính là Java nên em có sử dụng AI với tài liệu để giải quyết bài test
- Chưa deploy ứng dụng: tuy nhiên đã có luồng để chạy
    - Tạo Dockerfile -> Build ứng dụng thành image -> đẩy lên docker hub -> dùng render pull image đó về -> chạy ứng dụng    
