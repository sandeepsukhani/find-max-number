package grpc_test

import (
	fmnmock "github.com/sandlis/find-max-number/client/grpc/mock_findMaxNumber"
	rgpb "github.com/sandlis/find-max-number/proto"
	"testing"
	"github.com/golang/mock/gomock"
	"time"
	"context"
	"github.com/golang/protobuf/proto"
	"fmt"
)

var req = &rgpb.FindMaxNumberRequest{Number:5}
var res = &rgpb.FindMaxNumberResponse{MaxNumber:5}

func TestFindMaxNumberMockClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock for the stream returned by FindMaxNumberClient
	stream := fmnmock.NewMockNumbers_FindMaxNumberClient(ctrl)
	// set expectation on sending.
	stream.EXPECT().Send(
		req,
	).Return(nil)
	// Set expectation on receiving.
	stream.EXPECT().Recv().Return(res, nil)

	// Create mock for the client interface.
	rgclient := fmnmock.NewMockNumbersClient(ctrl)
	// Set expectation on FindMaxNumber
	rgclient.EXPECT().FindMaxNumber(
		gomock.Any(),
	).Return(stream, nil)
	if err := testFindMaxNumber(rgclient); err != nil {
		t.Fatalf("Test failed: %v", err)
	}
}


func testFindMaxNumber(client rgpb.NumbersClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.FindMaxNumber(ctx)
	if err != nil {
		return err
	}
	if err := stream.Send(req); err != nil {
		return err
	}

	got, err := stream.Recv()
	if err != nil {
		return err
	}
	if !proto.Equal(got, res) {
		return fmt.Errorf("stream.Recv() = %v, want %v", got, res)
	}
	return nil
}