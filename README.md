# Tổng quan Services trong `noti-contract`

## 1. `GetReceiverAllNotification`
- **Mô tả**: Lấy tất cả thông báo của người nhận dựa trên các tham số như `Receiver`, `DeviceToken`, `Repo` với phân trang.
- **Luồng hoạt động**:
  - Nhận và ràng buộc các tham số từ query.
  - Xác thực cấu trúc dữ liệu yêu cầu.
  - Kiểm tra và thiết lập giá trị mặc định cho `Page`.
  - Gọi `usecase.GetNotificationsByReceiver` để lấy thông báo.
  - Tạo đối tượng `Pagination` và `GetReceiverNotificationsResponse`.
  - Trả về phản hồi với dữ liệu thông báo.

## 2. `MarkAllRead`
- **Mô tả**: Đánh dấu tất cả thông báo đã đọc dựa trên `Receiver`, `DeviceToken` và `Repo`.
- **Luồng hoạt động**:
  - Nhận và ràng buộc các tham số từ query.
  - Xác thực cấu trúc dữ liệu yêu cầu.
  - Gọi `usecase.MarkAllRead` để đánh dấu tất cả thông báo đã đọc.
  - Trả về phản hồi thành công.

## 3. `MarkRead`
- **Mô tả**: Đánh dấu thông báo đã đọc dựa trên `ID`.
- **Luồng hoạt động**:
  - Nhận và ràng buộc các tham số từ query.
  - Xác thực cấu trúc dữ liệu yêu cầu.
  - Chuyển đổi `ID` từ chuỗi sang số nguyên.
  - Gọi `usecase.MarkRead` để đánh dấu thông báo đã đọc.
  - Trả về phản hồi thành công.

## 4. `Delete`
- **Mô tả**: Xóa thông báo đã đọc dựa trên `ID`.
- **Luồng hoạt động**:
  - Nhận và ràng buộc các tham số từ query.
  - Xác thực cấu trúc dữ liệu yêu cầu.
  - Chuyển đổi `ID` từ chuỗi sang số nguyên.
  - Gọi `usecase.Delete` để xóa thông báo.
  - Trả về phản hồi thành công.

## 5. `SendNotification`
- **Mô tả**: Gửi thông báo đến thiết bị iOS.
- **Luồng hoạt động**:
  - Nhận và ràng buộc các tham số từ query.
  - Gọi `apns.PushIosNotification` để gửi thông báo đến thiết bị iOS.

# Luồng Services trong `noti-contract`

https://www.mermaidchart.com/play

## Luồng cho GetReceiverAllNotification

```text
sequenceDiagram
    participant Client
    participant Controller as NotiController
    participant Usecase as NotificationUsecase
    participant Repository as NotificationRepository

    Client->>Controller: Send GET request to /noti/all
    Controller->>Controller: Bind query parameters
    alt Validation fails
        Controller->>Client: Return 400 Bad Request
    else Validation succeeds
        Controller->>Usecase: Call GetNotificationsByReceiver
        Usecase->>Repository: Query notifications
        Repository->>Usecase: Return notifications and total count
        Usecase->>Controller: Return notifications and total count
        Controller->>Client: Return 200 OK with notifications
    end
```

## Luồng cho MarkRead

```text
sequenceDiagram
    participant Client
    participant Controller as NotiController
    participant Usecase as NotificationUsecase
    participant Repository as NotificationRepository

    Client->>Controller: Send POST request to /noti/mark_read
    Controller->>Controller: Bind query parameters
    alt Validation fails
        Controller->>Client: Return 400 Bad Request
    else Validation succeeds
        Controller->>Controller: Convert ID from string to integer
        alt Conversion fails
            Controller->>Client: Return 400 Bad Request
        else Conversion succeeds
            Controller->>Usecase: Call MarkRead with ID, Receiver, DeviceToken
            alt MarkRead fails
                Controller->>Client: Return 400 Bad Request
            else MarkRead succeeds
                Controller->>Client: Return 200 OK
            end
        end
    end
```

## Luồng cho MarkAllRead

```text
sequenceDiagram
    participant Client
    participant Controller as NotiController
    participant Usecase as NotificationUsecase

    Client->>Controller: Send POST request to /noti/mark_all_read
    Controller->>Controller: Bind query parameters
    alt Validation fails
        Controller->>Client: Return 400 Bad Request
    else Validation succeeds
        Controller->>Usecase: Call MarkAllRead
        alt MarkAllRead fails
            Controller->>Client: Return 400 Bad Request
        else MarkAllRead succeeds
            Controller->>Client: Return 200 OK with true
        end
    end
```

## Luồng cho Delete

```text
sequenceDiagram
    participant Client
    participant Controller as NotiController
    participant Usecase as NotificationUsecase
    participant Repository as NotificationRepository

    Client->>Controller: Send DELETE request to /noti/delete
    Controller->>Controller: Bind query parameters
    alt Validation fails
        Controller->>Client: Return 400 Bad Request
    else Validation succeeds
        Controller->>Usecase: Call Delete
        Usecase->>Repository: Delete notification by ID
        Repository->>Usecase: Return success or error
        alt Deletion successful
            Usecase->>Controller: Return success
            Controller->>Client: Return 200 OK
        else Deletion fails
            Usecase->>Controller: Return error
            Controller->>Client: Return 400 Bad Request
        end
    end
```
