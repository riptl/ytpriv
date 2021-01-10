package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/pflag"
	"github.com/terorie/ytwrk/api"
	"github.com/terorie/ytwrk/data"
	"github.com/valyala/fasthttp"
	"go.od2.network/hive/pkg/appctx"
	"go.od2.network/hive/pkg/auth"
	"go.od2.network/hive/pkg/types"
	"go.od2.network/hive/pkg/worker"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// Environment
	logConfig := zap.NewDevelopmentConfig()
	logConfig.DisableStacktrace = true
	logConfig.DisableCaller = true
	logConfig.Level.SetLevel(zapcore.DebugLevel)
	log, err := logConfig.Build()
	if err != nil {
		panic(err)
	}

	// Flags
	token := pflag.String("token", "", "Worker auth token")
	routines := pflag.Uint("routines", 16, "Number of worker routines")
	prefetch := pflag.Uint("prefetch", 256, "Assignment prefetch")
	seedList := pflag.StringSlice("seed", nil, "Seed items")
	pflag.Parse()
	if *routines <= 0 {
		log.Fatal("Invalid routines flag", zap.Uint("flag_routines", *routines))
	}

	// Connect to gRPC API
	ctx := appctx.Context()
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	client, err := grpc.Dial(
		"worker.hive.od2.network:443",
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(&auth.WorkerCredentials{Token: *token}),
	)
	if err != nil {
		log.Fatal("Failed to connect to worker API", zap.Error(err))
	}

	// Construct worker
	assignments := types.NewAssignmentsClient(client)
	discovery := types.NewDiscoveryClient(client)
	simpleWorker := &worker.Simple{
		Assignments:   assignments,
		Log:           log.Named("worker"),
		Handler:       &Handler{
			Log:       log.Named("handler"),
			Discovery: discovery,
		},
		Routines:      *routines,
		Prefetch:      *prefetch,
		GracePeriod:   30 * time.Second,
		FillRate:      3 * time.Second,
		ReportBatch:   128,
		ReportRate:    3 * time.Second,
		StreamBackoff: backoff.WithMaxRetries(backoff.NewConstantBackOff(3 * time.Second), 16),
		APIBackoff:    backoff.WithMaxRetries(backoff.NewConstantBackOff(2 * time.Second), 32),
	}

	// Push seed items.
	seedPointers := make([]*types.ItemPointer, 0, len(*seedList))
	for _, seed := range *seedList {
		compact, err := decodeVideoID(seed)
		if err != nil {
			log.Warn("Ignoring seed ID", zap.String("video_id", seed), zap.Error(err))
			continue
		}
		seedPointers = append(seedPointers, &types.ItemPointer{
			Dst:       &types.ItemLocator{
				Collection: "yt.videos",
				Id:         strconv.FormatInt(compact, 10),
			},
			Timestamp: ptypes.TimestampNow(),
		})
	}
	if len(seedPointers) > 0 {
		if _, err := discovery.ReportDiscovered(ctx, &types.ReportDiscoveredRequest{
			Pointers: seedPointers,
		}); err != nil {
			log.Fatal("Failed to push seed items")
		}
		log.Info("Pushed seed items")
	}

	if err := simpleWorker.Run(ctx); err != nil {
		log.Warn("Worker exited")
	}
}

// Handler discovers videos from YouTube.
type Handler struct {
	Log       *zap.Logger
	Discovery types.DiscoveryClient
}

// WorkAssignment processes a single video.
func (h *Handler) WorkAssignment(ctx context.Context, assign *types.Assignment) types.TaskStatus {
	compactID, err := strconv.ParseInt(assign.Locator.Id, 10, 64)
	if err != nil {
		panic(err)
	}
	videoID := encodeVideoID(compactID)
	h.Log.Info("Got assignment",
		zap.Int32("msg.partition", assign.KafkaPointer.Partition),
		zap.Int64("msg.offset", assign.KafkaPointer.Offset),
		zap.String("item.id", assign.Locator.Id),
		zap.String("video_id", videoID))
	req := api.GrabVideo(videoID)
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		h.Log.Error("Failed to grab video", zap.Error(err))
		return types.TaskStatus_CLIENT_FAILURE
	}
	var v data.Video
	if err := api.ParseVideo(&v, res); err != nil {
		h.Log.Error("Failed to parse video", zap.Error(err))
		return types.TaskStatus_CLIENT_FAILURE
	}
	var discovered []*types.ItemPointer
	for _, rel := range v.RelatedVideos {
		cid, err := decodeVideoID(rel.ID)
		if err != nil {
			h.Log.Error("Discovered weird video ID", zap.Error(err), zap.String("video_id", rel.ID))
			return types.TaskStatus_CLIENT_FAILURE
		}
		discovered = append(discovered, &types.ItemPointer{
			Dst: &types.ItemLocator{
				Collection: "yt.videos",
				Id:         strconv.FormatInt(cid, 10),
			},
			Timestamp: ptypes.TimestampNow(),
		})
	}
	if _, err := h.Discovery.ReportDiscovered(ctx, &types.ReportDiscoveredRequest{
		Pointers: discovered,
	}); err != nil {
		h.Log.Error("Failed to report discovered", zap.Error(err))
	} else {
		h.Log.Info("Reported discovered", zap.Int("discovered_count", len(discovered)))
	}
	return types.TaskStatus_SUCCESS
}

func decodeVideoID(id string) (num int64, err error) {
	if len(id) != 11 {
		return 0, fmt.Errorf("video ID length must be 11")
	}
	var buf [8]byte
	var n int
	n, err = base64.RawURLEncoding.Decode(buf[:], []byte(id))
	if err != nil {
		return 0, err
	} else if n != 8 {
		return 0, fmt.Errorf("decoded length is not 8")
	}
	num = int64(binary.BigEndian.Uint64(buf[:]))
	return
}

func encodeVideoID(num int64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(num))
	return base64.RawURLEncoding.EncodeToString(buf[:])
}
