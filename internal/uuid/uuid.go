package uuid

import (
	"fmt"
	"strconv"
)

func CreateUuid(senderIdentifier, receiverIdentifier []byte, counter uint64) string {
	return UuidIdentifier(senderIdentifier, receiverIdentifier) + strconv.FormatUint(counter, 10)
}

func UuidIdentifier(senderIdentifier, receiverIdentifier []byte) string {
	// TODO; fix this!! this prevents from sending!
	return string([]rune(string(senderIdentifier))) + string([]rune(string(receiverIdentifier)))
}

var BadUuid = fmt.Errorf("uuid is not identifier by sender and receiver identifier")

func PeelIdentifier(senderIdentifier, receiverIdentifier []byte, uuid string) (string, error) {
	identifier := UuidIdentifier(senderIdentifier, receiverIdentifier)
	if len(identifier) >= len(uuid) {
		return "", BadUuid
	}
	return uuid[len(identifier):], nil
}
