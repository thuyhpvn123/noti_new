# Luồng Hoạt Động Các Hàm Trong `Event.sol`

## 1. `SetNotifiers`
- **Mô tả**: Thêm một địa chỉ vào danh sách các notifiers.
- **Luồng hoạt động**:
  - Kiểm tra xem người gọi có phải là chủ sở hữu không.
  - Thêm địa chỉ vào danh sách `mNotifiers`.
  - Thêm địa chỉ vào mảng `notifiers`.
  - Trả về `true` nếu thành công.

## 2. `DeleteNotifier`
- **Mô tả**: Xóa một địa chỉ khỏi danh sách các notifiers.
- **Luồng hoạt động**:
  - Kiểm tra xem người gọi có phải là chủ sở hữu không.
  - Đặt giá trị `false` cho địa chỉ trong `mNotifiers`.
  - Xóa địa chỉ khỏi mảng `notifiers`.
  - Trả về `true` nếu thành công.

## 3. `SetDeviceToken`
- **Mô tả**: Thiết lập token thiết bị cho một địa chỉ cụ thể.
- **Luồng hoạt động**:
  - Kiểm tra xem người gọi có phải là một notifier không.
  - Tạo một đối tượng `Device` mới với token và platform.
  - Lưu đối tượng `Device` vào `mDeviceToken`.
  - Trả về `true` nếu thành công.

## 4. `GetDeviceToken`
- **Mô tả**: Lấy token thiết bị của một địa chỉ cụ thể.
- **Luồng hoạt động**:
  - Trả về token thiết bị từ `mDeviceToken`.

## 5. `SetController`
- **Mô tả**: Thiết lập một địa chỉ làm controller.
- **Luồng hoạt động**:
  - Kiểm tra xem người gọi có phải là chủ sở hữu không.
  - Đặt giá trị `true` cho địa chỉ trong `mController`.
  - Phát ra sự kiện `eSetController`.

## 6. `AddNoti`
- **Mô tả**: Thêm một thông báo mới.
- **Luồng hoạt động**:
  - Kiểm tra xem người gọi có phải là một notifier không.
  - Kiểm tra xem token thiết bị có tồn tại không.
  - Tạo một đối tượng `EventInfo` mới và lưu vào `mEventInfo`.
  - Thêm ID sự kiện vào `mEventList`.
  - Phát ra sự kiện `NotiEvent`.
  - Tăng `TotalEvent`.
  - Trả về `true` nếu thành công.

## 7. `AddMultipleNoti`
- **Mô tả**: Thêm nhiều thông báo cùng lúc.
- **Luồng hoạt động**:
  - Kiểm tra xem người gọi có phải là một notifier không.
  - Lặp qua danh sách địa chỉ và gọi hàm `AddNoti` cho từng địa chỉ.
  - Trả về `true` nếu thành công.

## 8. `GetNotiInfo`
- **Mô tả**: Lấy thông tin của một thông báo cụ thể.
- **Luồng hoạt động**:
  - Trả về đối tượng `EventInfo` từ `mEventInfo`.

## 9. `GetNotiList`
- **Mô tả**: Lấy danh sách thông báo của người gọi.
- **Luồng hoạt động**:
  - Lấy danh sách ID sự kiện từ `mEventList`.
  - Tạo một mảng `EventInfo` mới và điền thông tin từ `mEventInfo`.
  - Trả về mảng `EventInfo`.

# Luồng Smart Contract Noti

https://www.mermaidchart.com/play

## Luồng cho Event.sol

```text
sequenceDiagram
    participant dApp
    participant NotiContract as NotificationContract
    participant UserDevice

    dApp->>NotiContract: SetDeviceToken(userAddress, deviceToken, platform)
    alt DeviceToken not set
        NotiContract->>UserDevice: Register device token
    else DeviceToken already set
        NotiContract->>UserDevice: Device already registered
    end

    dApp->>NotiContract: AddNoti(params, userAddress)
    alt Notification added
        NotiContract->>UserDevice: Send notification
    else Notification failed
        NotiContract->>dApp: Return false
    end

    dApp->>NotiContract: AddMultipleNoti(params, userAddresses)
    loop For each userAddress
        alt Notification added
            NotiContract->>UserDevice: Send notification
        else Notification failed
            NotiContract->>dApp: Return false
        end
    end

    dApp->>NotiContract: GetNotiInfo(notiId)
    NotiContract->>dApp: Return EventInfo

    dApp->>NotiContract: GetNotiList()
    NotiContract->>dApp: Return list of EventInfo
```