package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/pflag"
	"github.com/terorie/ytwrk/api"
	"github.com/terorie/ytwrk/data"
	"github.com/valyala/fasthttp"
	"go.od2.network/hive/pkg/auth"
	"go.od2.network/hive/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.DisableStacktrace = true
	logConfig.DisableCaller = true
	log, err := logConfig.Build()
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	// TODO remove token
	token := pflag.String("token", "HCyp15tCAAhXAZlV1WiHSy7ADxenGX_CXRq9dvnAzdT30", "Worker auth token")
	client, err := grpc.Dial(
		"worker.hive.od2.network:443",
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(&auth.WorkerCredentials{Token: *token}),
	)
	if err != nil {
		log.Fatal("Failed to connect to worker API", zap.Error(err))
	}
	assignments := types.NewAssignmentsClient(client)
	discovery := types.NewDiscoveryClient(client)

	ctx := context.Background()
	/*_, err = discovery.ReportDiscovered(ctx, &types.ReportDiscoveredRequest{
		Pointers: []*types.ItemPointer{
			{
				Dst: &types.ItemLocator{
					Collection: "yt.videos",
					//Id:         strconv.FormatInt(mustDecodeVideoID("YPiOWJDdChM"), 10),
					Id:         strconv.FormatInt(mustDecodeVideoID("1uei2iNwyOo"), 10),
					//Id:         strconv.FormatInt(mustDecodeVideoID("h2f0vqgYdLc"), 10),
					//Id:         strconv.FormatInt(mustDecodeVideoID("h2f0vqgYdLc"), 10),
				},
				Timestamp: ptypes.TimestampNow(),
			},
		},
	})*/
	if err != nil {
		log.Error("Failed to discover video", zap.Error(err))
	}
	log.Info("Inserted seed item")
	stream, err := assignments.OpenAssignmentsStream(ctx, &types.OpenAssignmentsStreamRequest{})
	if err != nil {
		log.Fatal("Failed to open assignments stream", zap.Error(err))
		return
	}
	log.Info("Opened stream", zap.Int64("stream_id", stream.StreamId))
	defer func() {
		_, err = assignments.CloseAssignmentsStream(ctx, &types.CloseAssignmentsStreamRequest{
			StreamId: stream.StreamId,
		})
		if err != nil {
			log.Error("Failed to close stream", zap.Error(err))
		}
		log.Info("Closed stream")
	}()
	for {
		wanted, err := assignments.WantAssignments(ctx, &types.WantAssignmentsRequest{
			StreamId:     stream.StreamId,
			AddWatermark: 1,
		})
		if err != nil {
			log.Fatal("Failed to request assignments", zap.Error(err))
		}
		log.Info("Requested assignments",
			zap.Int32("watermark", wanted.Watermark),
			zap.Int32("added_watermark", wanted.AddedWatermark))
		assigns, err := assignments.StreamAssignments(ctx, &types.StreamAssignmentsRequest{StreamId: stream.StreamId})
		if err != nil {
			log.Fatal("Connection to assignments stream failed", zap.Error(err))
		}
		log.Info("Connected to assignments stream")
		batch, err := assigns.Recv()
		if err != nil {
			log.Fatal("Failed to recv assign", zap.Error(err))
		}
		log.Info("Got batch", zap.Int("batch_size", len(batch.Assignments)))
		for _, assign := range batch.Assignments {
			compactID, err := strconv.ParseInt(assign.Locator.Id, 10, 64)
			if err != nil {
				panic(err)
			}
			videoID := encodeVideoID(compactID)
			log.Info("Got assignment",
				zap.Int32("msg.partition", assign.KafkaPointer.Partition),
				zap.Int64("msg.offset", assign.KafkaPointer.Offset),
				zap.String("item.id", assign.Locator.Id),
				zap.String("video_id", videoID))
			req := api.GrabVideo(videoID)
			res := fasthttp.AcquireResponse()
			if err := fasthttp.Do(req, res); err != nil {
				log.Fatal("Failed to grab video", zap.Error(err))
			}
			var v data.Video
			if err := api.ParseVideo(&v, res); err != nil {
				log.Fatal("Failed to parse video", zap.Error(err))
			}
			var discovered []*types.ItemPointer
			for _, rel := range v.RelatedVideos {
				cid, err := decodeVideoID(rel.ID)
				if err != nil {
					log.Fatal("Discovered weird video ID", zap.Error(err), zap.String("video_id", rel.ID))
				}
				discovered = append(discovered, &types.ItemPointer{
					Dst: &types.ItemLocator{
						Collection: "yt.videos",
						Id:         strconv.FormatInt(cid, 10),
					},
					Timestamp: ptypes.TimestampNow(),
				})
			}
			if _, err := discovery.ReportDiscovered(ctx, &types.ReportDiscoveredRequest{
				Pointers: discovered,
			}); err != nil {
				log.Fatal("Failed to report discovered", zap.Error(err))
			}
			log.Info("Reported discovered", zap.Int("discovered_count", len(discovered)))
		}
	}
}

func mustDecodeVideoID(id string) int64 {
	num, err := decodeVideoID(id)
	if err != nil {
		panic(err)
	}
	return num
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
