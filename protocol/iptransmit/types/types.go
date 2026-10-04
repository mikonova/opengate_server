package types

const (
	Hello                 byte = iota // hello type to identify other client
	KeepAlive                         // Keepalive to help the connection stay open
	Abort                             // Stop communicating. Universal request
	ClientSend                        // Client sends a message
	ClientReceiveResponse             // Answer that the message has been delivered
	ClientSyncReq                     // Client asks the signal server for syncing. SPECIAL request type containing ip within the MESSAGE field
	MalformedReqErr                   // Client failed to identify the message from server or from another client
	ClientError                       // Client encountered an error
	ServerError                       // Server Encountered an error and wants the client to know
)
