package infobip

import (
	"context"
	"encoding/json"
	"fmt"
	"infobip-go-client/pkg/infobip/models"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactValidReq(t *testing.T) {
	apiKey := "secret"
	msg := models.ContactMessage{
		MessageCommon: models.MessageCommon{
			From:         "16175551213",
			To:           "16175551212",
			MessageID:    "a28dd97c-1ffb-4fcf-99f1-0b557ed381da",
			CallbackData: "some data",
			NotifyURL:    "https://www.google.com",
		},
		Content: models.ContactContent{
			Contacts: []models.Contact{{Name: models.ContactName{FirstName: "John", FormattedName: "Mr. John Smith"}}},
		},
	}
	rawJSONResp := []byte(`{
		"to": "441134960001",
		"messageCount": 1,
		"messageId": "a28dd97c-1ffb-4fcf-99f1-0b557ed381da",
		"status": {
			"groupId": 1,
			"groupName": "PENDING",
			"id": 7,
			"name": "PENDING_ENROUTE",
			"description": "Message sent to next instance"
		}
	}`)
	var expectedResp models.MessageResponse
	err := json.Unmarshal(rawJSONResp, &expectedResp)
	require.Nil(t, err)

	serv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasSuffix(r.URL.Path, sendContactPath))
		assert.Equal(t, fmt.Sprintf("App %s", apiKey), r.Header.Get("Authorization"))
		parsedBody, servErr := ioutil.ReadAll(r.Body)
		assert.Nil(t, servErr)

		var receivedMsg models.ContactMessage
		servErr = json.Unmarshal(parsedBody, &receivedMsg)
		assert.Nil(t, servErr)
		assert.Equal(t, receivedMsg, msg)

		_, servErr = w.Write(rawJSONResp)
		assert.Nil(t, servErr)
	}))
	defer serv.Close()

	host := serv.URL
	whatsApp := whatsAppChannel{reqHandler: httpHandler{
		httpClient: http.Client{},
		baseURL:    host,
		apiKey:     apiKey,
	}}
	messageResponse, respDetails, err := whatsApp.SendContactMessage(context.Background(), msg)

	require.Nil(t, err)
	assert.NotEqual(t, models.MessageResponse{}, messageResponse)
	assert.Equal(t, expectedResp, messageResponse)
	require.Nil(t, err)
	assert.NotNil(t, respDetails)
	assert.Equal(t, http.StatusOK, respDetails.HTTPResponse.StatusCode)
	assert.Equal(t, models.ErrorDetails{}, respDetails.ErrorResponse)
}

func TestInvalidContactMsg(t *testing.T) {
	apiKey := "secret"
	whatsApp := whatsAppChannel{reqHandler: httpHandler{
		httpClient: http.Client{},
		baseURL:    "https://something.api.infobip.com",
		apiKey:     apiKey,
	}}
	msg := models.ContactMessage{
		MessageCommon: models.MessageCommon{
			From:         "16175551213",
			To:           "16175551212",
			MessageID:    "a28dd97c-1ffb-4fcf-99f1-0b557ed381da",
			CallbackData: "some data",
			NotifyURL:    "https://www.google.com",
		},
		Content: models.ContactContent{
			Contacts: []models.Contact{{Name: models.ContactName{FormattedName: "Mr. John Smith"}}},
		},
	}

	messageResponse, respDetails, err := whatsApp.SendContactMessage(context.Background(), msg)
	require.NotNil(t, err)
	assert.IsType(t, err, validator.ValidationErrors{})
	assert.Equal(t, models.MessageResponse{}, messageResponse)
	assert.Equal(t, models.ResponseDetails{}, respDetails)
}

func TestContact4xxErrors(t *testing.T) {
	tests := []struct {
		rawJSONResp []byte
		statusCode  int
	}{
		{
			rawJSONResp: []byte(`{
				"requestError": {
					"serviceException": {
						"messageId": "BAD_REQUEST",
						"text": "Bad request",
						"validationErrors": {
							"content.contacts[0].birthday": [
								"must be in the YYYY-MM-DD format"
							]
						}
					}
				}
			}`),
			statusCode: http.StatusBadRequest,
		},
		{
			rawJSONResp: []byte(`{
				"requestError": {
					"serviceException": {
						"messageId": "UNAUTHORIZED",
						"text": "Invalid login details"
					}
				}
			}`),
			statusCode: http.StatusUnauthorized,
		},
		{
			rawJSONResp: []byte(`{
				"requestError": {
					"serviceException": {
						"messageId": "TOO_MANY_REQUESTS",
						"text": "Too many requests"
					}
				}
			}`),
			statusCode: http.StatusTooManyRequests,
		},
	}
	apiKey := "secret"
	msg := models.ContactMessage{
		MessageCommon: models.MessageCommon{
			From:         "16175551213",
			To:           "16175551212",
			MessageID:    "a28dd97c-1ffb-4fcf-99f1-0b557ed381da",
			CallbackData: "some data",
			NotifyURL:    "https://www.google.com",
		},
		Content: models.ContactContent{
			Contacts: []models.Contact{{Name: models.ContactName{FirstName: "John", FormattedName: "Mr. John Smith"}}},
		},
	}

	for _, tc := range tests {
		t.Run(strconv.Itoa(tc.statusCode), func(t *testing.T) {
			var expectedResp models.ErrorDetails
			err := json.Unmarshal(tc.rawJSONResp, &expectedResp)
			require.Nil(t, err)
			serv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, servErr := w.Write(tc.rawJSONResp)
				assert.Nil(t, servErr)
			}))

			host := serv.URL
			whatsApp := whatsAppChannel{reqHandler: httpHandler{
				httpClient: http.Client{},
				baseURL:    host,
				apiKey:     apiKey,
			}}
			messageResponse, respDetails, err := whatsApp.SendContactMessage(context.Background(), msg)
			serv.Close()

			require.Nil(t, err)
			assert.NotEqual(t, http.Response{}, respDetails.HTTPResponse)
			assert.NotEqual(t, models.ErrorDetails{}, respDetails.ErrorResponse)
			assert.Equal(t, expectedResp, respDetails.ErrorResponse)
			assert.Equal(t, tc.statusCode, respDetails.HTTPResponse.StatusCode)
			assert.Equal(t, models.MessageResponse{}, messageResponse)
		})
	}
}