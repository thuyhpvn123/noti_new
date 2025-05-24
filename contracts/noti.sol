// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title Hệ thống thông báo phi tập trung cho Dapp
 * @dev Hệ thống này gồm 4 smart contract chính:
 * 
 * 1️⃣ NotiFactory: Quản lý đăng ký Dapp & người dùng
 * 2️⃣ PermissionManager: Quản lý quyền nhận thông báo của người dùng
 * 3️⃣ NotiStorage: Lưu trữ & quản lý thông báo của từng Dapp
 * 4️⃣ NotiHub: Trung tâm tiếp nhận tất cả thông báo từ các Dapp
 *
 * 🔹 Luồng hoạt động:
 * 
 * 1️⃣ **Dapp đăng ký hệ thống thông báo:**
 *    - Dapp gọi `registerDapp()` trên `NotiFactory`.
 *    - `NotiFactory` triển khai `PermissionManager` & `NotiStorage` riêng cho Dapp đó.
 *    - `NotiStorage` được đăng ký trên `NotiHub` để gửi thông báo lên off-chain.
 *
 * 2️⃣ **Người dùng đăng ký nhận thông báo từ Dapp:**
 *    - Người dùng gọi `subscribeUser(dappAddress, publicKey, encryptedDeviceToken)`.
 *    - `NotiFactory` gọi `registerUser(userAddress)` trên `PermissionManager` của Dapp.
 *    - Người dùng gửi thêm **publicKey** và **encryptedDeviceToken** (mã hóa bằng ECDH).
 *    - **Sự kiện `UserSubscribed` phát ra** để backend xử lý lưu trữ token của thiết bị.
 *
 * 3️⃣ **Dapp gửi thông báo đến người dùng:**
 *    - Dapp gọi `AddNoti(user, message)` trên `NotiStorage`.
 *    - `NotiStorage` kiểm tra quyền của user qua `PermissionManager`.
 *    - Nếu hợp lệ, thông báo được lưu vào blockchain & phát sự kiện `NotificationSent`.
 *    - `NotiStorage` gọi `emitNotification()` trên `NotiHub` để broadcast lên off-chain.
 *
 * 4️⃣ **Hệ thống off-chain lắng nghe thông báo từ `NotiHub`:**
 *    - **NotiHub phát sự kiện `GlobalNotification`** khi có thông báo từ Dapp.
 *    - Hệ thống off-chain lắng nghe sự kiện này và xử lý gửi push notification đến user.
 *
 * 5️⃣ **Người dùng quản lý thông báo của mình:**
 *    - **Lấy danh sách thông báo**: `getNotiList(startIndex, limit)`.
 *    - **Đánh dấu đã đọc**: `markRead(eventId)`, `markAllRead(beforeTime)`.
 *    - **Xóa thông báo**: `deleteNoti(eventId)`.
 */
import "@openzeppelin/contracts/access/Ownable.sol";
import "./interfaces/INoti.sol";
// import "forge-std/console.sol";
contract PermissionManager {
    address public notiFactory;
    address public notiStorage;
    mapping(address => uint256) public users; // Lưu thời điểm user đăng ký

    modifier onlyNotiStorage() {
        require(msg.sender == notiStorage, "Not authorized-PermissionManager");
        _;
    }

    modifier onlyNotiFactory() {
        require(msg.sender == notiFactory, "Only NotiFactory can register users");
        _;
    }

    constructor( address _notiFactory) {
        notiFactory = _notiFactory;
    }
    function setNotiStorage(address _notiStorage) external onlyNotiFactory {
        notiStorage = _notiStorage;
    }
    /**
     * @notice Đăng ký người dùng vào hệ thống thông báo của Dapp
     * @dev Chỉ Dapp hoặc chính user có thể đăng ký
     * @param _user Địa chỉ của người dùng cần đăng ký
     */
    function registerUser(address _user) external onlyNotiFactory {
        users[_user] = block.timestamp;
    }

    /**
     * @notice Hủy đăng ký người dùng khỏi hệ thống
     * @dev Chỉ Dapp hoặc chính user có thể hủy
     */
    function unregisterUser(address _user) external onlyNotiFactory {
        require(users[_user] > 0, "Unauthorized");
        users[_user] = 0;
    }

    /**
     * @notice Kiểm tra người dùng có đăng ký không
     * @param _user Địa chỉ người dùng
     * @return bool True nếu user đã đăng ký, False nếu chưa
     */
    function isUserRegistered(address _user) external view onlyNotiStorage returns (bool) {
        return users[_user] > 0;
    }
}

/**
 * @title NotiHub
 * @notice Trung tâm tiếp nhận tất cả thông báo từ các Dapp.
 * @dev Chỉ nhận thông báo từ các NotiStorage hợp lệ do NotiFactory đăng ký.
 */
contract NotiHub is Ownable{
    mapping(address => bool) public allowedNotiStorages; // Danh sách các NotiStorage hợp lệ
    address public notiFactory; // Địa chỉ NotiFactory

    event GlobalNotification(
        address indexed dappOwner,
        address indexed user,
        uint256 eventId,
        string title,
        string message,
        uint256 createdAt
    );

    modifier onlyAllowedNotiStorage() {
        require(allowedNotiStorages[msg.sender], "Not authorized-NotiHub");
        _;
    }

    modifier onlyNotiFactory() {
        require(msg.sender == notiFactory, "Only NotiFactory can register");
        _;
    }

    // constructor(address _notiFactory) {
    //     notiFactory = _notiFactory;
    // }
    constructor()Ownable(msg.sender){}

    function setNotiFactory(address _notiFactory)external onlyOwner{
        notiFactory =  _notiFactory;
    }    
    /**
     * @notice Chỉ NotiFactory có thể đăng ký NotiStorage hợp lệ
     * @param notiStorage Địa chỉ NotiStorage cần cấp quyền
     */
    function registerNotiStorage(address notiStorage) external onlyNotiFactory {
        allowedNotiStorages[notiStorage] = true;
    }

    /**
     * @notice Phát sự kiện khi có thông báo mới
     * @dev Chỉ chấp nhận từ các NotiStorage hợp lệ
     */
    function emitNotification(
        address dappOwner,
        address user,
        uint256 eventId,
        string calldata title,
        string calldata message
    ) external onlyAllowedNotiStorage {
        emit GlobalNotification(dappOwner, user, eventId, title, message,block.timestamp);
    }
}


contract NotiStorage {
    address public dappOwner;
    PermissionManager public permissionManager;
    NotiHub public notiHub; // Thêm tham chiếu đến NotiHub

    struct Notification {
        uint256 eventId;
        string message;
        uint256 timestamp;
        bool isRead;
    }
    mapping(address => uint256[]) private mEventList;
    mapping(uint256 => Notification) private mEventInfo;

    event NotificationSent(address indexed user, uint256 eventId, string message);
    event NotificationRead(address indexed user, uint256 eventId);
    event NotificationDeleted(address indexed user, uint256 eventId);
    modifier onlyDappOwner() {
        require(msg.sender == dappOwner, "Not authorized-NotiStorage");
        _;
    }

    constructor(address _dappOwner, address _permissionManager, address _notiHub) {
        dappOwner = _dappOwner;
        permissionManager = PermissionManager(_permissionManager);
        notiHub = NotiHub(_notiHub); // Khởi tạo địa chỉ NotiHub
    }

    /**
     * @notice Gửi thông báo đến một người dùng
     * @dev Chỉ Dapp mới có quyền gửi thông báo
     * @param _to Địa chỉ người nhận
     * @param params Nội dung thông báo
     */
    // function AddNoti(address _user,string calldata _title, string calldata _message) external onlyDappOwner {
    function AddNoti(
        NotiParams memory params,
        address _to
    ) external onlyDappOwner {

        require(permissionManager.isUserRegistered(_to), "User not registered");

        uint256 eventId = uint256(keccak256(abi.encodePacked(block.timestamp, _to, params.body)));
        Notification memory newNoti = Notification(eventId, params.body, block.timestamp, false);

        mEventInfo[eventId] = newNoti;
        mEventList[_to].push(eventId);

        emit NotificationSent(_to, eventId, params.body);

        // Gửi sự kiện lên NotiHub để off-chain bắt một nơi duy nhất
        notiHub.emitNotification(dappOwner, _to, eventId, params.title,params.body);
    }


    /**
     * @notice Lấy danh sách thông báo của người dùng (phân trang)
     * @param _user Địa chỉ người dùng
     * @param startIndex Vị trí bắt đầu
     * @param limit Số lượng thông báo cần lấy
     * @return Notification[] Danh sách thông báo
     */
    function getNotiList(address _user, uint256 startIndex, uint256 limit) external view returns (Notification[] memory) {
        require(permissionManager.isUserRegistered(_user), "User not registered");

        uint256 total = mEventList[_user].length;
        if (startIndex >= total) return new Notification[](0) ;

        uint256 fetchCount = (total - startIndex > limit) ? limit : total - startIndex;
        Notification[] memory notiList = new Notification[](fetchCount);

        for (uint256 i = 0; i < fetchCount; i++) {
            notiList[i] = mEventInfo[mEventList[_user][startIndex + i]];
        }

        return notiList;
    }
}

contract NotiFactory is Ownable {
    struct DappContracts {
        address permissionManager;
        address notiStorage;
    }

    /**
     * @notice Sự kiện phát ra khi user đăng ký nhận thông báo từ Dapp
     * @param user Địa chỉ user đăng ký
     * @param dapp Địa chỉ Dapp
     * @param permissionManager Địa chỉ contract PermissionManager của Dapp
     * @param publicKey Public key ngẫu nhiên của user dùng để mã hóa ECDH
     * @param encryptedDeviceToken Token của device đã mã hóa
     */
    event UserSubscribed(
        address indexed user, 
        address indexed dapp, 
        address permissionManager, 
        bytes publicKey, 
        bytes encryptedDeviceToken,
        PlatformEnum platform,
        uint256 createdAt
    );
    event UserUnsubscribed(
        address indexed user, 
        address indexed dapp, 
        PlatformEnum platform,
        uint256 createdAt
    );
    struct DappInfo{
        address dapp;
        string nameDapp;       
    }
    mapping(address => DappContracts) public dappToContracts;
    mapping(address => mapping(uint => uint256[])) private mUserToScheduledTimes; //user =>platform => scheduled times
    mapping(address => mapping(uint => address[])) private mUserToScheduledApps; //user =>platform => apps scheduled in platform
    address[] public dappList;
    DappInfo[] public dappInfoList;
    NotiHub public notiHub; // Tham chiếu đến NotiHub

    event DappRegistered(address indexed dappOwner, address permissionManager, address notiStorage);
    // event UserSubscribed(address indexed user, address indexed dapp, address permissionManager);
    event NotificationScheduled(address user , address[] apps, uint256[] scheduledTimes,  uint8 platform);
    event DeleteNotificationScheduledTimes(address user, uint256[] scheduledTimes, uint8 platform);
    event DeleteNotificationScheduledApps(address user , address[] scheduledApps, uint8 platform);


    // constructor(address _notiHub) {
    //     notiHub = NotiHub(_notiHub);
    // }
    constructor() Ownable(msg.sender){}
    function setNotiHub(address _notiHub) external onlyOwner {
        notiHub = NotiHub(_notiHub);
    }
    /**
     * @notice Đăng ký Dapp với hệ thống thông báo
     * @dev Chỉ cần gọi 1 lần duy nhất, sẽ triển khai PermissionManager & NotiStorage
     */
    function registerDapp(address service, string memory nameDapp) external onlyOwner {
        require(dappToContracts[service].permissionManager == address(0), "Dapp already registered");

        PermissionManager permissionManager = new PermissionManager(address(this));
        NotiStorage notiStorage = new NotiStorage(service, address(permissionManager), address(notiHub));
        permissionManager.setNotiStorage(address(notiStorage));
        dappToContracts[service] = DappContracts(address(permissionManager), address(notiStorage));
        dappList.push(service);
        DappInfo memory dappInfo = DappInfo({
            dapp: service , 
            nameDapp: nameDapp
        });
        dappInfoList.push(dappInfo);
        // Đăng ký NotiStorage vào NotiHub
        notiHub.registerNotiStorage(address(notiStorage));

        emit DappRegistered(service, address(permissionManager), address(notiStorage));
    }
    function unregisterDapp(address service) external onlyOwner {
        for(uint256 i; i< dappList.length; i++){
            if(dappList[i] == service){
                dappList[i] = dappList[dappList.length -1];
                dappList.pop();
                dappInfoList[i] = dappInfoList[dappInfoList.length -1];
                dappInfoList.pop();
                break;    
            }
        }
        delete dappToContracts[service];
    }
    /**
     * @notice Người dùng đăng ký nhận thông báo từ một Dapp
     * @dev Gửi kèm token của device (đã mã hóa bằng ECDH) + public key của user
     * @param dappOwner Địa chỉ của Dapp cần đăng ký
     * @param publicKey Public key ngẫu nhiên của user để backend dùng ECDH giải mã
     * @param encryptedDeviceToken Token của device đã mã hóa
     */
    function subscribeUser(
        address dappOwner, 
        bytes calldata publicKey, 
        bytes calldata encryptedDeviceToken,
        PlatformEnum platform
    ) external {
        require(dappToContracts[dappOwner].permissionManager != address(0), "Dapp not registered");

        PermissionManager permissionManager = PermissionManager(dappToContracts[dappOwner].permissionManager);
        permissionManager.registerUser(msg.sender);

        emit UserSubscribed(msg.sender, dappOwner, address(permissionManager), publicKey, encryptedDeviceToken,platform,block.timestamp);
    }
    function unSubscribeUser(
        address dappOwner, 
        PlatformEnum platform
    ) external {
        require(dappToContracts[dappOwner].permissionManager != address(0), "Dapp not registered");

        PermissionManager permissionManager = PermissionManager(dappToContracts[dappOwner].permissionManager);
        permissionManager.unregisterUser(msg.sender);

        emit UserUnsubscribed(msg.sender, dappOwner, platform,block.timestamp);
    }

    /**
     * @notice Lấy địa chỉ của PermissionManager & NotiStorage của một Dapp
     * @param dappOwner Địa chỉ Dapp
     * @return (address, address) Địa chỉ PermissionManager & NotiStorage
     */
    function getDappContracts(address dappOwner) external view returns (address, address) {
        return (
            dappToContracts[dappOwner].permissionManager,
            dappToContracts[dappOwner].notiStorage
        );
    }

    /**
     * @notice Lấy danh sách tất cả Dapp đã đăng ký hệ thống thông báo
     * @return address[] Danh sách Dapp
     */
    function getAllDapps() external view returns (address[] memory) {
        return dappList;
    }
    function getAllDappInfos() external view returns (DappInfo[] memory) {
        return dappInfoList;
    }

    function insertScheduleNoti(
        address[] memory toArr,
        uint256[] memory scheduledTimes,
        uint8 platform
    ) external  {

        address[] storage apps = mUserToScheduledApps[msg.sender][platform];
        uint256[] storage times = mUserToScheduledTimes[msg.sender][platform];

        for (uint256 i = 0; i < toArr.length; i++) {   
            apps.push(toArr[i]);
        }

        for (uint256 j = 0; j < scheduledTimes.length; j++) {
            times.push(scheduledTimes[j]);
        }

        emit NotificationScheduled(msg.sender, toArr, scheduledTimes,platform );
    }

    function deleteScheduleNotiApps(address[] memory toArr,uint8 platform) external  {

        address[] storage apps = mUserToScheduledApps[msg.sender][platform];

        for (uint256 i = 0; i < toArr.length; i++) {
            for (uint256 j = 0; j < apps.length; j++) {
                if (apps[j] == toArr[i]) {
                    apps[j] = apps[apps.length - 1]; // Move last element to current position
                    apps.pop(); // Remove last element
                    break; // Stop searching after first removal
                }
            }
        }

        emit DeleteNotificationScheduledApps(msg.sender, toArr, platform);
    }

    function deleteScheduleNotiTimes(uint256[] memory scheduledTimes,uint8 platform) external  {

        uint256[] storage times = mUserToScheduledTimes[msg.sender][platform];

        for (uint256 i = 0; i < scheduledTimes.length; i++) {
            for (uint256 j = 0; j < times.length; j++) {
                if (times[j] == scheduledTimes[i]) {
                    times[j] = times[times.length - 1]; // Swap with last element
                    times.pop(); // Remove last element
                    break; // Stop searching after first removal
                }
            }
        }

        emit DeleteNotificationScheduledTimes(msg.sender, times,platform);
    }


    function getScheduledNotiTimes(address _user,uint8 platform) external view returns (uint[] memory) {
        return mUserToScheduledTimes[_user][platform];
    }
    function getScheduledNotiApps(address _user,uint8 platform) external view returns (address[] memory) {
        return mUserToScheduledApps[_user][platform];
    }

}
