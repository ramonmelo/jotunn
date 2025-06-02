package tracker

import "github.com/LinharesAron/jotunn/internal/types"

func FilterUnseen(t Tracker, users, passwords []string) []types.Attempt {
	var unseen []types.Attempt

	for _, user := range users {
		for _, pass := range passwords {
			if !t.HasSeen(user, pass) {
				unseen = append(unseen, types.Attempt{
					Username: user,
					Password: pass,
				})
			}
		}
	}

	return unseen
}
