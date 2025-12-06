## Mục lục
1. [Bài toán](#bài-toán)
2. [Cách chạy](#cách-chạy)
3. [Thiết kế, Quyết định kỹ thuật & Trade-offs](#thiết-kế-quyết-định-kỹ-thuật--trade-offs)
4. [Challenges, Solutions & Learns](#challenges-solutions--learns)
5. [Performance & Scalability](#performance--scalability)
6. [Limitations & Improvements](#limitations--improvements)
---

## Bài toán
> Em có một 1 URL dài và muốn nó ngắn lại nên em tạo 1 dịch vụ bằng golang giúp độ dài URL ngắn lại, dễ theo dõi và chia sẻ cho người khác. Khi user click vào link rút gọn, hệ thống sẽ redirect về URL gốc và theo dõi số lần click.

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

---

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

**Quyết định**: Em chọn **PostgreSQL** vì:
- URLs cần lưu lâu dài
- PostgreSQL cũng hỗ trợ tốt truy vấn cơ bản và cả phức tạp
- Cost-effective: Disk rẻ hơn RAM
- Performance đủ nhanh

**Trade-off**: Chậm hơn Redis (~1ms vs ~0.1ms) để đổi lấy reliability và cost savings.

**Table Schema & Index Strategy:** Schema trong [schema.sql](schema/schema.sql) 

**API Design: REST**
| Công nghệ | Đặc điểm chính | Trường hợp sử dụng |
|-----------|----------------|--------------------|
| **REST**  | Phổ biến, được hỗ trợ rộng rãi, dễ test | API phổ biến, yêu cầu tương thích rộng |
| GraphQL | Linh hoạt, tránh over-fetching/under-fetching | Ứng dụng phức tạp, client cần dữ liệu tùy biến |
| gRPC  | Tốc độ cao, hỗ trợ real-time, tối ưu microservices | Giao tiếp nội bộ, hệ thống microservices |

**Quyết định**: Em chọn **REST** vì:
- Use case đơn giản (CRUD operations)
- REST phổ biến, được hỗ trợ rộng rãi, dễ test
- Phù hợp với yêu cầu đề bài

**Trade-off**: Over-fetching data (trả về toàn bộ object) là acceptable cho use case này.

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

**Quyết định**: Em chọn **Base62 Random** vì:
- Dễ hiểu và dễ triển khai
- Ngắn nhất: 6 chars vs 36 chars (UUID) vs 10+ chars (Auto-increment với 1M records)
- Security: Unpredictable (user không đoán được URLs khác)
- UX: Human-readable (không có ký tự đặc biệt), dễ share

**Trade-off**: 
- Với 1M URLs: Tỉ lệ có **ít nhất 1 cặp trùng nhau** là **99.985%** (Birthday Paradox)
- Xác suất 1 URL mới tạo trùng với bất kỳ URL nào khác trong 1M URLs: ~0.00176%

**Công thức:**
- Birthday Paradox: P = 1 - e^(-n²/(2×N)) 
- 1 url cụ thể trùng bất kỳ url nào trong 1m url:
    - P = 1 - (1 - 1/N)^(n-1)
- Trong đó: n=1M URLs, N = 62^6

---

## Challenges, Solutions & Learns
1. **FATAL:  sorry, too many clients already**
- Vấn đề: Em nhận thấy khi có nhiều kết nối sẽ bị lỗi *too many clients already*
- Giải pháp: Chỉnh connection pool 
- Học được:
    - Mỗi request thì Gin sẽ tạo 1 goroutines, tuy nhiên hiện tại chỉ có 1 connection tới database dẫn tới nghẽn cổ chai
    - `*sql.DB` trong Go là connection pool manager (giống HikariCP trong Java)
    - Dựa vào dự án [BankSiM](https://github.com/giaptai/BankSim), em đã config:
    ```go
      db.SetMaxOpenConns(100)
      db.SetMaxIdleConns(5)
      db.SetConnMaxIdleTime(60 * time.Second)
    ```

2. **SSRF Security Vulnerability**
- Vấn đề: Người dùng có thể tạo short link từ localhost, internal network - nói chung là dải địa chỉ tùy ý
- Giải pháp: chặn lại bằng hàm isPrivateIP kiểm tra hostname trước khi tạo trong [service.go](bll/service.go)
- Học được: SSRF là gì? Luôn valid nội dung trước khi xử lý, cách phòng tránh bằng Golang

3. **Async Click Tracking Trade-off**
- Vấn đề: khi user click vào short URL thì lúc này có 2 hoạt động: 1 tăng clicks trong database và 2 redirect tới link gốc
- Giải pháp: dùng từ khóa go cho việc tăng biến đếm click và redirect tới url gốc ngay
- Học được: asynchronous và PostgreSQL có cơ chế khóa hàng

---

## Performance & Scalability
> Performance: Nếu có 1 triệu links thì query ra sao? Có cần index không?

1. Nếu query tìm một url cụ thể trong 1M links thì trong [schema.sql](schema/schema.sql) có có cột short_code là unique - một index đặc biệt rồi

```sql

EXPLAIN ANALYZE SELECT * FROM url WHERE short_code ='QgCHdO';
                                                       QUERY PLAN                                                        
-------------------------------------------------------------------------------------------------------------------------
 Index Scan using url_short_code_key on url  (cost=0.29..8.30 rows=1 width=78) (actual time=1.932..1.935 rows=1 loops=1)
   Index Cond: ((short_code)::text = 'QgCHdO'::text)
 Planning Time: 1.566 ms
 Execution Time: 3.102 ms

```
**Phân tích:**
- Index Scan using `url_short_code_key` → PostgreSQL tự động dùng UNIQUE index
- B-tree index có độ phức tạp O(log n)
- Với 27K URLs: 3.102ms
- **Dự đoán với 1M URLs**: ~5-10ms (vẫn rất nhanh vì O(log n))


2. Nếu lấy danh sách thì em nghĩ sẽ đánh index trên cột created_at vì hiện tại là đang Seq Scan

```sql

EXPLAIN ANALYZE SELECT * FROM url ORDER BY created_at;
                                                  QUERY PLAN                                                   
---------------------------------------------------------------------------------------------------------------
 Sort  (cost=2767.39..2836.85 rows=27782 width=78) (actual time=55.040..57.140 rows=27785 loops=1)
   Sort Key: created_at
   Sort Method: quicksort  Memory: 3373kB
   ->  Seq Scan on url  (cost=0.00..716.82 rows=27782 width=78) (actual time=1.379..26.284 rows=27785 loops=1)
 Planning Time: 6.589 ms
 Execution Time: 59.645 ms

```
**Phân tích:**
- Sequential Scan → quét toàn bộ table (O(n))
- Với 27K URLs: 59.645ms
- **Dự đoán với 1M URLs**: ~2-3 giây (quá chậm!)
- Sau khi thêm index: dự đoán ~30-50ms (cải thiện 60x)


```sql
-- Đánh index trên cột created_at
CREATE INDEX idx_url_created_at ON url(created_at DESC);
```
**Bảng so sánh performance:**

| Số lượng URLs | Query by short_code (có index) | List URLs (không index) | List URLs (có index) |
|---------------|-------------------------------|------------------------|---------------------|
| 27K (hiện tại) | 3ms | 60ms | ~10ms |
| 100K | ~4ms | ~200ms | ~15ms |
| 1M | ~5-10ms | ~2-3s | ~30-50ms |

**Kết luận:**
- Query by `short_code`: KHÔNG cần thêm index (đã có UNIQUE constraint)
- List URLs: CẦN thêm index trên `created_at` để scale tốt

## Limitations & Improvements

**1. Thuật toán tạo link rút gọn**

**Collision với scale lớn:**
- Hiện tại: Base62 Random (6 chars) có 99.985% collision với 1M URLs
- Impact: Retry mechanism phải chạy nhiều lần, chậm hơn
- **Giải pháp tương lai:**
  1. **Feistel Network** (preferred): Pseudo-random permutation, 0% collision, reversible
  2. **Counter-based + Base62**: Auto-increment → encode Base62, cần distributed counter
  3. **Tăng length lên 8 chars**: Giảm collision xuống 0.017% với 1M URLs
  4. **Timestamp (3) + base62 (3)**: Giảm collision

**2. Missing Features:**
- Không có custom short codes, rate limiting, authentication
- Không có URL expiration, analytics dashboard
- Chưa có caching layer (Redis) cho hot URLs

**3. Production Readiness:**
- Chưa deploy ứng dụng: tuy nhiên đã có luồng để chạy
    - Tạo Dockerfile → Build ứng dụng thành image → đẩy lên docker hub → dùng render pull image đó về → chạy ứng dụng

**Acknowledgments:**
- Do ngôn ngữ lập trình chính là Java nên em có sử dụng AI với tài liệu Golang để giải quyết bài test
- Tham khảo từ [tài liệu golang](https://go.dev/doc/)

