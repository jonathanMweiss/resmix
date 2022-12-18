package rrpc

type EccSendable interface {
	InsertECCPayload(shards [][][]byte, relayIndex int)
}
type EccReceivable interface {
	PutIntoShards(shards [][][]byte)
}

type eccClientParcel Parcel

func (e *eccClientParcel) InsertECCPayload(shards [][][]byte, relayIndex int) {
	e.Payload = make([][]byte, len(shards))
	for j := 0; j < len(shards); j++ {
		e.Payload[j] = shards[j][relayIndex]
	}
}

func (e *eccClientParcel) PutIntoShards(shards [][][]byte) {
	for i := 0; i < len(shards); i++ {
		shards[i][e.RelayIndex] = e.Payload[i]
	}
}

type eccServerParcel RrpcResponse

func (e *eccServerParcel) InsertECCPayload(shards [][][]byte, relayIndex int) {
	e.Payload = make([][]byte, len(shards))
	for j := 0; j < len(shards); j++ {
		e.Payload[j] = shards[j][relayIndex]
	}
}

func (e *eccServerParcel) PutIntoShards(shards [][][]byte) {
	for i := 0; i < len(shards); i++ {
		shards[i][e.RelayIndex] = e.Payload[i]
	}
}
