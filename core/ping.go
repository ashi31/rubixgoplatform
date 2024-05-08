package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rubixchain/rubixgoplatform/core/model"
	"github.com/rubixchain/rubixgoplatform/wrapper/ensweb"
)

// PingRequest is the model for ping request
type PingRequest struct {
	Message string `json:"message"`
}

// PingResponse is the model for ping response
type PingResponse struct {
	model.BasicResponse
}

// PingSetup will setup the ping route
func (c *Core) PingSetup() {
	c.l.AddRoute(APIPingPath, "GET", c.PingRecevied)
	c.l.AddRoute(APIGetTokenCount, "GET", c.getTokenCount)
}

// PingRecevied is the handler for ping request
func (c *Core) PingRecevied(req *ensweb.Request) *ensweb.Result {
	c.log.Info("Ping Received")
	resp := &PingResponse{
		BasicResponse: model.BasicResponse{
			Status: false,
		},
	}
	resp.Status = true
	resp.Message = "Ping Received"
	return c.l.RenderJSON(req, &resp, http.StatusOK)
}

// PingPeer will ping the peer & get the response
func (c *Core) PingPeer(peerID string) (string, error) {
	p, err := c.pm.OpenPeerConn(peerID, "", c.getCoreAppName(peerID))
	if err != nil {
		return "", err
	}
	// Close the p2p before exit
	defer p.Close()
	var pingResp PingResponse
	err = p.SendJSONRequest("GET", APIPingPath, nil, nil, &pingResp, false, 2*time.Minute)
	if err != nil {
		return "", err
	}
	return pingResp.Message, nil
}

func (c *Core) PingPeerWithBalance(peerID string, did string) (string, error) {
	p, err := c.pm.OpenPeerConn(peerID, did, c.getCoreAppName(peerID))
	if err != nil {
		return "", err
	}
	q := make(map[string]string)
	q["peerID"] = peerID
	q["did"] = did

	var ps model.PeerTokenCountResponse
	err = p.SendJSONRequest("GET", APIGetTokenCount, q, nil, &ps, false)
	if err != nil {
		return "", err
	}
	balance := fmt.Sprintf("%v", ps.DIDBalance)
	count := fmt.Sprintf("%v", ps.TokenStatus12)
	msg := "Balance of peer ID : " + peerID + " and DID : " + did + " is = " + balance + " and number of tokens with token status 12 is : " + count
	c.log.Info(msg)
	defer p.Close()
	return msg, nil
}

func (c *Core) getTokenCount(req *ensweb.Request) *ensweb.Result {
	did := c.l.GetQuerry(req, "did")
	peerID := c.l.GetQuerry(req, "peerID")
	var ps model.PeerTokenCountResponse
	balance, count := c.w.GetBalance(did, peerID)
	ps.DIDBalance = balance
	ps.TokenStatus12 = count
	return c.l.RenderJSON(req, &ps, http.StatusOK)

}
