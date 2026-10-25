package user_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

func TestEmailAddress_New_RemovesTheSpacesAndChangesTheLettersToLowercase(t *testing.T) {
	t.Parallel()

	got, err := domain.NewEmailAddress("  Ann.Smith@Example.COM \t")

	require.NoError(t, err)
	assert.Equal(t, domain.EmailAddress("ann.smith@example.com"), got)
}

func TestEmailAddress_New_RejectsAnAddressThatIsNotValid(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"empty":                  "",
		"only spaces":            "   ",
		"no at sign":             "ann.example.com",
		"no local part":          "@example.com",
		"no domain":              "ann@",
		"domain without a dot":   "ann@example",
		"two at signs":           "ann@@example.com",
		"space in the address":   "an n@example.com",
		"display name included":  "Ann <ann@example.com>",
		"more than 254 chars":    strings.Repeat("a", 250) + "@example.com",
		"non-ASCII local part":   "annä@example.com",
		"non-ASCII domain":       "ann@exämple.com",
		"zero-width character":   "ann@example.com\u200b",
		"Kelvin sign":            "ann@example\u212a.com",
		"NUL character":          "an\x00n@example.com",
		"domain literal":         "ann@[127.0.0.1]",
		"label with only dash":   "ann@-.-",
		"label starts with dash": "ann@-example.com",
		"label ends with dash":   "ann@example-.com",
		"empty label":            "ann@example..com",
		"underscore in domain":   "ann@exa_mple.com",
		"label over 63 chars":    "ann@" + strings.Repeat("a", 64) + ".com",
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewEmailAddress(input)

			require.ErrorIs(t, err, domain.ErrInvalidEmailAddress)
		})
	}
}

func TestEmailAddress_New_AcceptsCommonAddresses(t *testing.T) {
	t.Parallel()
	tests := []string{"ann@example.com", "ann+tag@sub.example.co.uk", "a.b-c_d@example.com", "ann@xn--exmple-cua.com"}
	for _, input := range tests {
		got, err := domain.NewEmailAddress(input)

		require.NoError(t, err, input)
		assert.Equal(t, domain.EmailAddress(input), got)
	}
}

func TestDisplayName_New_RemovesTheSpacesAndKeepsTheLetters(t *testing.T) {
	t.Parallel()

	got, err := domain.NewDisplayName("  Zoë 李  ")

	require.NoError(t, err)
	assert.Equal(t, domain.DisplayName("Zoë 李"), got)
}

func TestDisplayName_New_RejectsAnEmptyName(t *testing.T) {
	t.Parallel()

	// The last 2 names have only a zero-width space, and only a word joiner. A person cannot see them.
	for _, input := range []string{"", "   ", "\t\n", "\u200b", " \u2060\u200b "} {
		_, err := domain.NewDisplayName(input)

		require.ErrorIs(t, err, domain.ErrEmptyDisplayName)
	}
}

func TestDisplayName_New_RejectsAControlCharacter(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"a\x00b", "line one\nline two", "tab\there", "bell\a", "\x7f"} {
		_, err := domain.NewDisplayName(input)

		require.ErrorIs(t, err, domain.ErrInvalidDisplayName, input)
	}
}

func TestDisplayName_New_AcceptsAnEmojiWithAZeroWidthJoiner(t *testing.T) {
	t.Parallel()

	got, err := domain.NewDisplayName("👩\u200d💻")

	require.NoError(t, err)
	assert.Equal(t, domain.DisplayName("👩\u200d💻"), got)
}

func TestDisplayName_New_AcceptsExactly100CharactersAndRejectsMore(t *testing.T) {
	t.Parallel()

	_, errAtLimit := domain.NewDisplayName(strings.Repeat("é", domain.MaxDisplayNameLength))
	_, errOverLimit := domain.NewDisplayName(strings.Repeat("é", domain.MaxDisplayNameLength+1))

	require.NoError(t, errAtLimit)
	require.ErrorIs(t, errOverLimit, domain.ErrDisplayNameTooLong)
}

func TestUser_New_MakesAUserWithAUUIDv7AndTheTimeInUTC(t *testing.T) {
	t.Parallel()
	local := time.FixedZone("UTC+6", 6*60*60)
	now := time.Date(2026, 10, 25, 16, 0, 0, 0, local)

	got, err := domain.NewUser(" Ann@Example.com ", " Ann ", now)

	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), got.ID.Version())
	assert.Equal(t, domain.EmailAddress("ann@example.com"), got.EmailAddress)
	assert.Equal(t, domain.DisplayName("Ann"), got.DisplayName)
	assert.Equal(t, time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC), got.CreatedAt)
	assert.Equal(t, time.UTC, got.CreatedAt.Location())
}

func TestUser_New_MakesADifferentIDForEachUser(t *testing.T) {
	t.Parallel()
	now := time.Now()

	first, err := domain.NewUser("a@example.com", "A", now)
	require.NoError(t, err)
	second, err := domain.NewUser("b@example.com", "B", now)
	require.NoError(t, err)

	assert.NotEqual(t, first.ID, second.ID)
}

func TestUser_New_ReturnsTheErrorOfTheRuleThatTheInputBreaks(t *testing.T) {
	t.Parallel()
	now := time.Now()

	_, emailErr := domain.NewUser("not-an-address", "Ann", now)
	_, nameErr := domain.NewUser("ann@example.com", " ", now)

	require.ErrorIs(t, emailErr, domain.ErrInvalidEmailAddress)
	require.ErrorIs(t, nameErr, domain.ErrEmptyDisplayName)
}
