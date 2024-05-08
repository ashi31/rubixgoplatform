package server

import (
	"net/http"
	"time"

	"github.com/rubixchain/rubixgoplatform/core/model"
	"github.com/rubixchain/rubixgoplatform/wrapper/ensweb"
)

// BasicResponse will send basic mode response
func (s *Server) BasicResponse(req *ensweb.Request, status bool, msg string, result interface{}) *ensweb.Result {
	resp := model.BasicResponse{
		Status:  status,
		Message: msg,
		Result:  result,
	}
	return s.RenderJSON(req, &resp, http.StatusOK)
}

// ShowAccount godoc
// @Summary      Start Core
// @Description  It will setup the core if not done before
// @Tags         Basic
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.BasicResponse
// @Router       /api/start [get]
func (s *Server) APIStart(req *ensweb.Request) *ensweb.Result {
	status, msg := s.c.Start()
	return s.BasicResponse(req, status, msg, nil)
}

// APIStart will setup the core
func (s *Server) APIShutdown(req *ensweb.Request) *ensweb.Result {
	go s.shutDown()
	return s.BasicResponse(req, true, "Shutting down...", nil)
}

// APIStart will setup the core
func (s *Server) APINodeStatus(req *ensweb.Request) *ensweb.Result {
	ok := s.c.NodeStatus()
	if ok {
		return s.BasicResponse(req, true, "Node is up and running", nil)
	} else {
		return s.BasicResponse(req, false, "Node is down, please check logs", nil)
	}
}

func (s *Server) shutDown() {
	s.log.Info("Shutting down...")
	time.Sleep(2 * time.Second)
	s.sc <- true
}

// APIPing will ping to given peer
func (s *Server) APIPing(req *ensweb.Request) *ensweb.Result {
	peerdID := s.GetQuerry(req, "peerID")
	str, err := s.c.PingPeer(peerdID)
	if err != nil {
		s.log.Error("ping failed", "err", err)
		return s.BasicResponse(req, false, str, nil)
	}
	return s.BasicResponse(req, true, str, nil)
}

func (s *Server) APIGetPeerBalance(req *ensweb.Request) *ensweb.Result {
	peerdID := s.GetQuerry(req, "peerID")
	did := s.GetQuerry(req, "did")
	str, err := s.c.PingPeerWithBalance(peerdID, did)
	if err != nil {
		s.log.Error("ping failed", "err", err)
		return s.BasicResponse(req, false, str, nil)
	}
	return s.BasicResponse(req, true, str, nil)
}
