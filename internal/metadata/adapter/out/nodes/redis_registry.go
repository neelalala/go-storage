package nodes

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/neelalala/go-storage/internal/metadata/domain"
)

const (
	SweepInterval                  = 1 * time.Second
	NoHeartbeatCountToMarkNodeDead = 3
)

const (
	defaultNodeKeyPrefix = "storage:nodes:"
	defaultNodeSetKey    = "storage:nodes:all"
)

type RedisNodeRegistry struct {
	client            *redis.Client

	heartbeatInterval time.Duration
	nodeTTL           time.Duration
	sweepInterval     time.Duration

	nodeKeyPrefix     string
	nodeSetKey        string

	log               *slog.Logger
}

func NewRedisNodeRegistry(
	ctx context.Context,
	rawURL string,
	heartbeatInterval time.Duration,
	ttlMultiplier int,
	log *slog.Logger,
) (*RedisNodeRegistry, error) {
	opts, err := parseRedisURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url %q: %w", rawURL, err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	if ttlMultiplier <= 0 {
		ttlMultiplier = NoHeartbeatCountToMarkNodeDead
	}

	nodeTTL := time.Duration(ttlMultiplier) * heartbeatInterval
	return &RedisNodeRegistry{
		client:            client,
		heartbeatInterval: heartbeatInterval,
		nodeTTL:           nodeTTL,
		sweepInterval:     SweepInterval,
		nodeKeyPrefix:     defaultNodeKeyPrefix,
		nodeSetKey:        defaultNodeSetKey,
		log:               log,
	}, nil
}

func (r *RedisNodeRegistry) Close() error {
	return r.client.Close()
}

func (r *RedisNodeRegistry) ProcessHeartbeat(ctx context.Context, node domain.StorageNode) {
	key := r.nodeKeyPrefix + node.ID.String()

	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, node.Address, r.nodeTTL)
	saddCmd := pipe.SAdd(ctx, r.nodeSetKey, node.ID.String())

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to process heartbeat in redis",
			slog.String("error", err.Error()),
			slog.String("node_id", node.ID.String()),
		)
		return
	}

	if saddCmd.Val() > 0 {
		r.log.InfoContext(
			ctx, "new node",
			slog.Group(
				"node",
				slog.String("id", node.ID.String()),
				slog.String("address", node.Address),
			),
		)
	}
}

func (r *RedisNodeRegistry) GetAllNodes(ctx context.Context) ([]domain.StorageNode, error) {
	members, err := r.client.SMembers(ctx, r.nodeSetKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node members from redis: %w", err)
	}

	if len(members) == 0 {
		return []domain.StorageNode{}, nil
	}

	keys := make([]string, len(members))
	for i, m := range members {
		keys[i] = r.nodeKeyPrefix + m
	}

	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to mget nodes from redis: %w", err)
	}

	nodes := make([]domain.StorageNode, 0, len(members))
	var deadIDs []any

	for i, val := range vals {
		idStr := members[i]
		if val == nil {
			deadIDs = append(deadIDs, idStr)
			continue
		}

		addr, ok := val.(string)
		if !ok || addr == "" {
			deadIDs = append(deadIDs, idStr)
			continue
		}

		nodeID, err := uuid.Parse(idStr)
		if err != nil {
			deadIDs = append(deadIDs, idStr)
			continue
		}

		nodes = append(nodes, domain.StorageNode{
			ID:      nodeID,
			Address: addr,
		})
	}

	if len(deadIDs) > 0 {
		if err := r.client.SRem(ctx, r.nodeSetKey, deadIDs...).Err(); err != nil {
			r.log.WarnContext(ctx, "failed to remove dead nodes from set", slog.String("error", err.Error()))
		}
		for _, deadID := range deadIDs {
			r.log.InfoContext(
				ctx, "node died",
				slog.Group(
					"node",
					slog.String("id", fmt.Sprint(deadID)),
				),
			)
		}
	}

	return nodes, nil
}

func (r *RedisNodeRegistry) GetNode(ctx context.Context, id uuid.UUID) (domain.StorageNode, error) {
	key := r.nodeKeyPrefix + id.String()
	addr, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			_ = r.client.SRem(ctx, r.nodeSetKey, id.String()).Err()
			return domain.StorageNode{}, fmt.Errorf("%w: id %s", domain.ErrStorageNodeNotFound, id)
		}
		return domain.StorageNode{}, fmt.Errorf("failed to get node %s from redis: %w", id, err)
	}

	return domain.StorageNode{
		ID:      id,
		Address: addr,
	}, nil
}

func parseRedisURL(rawURL string) (*redis.Options, error) {
	if !strings.HasPrefix(rawURL, "redis://") && !strings.HasPrefix(rawURL, "rediss://") {
		rawURL = "redis://" + rawURL
	}
	return redis.ParseURL(rawURL)
}
