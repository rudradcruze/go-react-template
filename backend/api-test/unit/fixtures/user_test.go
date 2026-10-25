package fixtures_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/rudradcruze/go-react-template/backend/api-test/fixtures"
	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

func TestFixtures_User_MakesValidUsersWithUniqueEmailAddresses(t *testing.T) {
	t.Parallel()

	first := fixtures.User()
	second := fixtures.User()

	assert.NotEqual(t, first.EmailAddress, second.EmailAddress)
	assert.NotEqual(t, first.ID, second.ID)
	for _, user := range []domain.User{first, second} {
		_, emailErr := domain.NewEmailAddress(string(user.EmailAddress))
		_, nameErr := domain.NewDisplayName(string(user.DisplayName))
		assert.NoError(t, emailErr)
		assert.NoError(t, nameErr)
		assert.Equal(t, uuid.Version(7), user.ID.Version())
	}
}

func TestFixtures_User_UsesTheGivenValues(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	got := fixtures.User(
		fixtures.WithEmailAddress("Ann@Example.com"),
		fixtures.WithDisplayName("Ann"),
		fixtures.WithCreatedAt(createdAt),
	)

	assert.Equal(t, domain.EmailAddress("ann@example.com"), got.EmailAddress)
	assert.Equal(t, domain.DisplayName("Ann"), got.DisplayName)
	assert.Equal(t, createdAt, got.CreatedAt)
}

func TestFixtures_User_PanicsWhenAGivenValueIsNotValid(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() { fixtures.User(fixtures.WithEmailAddress("not-an-address")) })
}
