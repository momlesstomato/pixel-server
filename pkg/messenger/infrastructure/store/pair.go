package store

// canonicalPair orders a user pair as (min,max) for stable single-row friendship storage.
func canonicalPair(userOneID, userTwoID int) (int, int) {
	if userOneID <= userTwoID {
		return userOneID, userTwoID
	}
	return userTwoID, userOneID
}
