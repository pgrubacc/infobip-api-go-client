package mms

import (
	"context"

	"github.com/infobip-community/infobip-api-go-sdk/internal"
	"github.com/infobip-community/infobip-api-go-sdk/pkg/infobip/models"
)

// MMS provides methods to interact with the Infobip MMS API.
// MMS API docs: https://www.infobip.com/docs/api#channels/mms
type MMS interface {
	SendMsg(context.Context, models.MMSMsg) (models.MMSResponse, models.ResponseDetails, error)
	GetOutboundMsgDeliveryReports(ctx context.Context, opts models.OutboundMsgDeliveryReportsOpts) (
		models.OutboundMMSDeliveryReportsResponse, models.ResponseDetails, error)
}

type Channel struct {
	ReqHandler internal.HTTPHandler
}

const sendMessagePath = "/mms/1/single"
const getOutboundMsgDeliveryReportsPath = "/mms/1/reports"

func (mms *Channel) SendMsg(
	ctx context.Context,
	msg models.MMSMsg,
) (msgResp models.MMSResponse, respDetails models.ResponseDetails, err error) {
	respDetails, err = mms.ReqHandler.PostMultipartReq(ctx, &msg, &msgResp, sendMessagePath)
	return msgResp, respDetails, err
}

func (mms *Channel) GetOutboundMsgDeliveryReports(
	ctx context.Context,
	opts models.OutboundMsgDeliveryReportsOpts,
) (msgResp models.OutboundMMSDeliveryReportsResponse, respDetails models.ResponseDetails, err error) {
	params := map[string]string{"bulkId": opts.BulkID, "messageId": opts.MessageID, "limit": opts.Limit}
	respDetails, err = mms.ReqHandler.GetRequest(ctx, &msgResp, getOutboundMsgDeliveryReportsPath, params)
	return msgResp, respDetails, err
}