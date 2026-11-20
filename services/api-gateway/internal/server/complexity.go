package server

import "api-gateway/graph/generated"

// complexity returns gqlgen complexity functions that multiply a list
// field's cost by the number of items it can return. Without this, every field
// costs 1 regardless of `limit`, so a query like
//
//	user(id: "a") { posts(limit: 100) { author { posts(limit: 100) { author { ... } } } } }
//
// looks cheap while fanning out 100x per level. Pair it with
// extension.FixedComplexityLimit in the server.
func complexity() generated.ComplexityRoot {
	var c generated.ComplexityRoot

	// Nested lists with a `limit` argument.
	c.User.Posts = func(child int, limit *int32) int { return listCost(child, limit, 10) }
	c.Post.Comments = func(child int, limit *int32) int { return listCost(child, limit, 10) }

	// Top-level lists.
	c.Query.Posts = func(child int, limit *int32) int { return listCost(child, limit, 10) }
	c.Query.Feed = func(child int, _ string, limit *int32, _ *string) int { return listCost(child, limit, 20) }
	c.Query.Notifications = func(child int, _ string, limit *int32) int { return listCost(child, limit, 20) }

	// Lists whose size is the number of IDs requested.
	c.Query.Users = func(child int, ids []string) int { return countCost(child, len(ids)) }
	c.Query.Profiles = func(child int, ids []string) int { return countCost(child, len(ids)) }
	c.Query.PostsByIds = func(child int, ids []string) int { return countCost(child, len(ids)) }

	return c
}

func listCost(child int, limit *int32, def int) int {
	n := def
	if limit != nil {
		n = int(*limit)
	}
	return countCost(child, n)
}

// countCost is n copies of the child selection, plus 1 for the field itself.
// Out-of-range limits are rejected by the resolver; here they only need to be
// non-negative so a hostile value cannot produce a negative (cheap) cost.
func countCost(child, n int) int {
	if n < 1 {
		n = 1
	}
	return 1 + n*child
}
