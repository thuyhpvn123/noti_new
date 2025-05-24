package network

type DeviceType uint8

const (
	ANDROID DeviceType = iota
	IOS     DeviceType = 1
	WEB     DeviceType = 2
)
