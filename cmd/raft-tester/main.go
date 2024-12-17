package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
	"github.com/lni/dragonboat/v4"
	"golang.org/x/exp/rand"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/internal/raft"
)

var cli struct {
	RaftConfig raft.RaftConfig `embed:"" prefix:"raft-"`
}

type IntStateMachine struct {
	sum int64
}

type IntEvent int64

func (i *IntEvent) UnmarshalBinary(data []byte) error { //nolint:unparam
	*i = IntEvent(binary.BigEndian.Uint64(data))
	return nil
}

func (i IntEvent) MarshalBinary() ([]byte, error) { //nolint:unparam
	return binary.BigEndian.AppendUint64([]byte{}, uint64(i)), nil
}

var _ raft.StateMachine[int64, int64, IntEvent, *IntEvent] = &IntStateMachine{}

func (s IntStateMachine) Lookup(key int64) (int64, error) {
	return s.sum, nil
}

func (s *IntStateMachine) Update(msg IntEvent) error {
	s.sum += int64(msg)
	return nil
}

func (s IntStateMachine) Close() error {
	return nil
}

func (s IntStateMachine) Recover(reader io.Reader) error {
	err := binary.Read(reader, binary.BigEndian, &s.sum)
	if err != nil {
		return fmt.Errorf("failed to recover from snapshot: %w", err)
	}
	return nil
}

func (s IntStateMachine) Save(writer io.Writer) error {
	err := binary.Write(writer, binary.BigEndian, s.sum)
	if err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}
	return nil
}

func main() {
	kctx := kong.Parse(&cli)
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cluster := raft.New(&cli.RaftConfig)
	shard := raft.AddShard(ctx, cluster, 1, &IntStateMachine{})

	wg, ctx := errgroup.WithContext(ctx)
	messages := make(chan int)

	wg.Go(func() error {
		defer close(messages)
		// send a random number every 10 seconds
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				messages <- rand.Intn(1000)
			case <-ctx.Done():
				return nil
			}
		}
	})
	wg.Go(func() error {
		return cluster.Start(ctx, nil)
	})
	wg.Go(func() error {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case msg := <-messages:
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				err := shard.Propose(ctx, IntEvent(msg))
				if errors.Is(err, dragonboat.ErrShardNotReady) {
					log.Println("shard not ready")
				} else if err != nil {
					return fmt.Errorf("failed to propose event: %w", err)
				}
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				state, err := shard.Query(ctx, 1)
				if err != nil {
					return fmt.Errorf("failed to query shard: %w", err)
				}
				log.Println("state: ", state)
			case <-ctx.Done():
				return nil
			}
		}
	})

	if err := wg.Wait(); err != nil {
		kctx.FatalIfErrorf(err)
	}
}
