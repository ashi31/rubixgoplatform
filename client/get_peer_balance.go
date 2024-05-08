package client

import (
	"github.com/rubixchain/rubixgoplatform/core/model"
	"github.com/rubixchain/rubixgoplatform/setup"
)

func (c *Client) PeerBalance(peerID string, did string) (string, bool) {
	q := make(map[string]string)
	q["peerID"] = peerID
	q["did"] = did
	var rm model.BasicResponse
	err := c.sendJSONRequest("GET", setup.APIGetPeerBalance, q, nil, &rm)
	if err != nil {
		return "Failed to get Balance or tokens" + err.Error(), false
	}
	return rm.Message, rm.Status
}
