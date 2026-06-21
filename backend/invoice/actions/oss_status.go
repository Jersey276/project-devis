package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
	usersGrpc "project-devis-invoice/services/usersgrpc"
)

// GetOSSThresholdStatus reports where the issuer stands against the OSS
// distance-selling threshold (art. 259 D CGI) for the current civil year. The
// cumulative figure is the net B2C intra-EU distance-sale turnover (issued
// invoices minus credit notes carrying the frozen assiette flag). OSS is active
// when the issuer opted in OR the threshold has already been reached.
func (s *Server) GetOSSThresholdStatus(ctx context.Context, req *invoiceGrpc.GetOSSThresholdStatusRequest) (resp *invoiceGrpc.GetOSSThresholdStatusResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("get_oss_threshold_status", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GetOSSThresholdStatusResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	now := time.Now()
	cumulative, err := s.ossCumulativeHTForYear(ctx, req.UserId, "", now)
	if err != nil {
		return &invoiceGrpc.GetOSSThresholdStatusResponse{Success: false, Code: codes.InternalError}, err
	}

	userResp, err := s.usersClient.GetUser(ctx, &usersGrpc.GetUserRequest{UserId: req.UserId})
	if err != nil {
		return &invoiceGrpc.GetOSSThresholdStatusResponse{Success: false, Code: codes.InternalError}, fmt.Errorf("get user: %w", err)
	}
	if !userResp.GetSuccess() || userResp.GetUser() == nil {
		return &invoiceGrpc.GetOSSThresholdStatusResponse{Success: false, Code: codes.DependencyMissing}, nil
	}

	ossEnabled := userResp.GetUser().GetOssEnabled()
	ossActive := ossEnabled || cumulative >= ossThresholdCents

	return &invoiceGrpc.GetOSSThresholdStatusResponse{
		Success:           true,
		Code:              codes.Success,
		Year:              int32(now.In(invoiceTZ).Year()),
		CumulativeHtCents: cumulative,
		ThresholdCents:    ossThresholdCents,
		OssEnabled:        ossEnabled,
		OssActive:         ossActive,
	}, nil
}
