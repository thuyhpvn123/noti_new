// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../contracts/noti.sol";

contract NotificationSystemTest is Test {
    // Main contracts
    NotiHub public notiHub;
    NotiFactory public notiFactory;
    
    // Addresses for testing
    address public dappOwner;
    address public user1;
    address public user2;
    address public deployer= address(0x111);
    // Generated contracts
    PermissionManager public permissionManager;
    NotiStorage public notiStorage;
    
    // Test data
    bytes public publicKey;
    bytes public encryptedDeviceToken;

    constructor(){
        // Create test accounts
        dappOwner = makeAddr("dappOwner");
        user1 = makeAddr("user1");
        user2 = makeAddr("user2");
        vm.startPrank(deployer);
        // Deploy the main contracts
        notiHub = new NotiHub(); // Test contract as factory for now
        notiFactory = new NotiFactory();
        notiFactory.setNotiHub(address(notiHub));
        notiHub.setNotiFactory(address(notiFactory));
        // Update NotiHub to recognize the actual factory
        
        // We would need a function to update notiFactory in NotiHub
        // For the test we can assume it's working or deploy a modified version
        vm.stopPrank();
        
        // Test data setup
        publicKey = bytes("userPublicKey");
        encryptedDeviceToken = abi.encodePacked("encryptedDeviceToken");
        
    }
    
    function testDappRegistration() public {
        // Register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        // Get the generated contracts
        (address pmAddr, address nsAddr) = notiFactory.getDappContracts(dappOwner);
        
        // Verify registration
        assertTrue(pmAddr != address(0), "PermissionManager was not deployed");
        assertTrue(nsAddr != address(0), "NotiStorage was not deployed");
        
        // Store for later tests
        permissionManager = PermissionManager(pmAddr);
        notiStorage = NotiStorage(nsAddr);
        
        // Verify the correct owner
        // assertEq(permissionManager.notiStorage(), notiStorage, "Wrong dapp owner in PermissionManager");
        assertEq(notiStorage.dappOwner(), dappOwner, "Wrong dapp owner in NotiStorage");
    }
    
    function testUserSubscription() public {
        // First register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        (address pmAddr, address nsAddr) = notiFactory.getDappContracts(dappOwner);
        permissionManager = PermissionManager(pmAddr);
        notiStorage = NotiStorage(nsAddr);
        
        // Now subscribe a user
        vm.startPrank(user1);
        notiFactory.subscribeUser(dappOwner, publicKey, encryptedDeviceToken,PlatformEnum.ANDROID);
        vm.stopPrank();
        
        // Verify user registration in the PermissionManager
        vm.startPrank(address(notiStorage));
        bool isRegistered = permissionManager.isUserRegistered(user1);
        vm.stopPrank();
        
        assertTrue(isRegistered, "User was not registered correctly");
    }
    
    function testSendNotification() public {
        // First register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        (address pmAddr, address nsAddr) = notiFactory.getDappContracts(dappOwner);
        permissionManager = PermissionManager(pmAddr);
        notiStorage = NotiStorage(nsAddr);
        
        // Subscribe a user
        vm.startPrank(user1);
        notiFactory.subscribeUser(dappOwner, publicKey, encryptedDeviceToken,PlatformEnum.ANDROID);
        vm.stopPrank();
        
        // Send a notification to the user
        string memory message = "Hello, this is a test notification!";
        NotiParams memory params = NotiParams({
            title: "Ecom",
            body: message
        });
        uint256 eventId = uint256(keccak256(abi.encodePacked(block.timestamp, user1, message)));
        vm.startPrank(dappOwner);
        // Expect an event to be emitted
        vm.expectEmit(true, true, false, true);
        emit NotiStorage.NotificationSent(user1, eventId, message); // eventId is not predictable, so we use 0
        
        notiStorage.AddNoti(params,user1);
        vm.stopPrank();
        
        // Verify the notification was stored properly
        // Note: we need to know the eventId to get the notification
        // Since it's a keccak256 hash, we'll check through the returned list
        vm.startPrank(dappOwner);
        NotiStorage.Notification[] memory notifications = notiStorage.getNotiList(user1, 0, 10);
        vm.stopPrank();
        
        assertEq(notifications.length, 1, "Should have 1 notification");
        assertEq(notifications[0].message, message, "Message content doesn't match");
        assertEq(notifications[0].isRead, false, "Notification should be unread initially");
    }
    
    function testNotificationPermissions() public {
        // First register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        (address pmAddr, address nsAddr) = notiFactory.getDappContracts(dappOwner);
        notiStorage = NotiStorage(nsAddr);
        
        // Try to send a notification to an unregistered user
        vm.startPrank(dappOwner);
        vm.expectRevert("User not registered");
         string memory message = "Hello, this is a test notification!";
        NotiParams memory params = NotiParams({
            title: "Ecom",
            body: message
        });
        notiStorage.AddNoti(params,user1);
        vm.stopPrank();
        
        // Try to send a notification as a non-owner
        vm.startPrank(user2);
        vm.expectRevert("Not authorized-NotiStorage");
        notiStorage.AddNoti(params,user1);
        vm.stopPrank();
    }
    
    function testUnregisterUser() public {
        // First register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        (address pmAddr, address nsAddr) = notiFactory.getDappContracts(dappOwner);
        permissionManager = PermissionManager(pmAddr);
        notiStorage = NotiStorage(nsAddr);
        
        // Subscribe a user
        vm.startPrank(user1);
        notiFactory.subscribeUser(dappOwner, publicKey, encryptedDeviceToken,PlatformEnum.ANDROID);
        vm.stopPrank();
        
        // Verify user is registered
        vm.startPrank(address(notiStorage));
        bool isRegistered = permissionManager.isUserRegistered(user1);
        vm.stopPrank();
        assertTrue(isRegistered, "User should be registered");
        
        // Unregister the user
        vm.startPrank(user1);
        notiFactory.unSubscribeUser(dappOwner,PlatformEnum.ANDROID);
        vm.stopPrank();
        
        // Verify user is no longer registered
        vm.startPrank(address(notiStorage));
        isRegistered = permissionManager.isUserRegistered(user1);
        vm.stopPrank();
        assertFalse(isRegistered, "User should not be registered after unregistering");
        
        // Try to send a notification to the unregistered user
        vm.startPrank(dappOwner);
        vm.expectRevert("User not registered");
         string memory message = "Hello, this is a test notification!";
        NotiParams memory params = NotiParams({
            title: "Ecom",
            body: message
        });
        notiStorage.AddNoti(params,user1);
        vm.stopPrank();
    }
    
    function testGetNotifications() public {
        // First register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        (address pmAddr, address nsAddr) = notiFactory.getDappContracts(dappOwner);
        permissionManager = PermissionManager(pmAddr);
        notiStorage = NotiStorage(nsAddr);
        
        // Subscribe a user
        vm.startPrank(user1);
        notiFactory.subscribeUser(dappOwner, publicKey, encryptedDeviceToken,PlatformEnum.ANDROID);
        vm.stopPrank();
        
        // Send multiple notifications
        vm.startPrank(dappOwner);
        for (uint i = 0; i < 5; i++) {
            string memory message = string(abi.encodePacked("Notification #", vm.toString(i+1)));
            NotiParams memory params = NotiParams({
                title: "Ecom",
                body: message
            });
            notiStorage.AddNoti(params,user1);
            }
        vm.stopPrank();
        
        // Retrieve notifications with pagination
        vm.startPrank(dappOwner);
        // Get first 3 notifications
        NotiStorage.Notification[] memory notifications1 = notiStorage.getNotiList(user1, 0, 3);
        // Get the remaining 2 notifications
        NotiStorage.Notification[] memory notifications2 = notiStorage.getNotiList(user1, 3, 3);
        vm.stopPrank();
        
        // Verify pagination works correctly
        assertEq(notifications1.length, 3, "Should have 3 notifications in the first page");
        assertEq(notifications2.length, 2, "Should have 2 notifications in the second page");
        
        // Verify content of notifications
        assertEq(notifications1[0].message, "Notification #1", "First notification message doesn't match");
        assertEq(notifications1[2].message, "Notification #3", "Third notification message doesn't match");
        assertEq(notifications2[0].message, "Notification #4", "Fourth notification message doesn't match");
        assertEq(notifications2[1].message, "Notification #5", "Fifth notification message doesn't match");
    }
    
    function testEmptyNotificationList() public {
        // First register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        (address pmAddr, address nsAddr) = notiFactory.getDappContracts(dappOwner);
        permissionManager = PermissionManager(pmAddr);
        notiStorage = NotiStorage(nsAddr);
        
        // Subscribe a user
        vm.startPrank(user1);
        notiFactory.subscribeUser(dappOwner, publicKey, encryptedDeviceToken,PlatformEnum.ANDROID);
        vm.stopPrank();
        
        // Retrieve notifications when there are none
        vm.startPrank(dappOwner);
        NotiStorage.Notification[] memory notifications = notiStorage.getNotiList(user1, 0, 10);
        vm.stopPrank();
        
        // Verify empty list
        assertEq(notifications.length, 0, "Should have no notifications");
        
        // Test with out-of-range index
        vm.startPrank(dappOwner);
        notifications = notiStorage.getNotiList(user1, 10, 10);
        vm.stopPrank();
        
        // Should also return empty array
        assertEq(notifications.length, 0, "Should have no notifications when index out of range");
    }
    
    function testMultipleDapps() public {
        address dappOwner2 = makeAddr("dappOwner2");
        
        // Register two different dapps
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner2,"ecom");
        vm.stopPrank();
        
        // Get contract addresses for both dapps
        (address pm1, address ns1) = notiFactory.getDappContracts(dappOwner);
        (address pm2, address ns2) = notiFactory.getDappContracts(dappOwner2);
        
        // Verify different contracts were created
        assertTrue(pm1 != pm2, "PermissionManagers should be different");
        assertTrue(ns1 != ns2, "NotiStorages should be different");
        
        // User subscribes to both dapps
        vm.startPrank(user1);
        notiFactory.subscribeUser(dappOwner, publicKey, encryptedDeviceToken,PlatformEnum.ANDROID);
        notiFactory.subscribeUser(dappOwner2, publicKey, encryptedDeviceToken,PlatformEnum.ANDROID);
        vm.stopPrank();
        
        // Send notifications from both dapps
        vm.startPrank(dappOwner);
        string memory message = "Message from Dapp 1";
        NotiParams memory params = NotiParams({
            title: "Ecom",
            body: message
        });
        NotiStorage(ns1).AddNoti(params,user1);
        vm.stopPrank();

        message = "Message from Dapp 2";
        params = NotiParams({
            title: "Ecom",
            body: message
        });
        vm.startPrank(dappOwner2);
        NotiStorage(ns2).AddNoti(params,user1);
        vm.stopPrank();
        
        // Check notifications from each dapp
        vm.startPrank(dappOwner);
        NotiStorage.Notification[] memory notis1 = NotiStorage(ns1).getNotiList(user1, 0, 10);
        vm.stopPrank();
        
        vm.startPrank(dappOwner2);
        NotiStorage.Notification[] memory notis2 = NotiStorage(ns2).getNotiList(user1, 0, 10);
        vm.stopPrank();
        
        // Verify notifications are stored separately
        assertEq(notis1.length, 1, "Should have 1 notification from Dapp 1");
        assertEq(notis2.length, 1, "Should have 1 notification from Dapp 2");
        assertEq(notis1[0].message, "Message from Dapp 1", "Message from Dapp 1 doesn't match");
        assertEq(notis2[0].message, "Message from Dapp 2", "Message from Dapp 2 doesn't match");
    }
    
    function testGetAllDapps() public {
        address dappOwner2 = makeAddr("dappOwner2");
        
        // Initially should be empty
        address[] memory initialDapps = notiFactory.getAllDapps();
        assertEq(initialDapps.length, 0, "Should have no dapps initially");
        
        // Register two dapps
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner2,"ecom");
        vm.stopPrank();
        
        // Get all dapps
        address[] memory dapps = notiFactory.getAllDapps();
        
        // Verify list
        assertEq(dapps.length, 2, "Should have 2 dapps");
        assertEq(dapps[0], dappOwner, "First dapp should be dappOwner");
        assertEq(dapps[1], dappOwner2, "Second dapp should be dappOwner2");
    }
    
    function testNotiHubIntegration() public {
        // Register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        
        (address pmAddr, address nsAddr) = notiFactory.getDappContracts(dappOwner);
        notiStorage = NotiStorage(nsAddr);
        
        // Subscribe a user
        vm.startPrank(user1);
        notiFactory.subscribeUser(dappOwner, publicKey, encryptedDeviceToken,PlatformEnum.ANDROID);
        vm.stopPrank();
        
        // Prepare to check for GlobalNotification event from NotiHub
        vm.startPrank(dappOwner);
        
        // We expect the event to be emitted from the NotiHub
        vm.expectEmit(true, true, false, true);
        uint256 eventId = uint256(keccak256(abi.encodePacked(block.timestamp, user1, "Test notification")));
        emit NotiHub.GlobalNotification(dappOwner, user1, eventId,"Ecom", "Test notification",block.timestamp); // eventId is not predictable
        
        // Send notification
        string memory message = "Test notification";
        NotiParams memory params = NotiParams({
            title: "Ecom",
            body: message
        });
        notiStorage.AddNoti(params,user1);
        vm.stopPrank();
        
        // Verify NotiStorage is registered in NotiHub
        bool isRegistered = notiHub.allowedNotiStorages(nsAddr);
        assertTrue(isRegistered, "NotiStorage should be registered in NotiHub");
    }
    
    // Additional helper test to verify dapp can't register twice
    function testDappCannotRegisterTwice() public {
        // Register a dapp
        vm.startPrank(deployer);
        notiFactory.registerDapp(dappOwner,"ecom");
        
        // Try to register again
        vm.expectRevert("Dapp already registered");
        notiFactory.registerDapp(dappOwner,"ecom");
        vm.stopPrank();
        GetByteCode();
    }
    function GetByteCode()public {
        //getAllDappInfos
        bytes memory bytesCodeCall = abi.encodeCall(
            notiFactory.getAllDappInfos,
            ()
        );
        console.log("notiFactory getAllDappInfos: ");
        console.logBytes(bytesCodeCall);
        console.log(
            "-----------------------------------------------------------------------------"
        );
    }
}