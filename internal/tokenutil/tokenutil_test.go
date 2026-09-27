package tokenutil

import (
	"testing"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const testSecret = "test-access-secret-that-is-long-enough-123456"

func TestAccessTokenRoundTrip(t *testing.T) {
	user := &domain.User{ID: primitive.NewObjectID(), Name: "Test", Admin: domain.VerifiedAdmin}
	token, err := CreateAccessToken(user, testSecret, 1)
	require.NoError(t, err)

	id, err := ExtractIDFromToken(token, testSecret)
	require.NoError(t, err)
	require.Equal(t, user.ID.Hex(), id)

	role, err := ExtractRoleFromToken(token, testSecret)
	require.NoError(t, err)
	require.Equal(t, string(domain.VerifiedAdmin), role)
}

func TestMalformedClaimsReturnError(t *testing.T) {
	require.NotPanics(t, func() {
		_, err := ExtractRoleFromToken("eyJhbGciOiJIUzI1NiJ9.eyJpZCI6MX0.invalid", testSecret)
		require.Error(t, err)
	})
}

func TestRefreshTokenUsesRefreshClaims(t *testing.T) {
	user := &domain.User{ID: primitive.NewObjectID(), Name: "Test", Admin: domain.SuperAdmin}
	token, err := CreateRefreshToken(user, testSecret, 1)
	require.NoError(t, err)

	id, err := ExtractIDFromRefreshToken(token, testSecret)
	require.NoError(t, err)
	require.Equal(t, user.ID.Hex(), id)
}
