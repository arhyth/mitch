package internal_test

import (
	"testing"

	"github.com/arhyth/mitch/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindUnappliedVersions(t *testing.T) {
	as := assert.New(t)
	reqrd := require.New(t)

	inDB := []internal.Version{
		{ID: 1, ContentHash: "c93ec757118fa37d546ace51a5af5340b86589b2c9754d2a10511e3e0a2d9476"},
		{ID: 3, ContentHash: "605bcd56fc7890ac97975811ee7061aec97e4481014eecdb8a19d7e533131a43"},
		{ID: 4, ContentHash: "438194a503eb004cf15019a4fe4511b8d2d2d24633a8842dcc2802fa3bc63b26"},
		{ID: 5, ContentHash: "d365a958f4479c749c6d77e5f2a94617894fc516411da5ac3c019ad4c52f0fd6"},
		{ID: 7, ContentHash: "79845a9365de80a1cef9128ae46409e5737aef0009634725e4f006fa8f68caf9"}, // <-- max version_id
	}
	inFS := []internal.Version{
		{ID: 1, ContentHash: "c93ec757118fa37d546ace51a5af5340b86589b2c9754d2a10511e3e0a2d9476"},
		{ID: 2, ContentHash: "e9f661e94cf209c915010dbcfb40cdf08e60c429f6a9bcfef65b68ad1c3e082a"}, // missing
		{ID: 3, ContentHash: "605bcd56fc7890ac97975811ee7061aec97e4481014eecdb8a19d7e533131a43"},
		{ID: 4, ContentHash: "438194a503eb004cf15019a4fe4511b8d2d2d24633a8842dcc2802fa3bc63b26"},
		{ID: 5, ContentHash: "d365a958f4479c749c6d77e5f2a94617894fc516411da5ac3c019ad4c52f0fd6"},
		{ID: 6, ContentHash: "bde0f69b56723a35bdeac9eb3ba55cf473ac07caab753742aaea498128cb7d50"}, // missing
		{ID: 7, ContentHash: "79845a9365de80a1cef9128ae46409e5737aef0009634725e4f006fa8f68caf9"}, // max version_id
		{ID: 8, ContentHash: "498453a98efe80ef5e2c451c0601b335f91c8baa570dda4af7917d8d7ef2a2bd"}, // new migration
	}
	unapplied, hasMissing := internal.FindUnappliedVersions(inDB, inFS)
	as.True(hasMissing)
	reqrd.NotEmpty(unapplied)
	as.Equal(1, len(unapplied))
	as.Equal(inFS[7].ID, unapplied[0].ID)
}
