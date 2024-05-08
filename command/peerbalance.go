package command

func (cmd *Command) getPeerBalance() {
	msg, status := cmd.c.PeerBalance(cmd.peerID, cmd.did)
	if !status {
		cmd.log.Error("Ping failed", "message", msg)
	} else {
		cmd.log.Info("Token Balance and number of tokens with status '12' retrieved successfully.", "message", msg)
	}

}
