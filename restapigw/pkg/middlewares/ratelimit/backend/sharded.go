// Package backend - MemoeryBackend에 대한 sharding 처리 패키지
package backend

import (
	"context"
	"time"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

// ShardedMemoryBackend - Mutex 경합을 피해 데이터를 Sharding하는 Memory Backend 구조
type ShardedMemoryBackend struct {
	shards []*MemoryBackend
	total  uint64
	hasher Hasher
}

// ===== [ Implementations ] =====

// shard - 지정한 키에 해당하는 Shard 정보 반환
func (smb *ShardedMemoryBackend) shard(key string) uint64 {
	return smb.hasher(key) % smb.total
}

// Load - 지정한 키에 해당하는 정보를 ShardedMemoryBackend에서 로드
func (smb *ShardedMemoryBackend) Load(key string, f func() interface{}) interface{} {
	return smb.shards[smb.shard(key)].Load(key, f)
}

// Store - 지정한 키에 해당하는 정보를 ShardedMemoryBackend로 저장
func (smb *ShardedMemoryBackend) Store(key string, v interface{}) error {
	return smb.shards[smb.shard(key)].Store(key, v)
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// DefaultShardedMemoryBackend - 기본 Shard 갯수 (2048)를 기반으로 분산하는 MemoryBackend 생성
func DefaultShardedMemoryBackend(ctx context.Context) *ShardedMemoryBackend {
	return NewShardedMemoryBackend(ctx, shards, DataTTL, PseudoFNV64a)
}

// NewShardedMemoryBackend - 지정된 shard 수에 맞는 Memory기반 ShardedBackend 생성
func NewShardedMemoryBackend(ctx context.Context, shards uint64, ttl time.Duration, h Hasher) *ShardedMemoryBackend {
	b := &ShardedMemoryBackend{
		shards: make([]*MemoryBackend, shards),
		total:  shards,
		hasher: h,
	}

	var i uint64
	for i = 0; i < shards; i++ {
		b.shards[i] = NewMemoryBackend(ctx, ttl)
	}
	return b
}
