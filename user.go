package main

// Write DN and User ID to Cache
func writeUserToCache(DN string, ID string) {
	logger(1, "Writting User to DN Cache: "+ID, false)
	_, found := HornbillCache.DN[DN]
	if !found {
		HornbillCache.DN[DN] = ID
	}
}

// Get User ID From Cache By DN
func getUserFromDNCache(DN string) string {
	_, found := HornbillCache.DN[DN]
	if found {
		return HornbillCache.DN[DN]
	}
	return ""
}
