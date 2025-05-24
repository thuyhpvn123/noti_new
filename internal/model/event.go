package model

type EventLog struct {
	Address         string   `json:"address"`
	BlockHash       string   `json:"blockHash"`
	BlockNumber     string   `json:"blockNumber"`
	Data        string   `json:"data"`
	LogIndex        string   `json:"logIndex"`
	Topics          []string `json:"topics"`
	TransactionHash string   `json:"TransactionHash"`
}
