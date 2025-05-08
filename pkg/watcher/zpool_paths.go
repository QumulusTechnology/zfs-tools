package watcher

// Common zpool binary paths
const (
	ZpoolDefault      = "zpool"                 // Let the system find it in PATH
	ZpoolUsrSbin      = "/usr/sbin/zpool"       // Common location on many Linux/Unix systems
	ZpoolSbin         = "/sbin/zpool"           // Alternative location on some systems
	ZpoolUsrLocalSbin = "/usr/local/sbin/zpool" // Common on FreeBSD or custom installations
)

// ZpoolCommand represents a zpool command path option
type ZpoolCommand string

// Available zpool command paths
const (
	ZpoolCmdDefault      ZpoolCommand = ZpoolDefault
	ZpoolCmdUsrSbin      ZpoolCommand = ZpoolUsrSbin
	ZpoolCmdSbin         ZpoolCommand = ZpoolSbin
	ZpoolCmdUsrLocalSbin ZpoolCommand = ZpoolUsrLocalSbin
)
