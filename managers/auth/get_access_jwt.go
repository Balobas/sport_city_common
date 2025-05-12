package authManager

func (cm *ClientsAuthManager) GetAccessToken() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.accessJwt
}
