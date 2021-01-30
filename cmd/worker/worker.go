package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/golang/protobuf/ptypes"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/spf13/pflag"
	yt "github.com/terorie/ytpriv"
	"go.od2.network/hive-api"
	worker_api "go.od2.network/hive-api/worker"
	"go.od2.network/hive-worker"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
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
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
	}()
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	client, err := grpc.Dial(
		"worker.hive.od2.network:443",
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(&worker_api.Credentials{Token: *token}),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(10),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(3*time.Second, 0.8)))),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{}),
	)
	if err != nil {
		log.Fatal("Failed to connect to worker API", zap.Error(err))
	}

	// Construct worker
	assignments := worker_api.NewAssignmentsClient(client)
	discovery := worker_api.NewDiscoveryClient(client)
	simpleWorker := &worker.Simple{
		Collection:  "yt.videos",
		Assignments: assignments,
		Log:         log.Named("worker"),
		Handler: &Handler{
			Log:       log.Named("handler"),
			Discovery: discovery,
		},
		Routines:      *routines,
		Prefetch:      *prefetch,
		GracePeriod:   30 * time.Second,
		FillRate:      3 * time.Second,
		ReportBatch:   128,
		ReportRate:    3 * time.Second,
		StreamBackoff: backoff.NewConstantBackOff(3 * time.Second),
	}

	// Push seed items.
	seedPointers := make([]*hive.ItemPointer, 0, len(*seedList))
	for _, seed := range *seedList {
		compact, err := decodeVideoID(seed)
		if err != nil {
			log.Warn("Ignoring seed ID", zap.String("video_id", seed), zap.Error(err))
			continue
		}
		seedPointers = append(seedPointers, &hive.ItemPointer{
			Dst: &hive.ItemLocator{
				Collection: "yt.videos",
				Id:         strconv.FormatInt(compact, 10),
			},
			Timestamp: ptypes.TimestampNow(),
		})
	}
	if len(seedPointers) > 0 {
		if _, err := discovery.ReportDiscovered(ctx, &worker_api.ReportDiscoveredRequest{
			Pointers: seedPointers,
		}); err != nil {
			log.Fatal("Failed to push seed items", zap.Error(err))
		}
		log.Info("Pushed seed items")
	}

	if err := simpleWorker.Run(ctx); err != nil {
		log.Warn("Worker exited", zap.Error(err))
	}
}

// Handler discovers videos from YouTube.
type Handler struct {
	Client    *yt.Client
	Log       *zap.Logger
	Discovery worker_api.DiscoveryClient
}

// WorkAssignment processes a single video.
func (h *Handler) WorkAssignment(ctx context.Context, assign *hive.Assignment) hive.TaskStatus {
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
	v, err := h.Client.RequestVideo(videoID).Do()
	if err != nil {
		h.Log.Error("Failed to parse video", zap.Error(err))
		return hive.TaskStatus_CLIENT_FAILURE
	}
	var discovered []*hive.ItemPointer
	for _, rel := range v.RelatedVideos {
		cid, err := decodeVideoID(rel.ID)
		if err != nil {
			h.Log.Error("Discovered weird video ID", zap.Error(err), zap.String("video_id", rel.ID))
			return hive.TaskStatus_CLIENT_FAILURE
		}
		discovered = append(discovered, &hive.ItemPointer{
			Dst: &hive.ItemLocator{
				Collection: "yt.videos",
				Id:         strconv.FormatInt(cid, 10),
			},
			Timestamp: ptypes.TimestampNow(),
		})
	}
	if _, err := h.Discovery.ReportDiscovered(ctx, &worker_api.ReportDiscoveredRequest{
		Pointers: discovered,
	}); err != nil {
		h.Log.Error("Failed to report discovered", zap.Error(err))
	} else {
		h.Log.Info("Reported discovered", zap.Int("discovered_count", len(discovered)))
	}
	return hive.TaskStatus_SUCCESS
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
