package sqs

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/osmosis-labs/sqs/sqsdomain"
	"github.com/osmosis-labs/sqs/sqsdomain/json"
	prototypes "github.com/osmosis-labs/sqs/sqsdomain/proto/types"

	"github.com/osmosis-labs/osmosis/v23/ingest"
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v23/x/poolmanager/types"
)

const sqsIngesterName = "sidecar-query-server"

var _ ingest.Ingester = &sqsIngester{}

const (
	// TODO: find optimal size but 100 is probably good to start
	poolChunkSize = 100

	// 10 KB
	maxCallMsgSize = 10 * 1024 * 1024
	// 5 KB
	payloadThreshold = maxCallMsgSize / 2

	resetBlockInterval = 30

	synchingBlockInterval = 15
)

// sqsIngester is a sidecar query server (SQS) implementation of Ingester.
// It encapsulates all individual SQS ingesters.
type sqsIngester struct {
	poolsIngester domain.AtomicIngester
	host          string
	port          string
	grpcConn      *grpc.ClientConn
	appCodec      codec.Codec
}

// NewSidecarQueryServerIngester creates a new sidecar query server ingester.
// poolsRepository is the storage for pools.
// gammKeeper is the keeper for Gamm pools.
func NewSidecarQueryServerIngester(poolsIngester domain.AtomicIngester, host, port string, appCodec codec.Codec) ingest.Ingester {
	return &sqsIngester{
		poolsIngester: poolsIngester,
		appCodec:      appCodec,
	}
}

type IngestProcessBlockArgs struct {
	Pools []sqsdomain.PoolI
}

// ProcessBlock implements ingest.Ingester.
func (i *sqsIngester) ProcessBlock(ctx sdk.Context) (err error) {
	// TODO: track how long this takes

	if i.grpcConn == nil {
		// TODO: move to config
		i.grpcConn, err = grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxCallMsgSize)))
		if err != nil {
			return err
		}
	}

	ingesterClient := prototypes.NewIngesterClient(i.grpcConn)

	defer func() {
		if err != nil {

			// This may occur if we start processing a block and then the connection is lost or error occurs.
			// As a result, sqs may be in the middle of processing a block and we need to reset the connection.
			// However, we do not want to reset the connection every block since during syncing, the node might process blocks much faster than
			// the sqs can process them. As a result, we reset the connection every resetBlockInterval blocks.
			if strings.Contains(err.Error(), "transaction is already in progress") && i.grpcConn != nil && ctx.BlockHeight()%resetBlockInterval == 0 {
				if _, resetErr := ingesterClient.ResetBlockProcessing(sdk.UnwrapSDKContext(ctx), &prototypes.ResetBlockProcessingRequest{}); err != nil {
					err = fmt.Errorf("failed to reset block processing: (%w), original error (%w) ", resetErr, err)
					// TODO: consider adding telemetry here
				}
			}
		}
	}()

	// // Process block by reading and writing data and ingesting data into sinks
	pools, takerFeeMap, err := i.poolsIngester.ProcessBlock(ctx)
	if err != nil {
		return err
	}

	// Serialize taker fee map
	takerFeeMapBz, err := takerFeeMap.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = ingesterClient.ProcessChainTakerFees(sdk.UnwrapSDKContext(ctx), &prototypes.ProcessTakerFeesRequest{
		TakerFeesMap: takerFeeMapBz,
	})
	if err != nil {
		return err
	}

	proceccChainPoolsClient, err := ingesterClient.ProcessChainPools(ctx)
	if err != nil {
		return err
	}

	chunk := &prototypes.ChainPoolsDataChunk{
		Pools: make([]*prototypes.PoolData, 0),
	}

	byteCount := 0
	for j := 0; j < len(pools); j++ {

		// Serialize chain pool
		chainPoolBz, err := i.appCodec.MarshalInterfaceJSON(pools[j].GetUnderlyingPool())
		if err != nil {
			return err
		}

		byteCount += len(chainPoolBz)

		// Serialize sqs pool model
		sqsPoolBz, err := json.Marshal(pools[j].GetSQSPoolModel())
		if err != nil {
			return err
		}

		byteCount += len(sqsPoolBz)

		var tickModelBz []byte
		if pools[j].GetType() == poolmanagertypes.Concentrated {
			tickModel, err := pools[j].GetTickModel()
			if err != nil {
				return err
			}

			tickModelBz, err = json.Marshal(tickModel)
			if err != nil {
				return err
			}

			byteCount += len(tickModelBz)
		}

		chunk.Pools = append(chunk.Pools, &prototypes.PoolData{
			ChainModel: chainPoolBz,
			SqsModel:   sqsPoolBz,
			TickModel:  tickModelBz,
		})

		shouldSendChunk := byteCount > payloadThreshold || j == len(pools)-1
		if shouldSendChunk {
			if err := proceccChainPoolsClient.Send(chunk); err != nil {
				return err
			}

			byteCount = 0

			// Initialize new chunk
			chunk = &prototypes.ChainPoolsDataChunk{
				Pools: make([]*prototypes.PoolData, 0),
			}
		}
	}

	if err := proceccChainPoolsClient.CloseSend(); err != nil {
		return err
	}

	if _, err := ingesterClient.EndBlockProcessing(sdk.UnwrapSDKContext(ctx), &prototypes.EndBlockProcessingRequest{
		BlockHeight: uint64(ctx.BlockHeight()),
	}); err != nil {
		return err
	}

	return nil
}

// GetName implements ingest.Ingester.
func (*sqsIngester) GetName() string {
	return sqsIngesterName
}
