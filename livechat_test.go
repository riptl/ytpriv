package yt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terorie/ytpriv/types"
	"github.com/valyala/fastjson"
)

// FYI don't agree with the shit in this message, this just happened to be the one I grabbed while debugging.
func TestParseLivechatPage(t *testing.T) {
	const examplePaidMsg = `{"clickTrackingParams":"CAEQl98BIhMIko7U3Iis7gIVRifgCh3L9QG1","addChatItemAction":{"item":{"liveChatPaidMessageRenderer":{"id":"ChwKGkNOS0VudGFJck80Q0ZZbUl3UW9kTlNvTC1n","timestampUsec":"1611199191038289","authorName":{"simpleText":"Garrett Raleigh"},"authorPhoto":{"thumbnails":[{"url":"https://yt4.ggpht.com/ytc/AAUvwniwP55X_gi1Y6kYE0JLKOmmmR_kpV1GWLzCSQ=s32-c-k-c0x00ffffff-no-rj","width":32,"height":32},{"url":"https://yt4.ggpht.com/ytc/AAUvwniwP55X_gi1Y6kYE0JLKOmmmR_kpV1GWLzCSQ=s64-c-k-c0x00ffffff-no-rj","width":64,"height":64}]},"purchaseAmountText":{"simpleText":"$5.00"},"message":{"runs":[{"text":"WHAT ABOUT HARRIS BEING PICKED AS VP ONLY BECAUSE SHE'S A WOMAN AND A MINORITY"}]},"headerBackgroundColor":4278239141,"headerTextColor":4278190080,"bodyBackgroundColor":4280150454,"bodyTextColor":4278190080,"authorExternalChannelId":"UC1AmEt6UQn66GFvR6XlPYHA","authorNameTextColor":2315255808,"contextMenuEndpoint":{"clickTrackingParams":"CAYQ7rsEIhMIko7U3Iis7gIVRifgCh3L9QG1","commandMetadata":{"webCommandMetadata":{"ignoreNavigation":true}},"liveChatItemContextMenuEndpoint":{"params":"Q2g0S0hBb2FRMDVMUlc1MFlVbHlUelJEUmxsdFNYZFJiMlJPVTI5TUxXY1FBQm80Q2cwS0N6ZE5ObXBVWkRkdldFUjNLaWNLR0ZWRFRIZE9WRmhYUldwV1pESnhTVWhNWTFoNFVWZDRRUklMTjAwMmFsUmtOMjlZUkhjZ0FTZ0VNaG9LR0ZWRE1VRnRSWFEyVlZGdU5qWkhSblpTTmxoc1VGbElRUSUzRCUzRA=="}},"timestampColor":2147483648,"contextMenuAccessibility":{"accessibilityData":{"label":"Comment actions"}},"trackingParams":"CAYQ7rsEIhMIko7U3Iis7gIVRifgCh3L9QG1"}}}}`
	value, err := fastjson.Parse(examplePaidMsg)
	require.NoError(t, err)
	require.NotNil(t, value)
	msg := parseLivechatMessage(value)
	require.NotNil(t, msg)
	assert.Equal(t, examplePaidMsg[600:691], string(msg.Message))
	msg.Message = nil
	assert.Equal(t, &types.LivechatMessage{
		ID:         "ChwKGkNOS0VudGFJck80Q0ZZbUl3UW9kTlNvTC1n",
		AuthorID:   "UC1AmEt6UQn66GFvR6XlPYHA",
		Author:     "Garrett Raleigh",
		Timestamp:  1611199191038289,
		SuperChat:  true,
		PaidAmount: "$5.00",
	}, msg)
}
