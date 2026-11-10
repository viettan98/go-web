Cách chạy: giải nén file 
1. Mở MySql lên chạy lệnh này:
    DROP DATABASE IF EXISTS gobook_db;
    CREATE DATABASE IF NOT EXISTS gobook_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
2. Xem Hostname, Port với mật khẩu trong mysql rồi thế vô trong 2 file main.go (như hình)
<img width="313" height="127" alt="mysql" src="https://github.com/user-attachments/assets/d4aed201-42b8-4d15-bb6c-c470fbd1d427" />
<img width="1243" height="461" alt="code" src="https://github.com/user-attachments/assets/e6b216e6-e6d6-409e-ab04-0f342f24f810" />

3. Mở folder go-web mới tải về trong vscode
4. Chạy lệnh này: go mod tidy
5. Mở file excel trong cmd/import lên nhập dữ liệu vào
6. Sau khi nhập hết dữ liệu vô file excel xong thì chạy lệnh go run cmd/server/main.go
7. Mở thêm 1 tab terminal
8. Đi vào folder import (lệnh đi nhanh: cd cmd/import) rồi chạy lệnh go run main.go
9. Muốn test thử được chưa thì kéo file books_nosql.json trong folder import ra folder go-web rồi mở web bằng link: http://127.0.0.1:18085,
tài khoản Admin: admin/admin123, tài khoản user tự tạo.

*Trong file excel điền cho t tên sách, tên tác giả, thể loại, price, như trên mạng
 stock điền bao nhiêu cũng được tầm trên 100
 description có sẵn thì điền vào, không có thì kêu AI chế
 image_url copy image link rồi điền vào

 T để sẵn 3 cái ví dụ
