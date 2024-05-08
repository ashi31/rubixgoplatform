package wallet

//TODO:Change function for part tokens

func (w *Wallet) GetBalance(did string, peerID string) (float64, float64) {
	var t []Token
	err := w.s.Read(TokenStorage, &t, "did=?", did)
	if err != nil {
		w.log.Error("Failed to get tokens", "err", err)
		return 0.0, 0.0
	}
	var balance float64
	var count float64
	for _, at := range t {
		if at.TokenStatus == TokenIsFree {
			balance++
		} else if at.TokenStatus == 12 {
			count++
		}

	}
	// fmt.Println(balance)
	return balance, count
}
