package core

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/EnsurityTechnologies/uuid"
	"github.com/rubixchain/rubixgoplatform/block"
	"github.com/rubixchain/rubixgoplatform/contract"
	"github.com/rubixchain/rubixgoplatform/core/model"
	"github.com/rubixchain/rubixgoplatform/core/wallet"
	"github.com/rubixchain/rubixgoplatform/rac"
	"github.com/rubixchain/rubixgoplatform/util"
)

const (
	DTUserIDField      string = "UserID"
	DTUserInfoField    string = "UserInfo"
	DTFileInfoField    string = "FileInfo"
	DTFileHashField    string = "FileHash"
	DTFileURLField     string = "FileURL"
	DTCommiterDIDField string = "CommitterDID"
)

type DataTokenReq struct {
	DID        string
	Fields     map[string][]string
	FileNames  []string
	FolderName string
}

func (c *Core) CreateDataToken(reqID string, dr *DataTokenReq) {
	br := c.createDataToken(reqID, dr)
	dc := c.GetWebReq(reqID)
	if dc == nil {
		c.log.Error("Failed to create data token, failed to get did channel")
		return
	}
	dc.OutChan <- br
}

func (c *Core) createDataToken(reqID string, dr *DataTokenReq) *model.BasicResponse {
	br := model.BasicResponse{
		Status: false,
	}
	userID, ok := dr.Fields[DTUserIDField]
	if !ok {
		c.log.Error("Failed to create data token, user ID missing")
		br.Message = "Failed to create data token, user ID missing"
		return &br
	}
	rt := rac.RacType{
		Type:        rac.RacDataTokenType,
		DID:         dr.DID,
		TotalSupply: 1,
		CreatorID:   userID[0],
	}
	userInfo, ok := dr.Fields[DTUserIDField]
	if ok {
		rt.CreatorInput = userInfo[0]
	}
	comDid := dr.DID
	cdid, ok := dr.Fields[DTCommiterDIDField]
	if ok {
		comDid = cdid[0]
	}
	fileInfo, fok := dr.Fields[DTFileInfoField]
	if fok {
		var fi map[string]map[string]string
		err := json.Unmarshal([]byte(fileInfo[0]), &fi)
		if err != nil {
			c.log.Error("Failed to create data token, invalid file info")
			br.Message = "Failed to create data token, invalid file info"
			return &br
		}
		for k, v := range fi {
			ch, ok := v[DTFileHashField]
			if ok {
				if rt.ContentHash == nil {
					rt.ContentHash = make(map[string]string)
				}
				rt.ContentHash[k] = ch
			}
			cu, ok := v[DTFileURLField]
			if ok {
				if rt.ContentURL == nil {
					rt.ContentURL = make(map[string]string)
				}
				rt.ContentURL[k] = cu
			}
		}
	}
	dc, err := c.SetupDID(reqID, dr.DID)
	if err != nil {
		c.log.Error("Failed to create data token, failed to setup did", "err", err)
		br.Message = "Failed to create data token, failed to setup did"
		return &br
	}
	for _, file := range dr.FileNames {
		fn := strings.TrimPrefix(file, dr.FolderName+"/")
		fb, err := ioutil.ReadFile(file)
		if err != nil {
			c.log.Error("Failed to create data token, failed to read file", "err", err)
			br.Message = "Failed to create data token, failed to read file"
			return &br
		}
		hb := util.CalculateHash(fb, "SHA3-256")
		fbr := bytes.NewBuffer(fb)
		fileUrl, err := c.ipfs.Add(fbr)
		if err != nil {
			c.log.Error("Failed to create data token, failed to add file to ipfs", "err", err)
			br.Message = "Failed to create data token, failed to add file to ipfs"
			return &br
		}
		if rt.ContentHash == nil {
			rt.ContentHash = make(map[string]string)
		}
		rt.ContentHash[fn] = util.HexToStr(hb)
		if rt.ContentURL == nil {
			rt.ContentURL = make(map[string]string)
		}
		rt.ContentURL[fn] = fileUrl
	}
	dtb, err := rac.CreateRac(&rt)
	if err != nil {
		c.log.Error("Failed to create data token, failed to create rac block", "err", err)
		br.Message = "Failed to create data token, failed to create rac block"
		return &br
	}
	err = dtb[0].UpdateSignature(dc)
	if err != nil {
		c.log.Error("Failed to create data token, failed to update signature", "err", err)
		br.Message = "Failed to create data token, failed to update signature"
		return &br
	}
	rtb := dtb[0].GetBlock()
	td := util.HexToStr(rtb)
	fr := bytes.NewBuffer([]byte(td))
	dt, err := c.ipfs.Add(fr)
	if err != nil {
		c.log.Error("Failed to create data token, failed to add rac token to ipfs", "err", err)
		br.Message = "Failed to create data token, failed to add rac token to ipfs"
		return &br
	}
	err = c.w.CreateDataToken(&wallet.DataToken{TokenID: dt, DID: dr.DID, CommitterDID: comDid})
	if err != nil {
		c.log.Error("Failed to create data token, write failed", "err", err)
		br.Message = "Failed to create data token, write failed"
		return &br
	}
	dtm := make(map[string]interface{})
	dtm[dr.DID] = dt
	st := &contract.ContractType{
		Type:         contract.SCDataTokenType,
		OwnerDID:     dr.DID,
		CommitterDID: comDid,
		DataTokens:   dtm,
		Comment:      "Committed token",
	}
	sc := contract.CreateNewContract(st)
	if sc == nil {
		c.log.Error("Failed to create data token, failed to create smart contract", "err", err)
		br.Message = "Failed to create data token, failed to create smart contract"
		return &br
	}
	err = sc.UpdateSignature(dc)
	if err != nil {
		c.log.Error("Failed to create data token, smart contract signature failed", "err", err)
		br.Message = "Failed to create data token, smart contract signature failed"
		return &br
	}
	tcb := &block.TokenChainBlock{
		TransactionType: wallet.TokenGeneratedType,
		TokenOwner:      dr.DID,
		TokenType:       block.DataTokenType,
		TokenID:         dt,
		Contract:        sc.GetBlock(),
		Comment:         "Token generated at " + time.Now().String(),
	}
	ctcb := make(map[string]*block.Block)
	ctcb[dt] = nil
	blk := block.CreateNewBlock(ctcb, tcb)
	if blk == nil {
		c.log.Error("Failed to create data token, unable to create token chain")
		br.Message = "Failed to create data token, unable to create token chain"
		return &br
	}
	err = blk.UpdateSignature(dr.DID, dc)
	if err != nil {
		c.log.Error("Failed to create data token, failed to update signature", "err", err)
		br.Message = "Failed to create data token, failed to update signature"
		return &br
	}
	err = c.w.AddDataTokenBlock(dt, blk)
	if err != nil {
		c.log.Error("Failed to create data token, failed to add token chan block", "err", err)
		br.Message = "Failed to create data token, failed to add token chan block"
		return &br
	}
	br.Status = true
	br.Message = dt
	return &br
}

func (c *Core) CommitDataToken(reqID string, did string) {
	br := c.commitDataToken(reqID, did)
	dc := c.GetWebReq(reqID)
	if dc == nil {
		c.log.Error("Failed to create data token, failed to get did channel")
		return
	}
	dc.OutChan <- br
}

func (c *Core) commitDataToken(reqID string, did string) *model.BasicResponse {
	dt, err := c.w.GetDataToken(did)
	br := &model.BasicResponse{
		Status: false,
	}
	if err != nil {
		c.log.Error("Commit data token failed, failed to get data token", "err", err)
		br.Message = "Commit data token failed, failed to get data token"
		return br
	}

	dtm := make(map[string]interface{})

	for i := range dt {
		dtm[dt[i].TokenID] = dt[i].DID
	}
	dc, err := c.SetupDID(reqID, did)
	if err != nil {
		br.Message = "Failed to setup DID, " + err.Error()
		return br
	}
	sct := &contract.ContractType{
		Type:       contract.SCDataTokenCommitType,
		DataTokens: dtm,
		PledgeMode: contract.POWPledgeMode,
	}
	sc := contract.CreateNewContract(sct)
	err = sc.UpdateSignature(dc)
	if err != nil {
		c.log.Error(err.Error())
		br.Message = err.Error()
		return br
	}
	cr := &ConensusRequest{
		ReqID:         uuid.New().String(),
		Type:          QuorumTypeTwo,
		Mode:          DTCommitMode,
		SenderPeerID:  c.peerID,
		ContractBlock: sc.GetBlock(),
	}
	err = c.initiateConsensus(cr, sc, dc)
	if err != nil {
		c.log.Error("Consensus failed", "err", err)
		br.Message = "Consensus failed" + err.Error()
		return br
	}

	return br
}