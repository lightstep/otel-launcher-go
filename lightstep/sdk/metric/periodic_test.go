// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metric

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestPeriodicRepeats(t *testing.T) {
	var notime time.Time
	const cumulative = aggregation.CumulativeTemporality

	expectData := test.Metrics(resource.Empty(),
		test.Scope(
			test.Library("test"),
			test.Instrument(
				test.Descriptor("hello", sdkinstrument.SyncCounter, number.Int64Kind),
				test.Point(notime, notime, sum.NewMonotonicInt64(2), cumulative),
			),
		),
	)

	t.Run("export_once_and_stop", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		producer := NewMockProducer(ctrl)
		exporter := NewMockPushExporter(ctrl)
		periodic := NewPeriodicReader(exporter, time.Millisecond*25)

		require.Equal(t, time.Millisecond*25, periodic.timeout)
		require.Equal(t, time.Millisecond*25, periodic.interval)

		exporter.EXPECT().String().Return("mock").AnyTimes()

		require.Equal(t, "mock interval 25ms", periodic.String())

		producer.EXPECT().Produce(gomock.Not(gomock.Nil())).DoAndReturn(func(ptr *data.Metrics) data.Metrics {
			return expectData
		}).AnyTimes()

		exporter.EXPECT().ExportMetrics(gomock.Any(), gomock.Eq(expectData)).DoAndReturn(func(_ context.Context, _ data.Metrics) error {
			periodic.stop()
			return nil
		})

		periodic.Register(producer)
		periodic.wait.Wait()
	})

	t.Run("export_timeout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		producer := NewMockProducer(ctrl)
		exporter := NewMockPushExporter(ctrl)
		periodic := NewPeriodicReader(exporter, time.Millisecond*25)

		producer.EXPECT().Produce(gomock.Not(gomock.Nil())).DoAndReturn(func(ptr *data.Metrics) data.Metrics {
			return expectData
		}).AnyTimes()

		count := 0
		exporter.EXPECT().ExportMetrics(gomock.Any(), gomock.Eq(expectData)).DoAndReturn(func(ctx context.Context, _ data.Metrics) error {
			count++
			if count == 1 {
				<-ctx.Done()
				err := ctx.Err()
				require.True(t, errors.Is(err, context.DeadlineExceeded))
				return err
			}
			periodic.stop()
			return nil
		}).MinTimes(2)

		periodic.Register(producer)
		periodic.wait.Wait()
	})

	t.Run("export_interval", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		interval := time.Millisecond * 25
		producer := NewMockProducer(ctrl)
		exporter := NewMockPushExporter(ctrl)
		periodic := NewPeriodicReader(exporter, interval)

		producer.EXPECT().Produce(gomock.Not(gomock.Nil())).DoAndReturn(func(ptr *data.Metrics) data.Metrics {
			return expectData
		}).AnyTimes()

		count := 0
		exporter.EXPECT().ExportMetrics(gomock.Any(), gomock.Eq(expectData)).DoAndReturn(func(ctx context.Context, _ data.Metrics) error {
			count++
			if count == 10 {
				periodic.stop()
			}
			return nil
		}).MinTimes(10)

		before := time.Now()
		periodic.Register(producer)
		periodic.wait.Wait()
		after := time.Now()
		require.Less(t, 10*interval, after.Sub(before))
	})

	t.Run("double_register", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		errs := test.OTelErrors()

		producer := NewMockProducer(ctrl)
		exporter := NewMockPushExporter(ctrl)
		periodic := NewPeriodicReader(exporter, time.Second)

		exporter.EXPECT().String().Return("mock").Times(1)

		periodic.Register(producer)
		periodic.Register(producer)

		periodic.stop()

		require.Equal(t, 1, len(*errs))
		require.True(t, errors.Is((*errs)[0], ErrMultipleReaderRegistration))
	})

	t.Run("shutdown", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		producer := NewMockProducer(ctrl)
		exporter := NewMockPushExporter(ctrl)
		periodic := NewPeriodicReader(exporter, time.Millisecond, WithTimeout(time.Hour))

		producer.EXPECT().Produce(gomock.Not(gomock.Nil())).DoAndReturn(func(ptr *data.Metrics) data.Metrics {
			return expectData
		}).AnyTimes()

		var exporting sync.WaitGroup
		exporting.Add(1)

		exporter.EXPECT().ExportMetrics(gomock.Any(), gomock.Eq(expectData)).DoAndReturn(func(ctx context.Context, _ data.Metrics) error {
			exporting.Done()
			<-ctx.Done()
			return nil
		}).Times(1)

		exporter.EXPECT().ShutdownMetrics(gomock.Any(), gomock.Eq(expectData)).Return(nil).Times(1)

		periodic.Register(producer)

		exporting.Wait()

		require.NoError(t, periodic.Shutdown(context.Background()))
	})

	t.Run("flush", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		producer := NewMockProducer(ctrl)
		exporter := NewMockPushExporter(ctrl)
		periodic := NewPeriodicReader(exporter, time.Millisecond, WithTimeout(time.Hour))

		producer.EXPECT().Produce(gomock.Not(gomock.Nil())).DoAndReturn(func(ptr *data.Metrics) data.Metrics {
			return expectData
		}).AnyTimes()

		exporter.EXPECT().ExportMetrics(gomock.Any(), gomock.Eq(expectData)).Return(nil).AnyTimes()

		exporter.EXPECT().ForceFlushMetrics(gomock.Any(), gomock.Eq(expectData)).Return(nil).Times(1)

		periodic.Register(producer)

		require.NoError(t, periodic.ForceFlush(context.Background()))
	})

	t.Run("options", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		exporter := NewMockPushExporter(ctrl)
		periodic := NewPeriodicReader(exporter, -time.Millisecond, WithTimeout(-time.Second))

		require.Equal(t, DefaultInterval, periodic.interval)
		require.Equal(t, DefaultTimeout, periodic.timeout)
	})
}
