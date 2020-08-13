package auth

// Nonce a user's nonce
type Nonce struct {
	nonces []string
}

// NewNonce create a *Nonce struct
func NewNonce() *Nonce {
	return &Nonce{}
}

const (
	maxNonceTTL  = 3600
	maxNonceSize = 20
)

// Set save a nonce string, if existed, return false
func (this *Nonce) Set(nonce string) bool {
	for _, v := range this.nonces {
		if v == nonce {
			return false
		}
	}
	lenNonces := len(this.nonces)

	if lenNonces >= maxNonceSize {
		// 保留最近的3个
		this.nonces[0] = this.nonces[lenNonces-3]
		this.nonces[1] = this.nonces[lenNonces-2]
		this.nonces[2] = this.nonces[lenNonces-1]
		this.nonces = this.nonces[:3]
	}
	this.nonces = append(this.nonces, nonce)
	return true
}

// NonceCache cache for nonce
type NonceCache struct {
	nonces *Cache
}

// NewNonceCache create a *NonceCache struct
func NewNonceCache() *NonceCache {
	return &NonceCache{nonces: NewCache()}
}

// Size get the nonces' size
func (this *NonceCache) Size() int {
	if this != nil && this.nonces != nil {
		return this.nonces.Size()
	}

	return 0
}

// SetNonce set a user's nonce
func (this *NonceCache) SetNonce(userID, nonce string) bool {
	item, _, ok := this.nonces.Get(userID, false)
	if ok && item != nil {
		return item.(*Nonce).Set(nonce)
	}
	this.nonces.Set(userID, &Nonce{[]string{nonce}}, maxNonceTTL)
	return true
}
