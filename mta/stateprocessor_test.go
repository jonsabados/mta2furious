package mta

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateProcessor_ProcessUpdates(t *testing.T) {
	type testIteration struct {
		time         time.Time
		systemStatus []TripUpdate
	}

	strPtr := func(s string) *string {
		return &s
	}

	timeOrDie := func(str string) *time.Time {
		ret, err := time.Parse(time.RFC3339, str)
		require.NoError(t, err)
		return &ret
	}

	testCases := []struct {
		name             string
		iterations       []testIteration
		expectedSegments []Segment
	}{
		{
			name: "happy path, single trip",
			iterations: []testIteration{
				{
					time: *timeOrDie("2023-07-20T14:04:08-04:00"),
					systemStatus: []TripUpdate{
						{
							TripId:     "084421_G..N",
							RouteId:    "G",
							TrainId:    "1G 1404 CHU/CRS",
							IsAssigned: true,
							StopTimeUpdate: []StopTimeUpdate{
								{
									StopID:         "F27N",
									Arrival:        timeOrDie("2023-07-20T14:04:11-04:00"),
									Departure:      timeOrDie("2023-07-20T14:04:11-04:00"),
									ScheduledTrack: strPtr("B2"),
									ActualTrack:    strPtr("B2"),
								},
								{
									StopID:         "F26N",
									Arrival:        timeOrDie("2023-07-20T14:06:11-04:00"),
									Departure:      timeOrDie("2023-07-20T14:06:11-04:00"),
									ScheduledTrack: strPtr("B2"),
									ActualTrack:    strPtr("B2"),
								},
								{
									StopID:         "F25N",
									Arrival:        timeOrDie("2023-07-20T14:08:41-04:00"),
									Departure:      timeOrDie("2023-07-20T14:08:41-04:00"),
									ScheduledTrack: strPtr("B2"),
									ActualTrack:    strPtr("B2"),
								},
							},
						},
					},
				},
				{
					time: *timeOrDie("2023-07-20T14:05:08-04:00"),
					systemStatus: []TripUpdate{
						{
							TripId:     "084421_G..N",
							RouteId:    "G",
							TrainId:    "1G 1404 CHU/CRS",
							IsAssigned: true,
							StopTimeUpdate: []StopTimeUpdate{
								{
									StopID:         "F26N",
									Arrival:        timeOrDie("2023-07-20T14:06:15-04:00"),
									Departure:      timeOrDie("2023-07-20T14:06:15-04:00"),
									ScheduledTrack: strPtr("B2"),
									ActualTrack:    strPtr("B2"),
								},
								{
									StopID:         "F25N",
									Arrival:        timeOrDie("2023-07-20T14:08:41-04:00"),
									Departure:      timeOrDie("2023-07-20T14:08:41-04:00"),
									ScheduledTrack: strPtr("B2"),
									ActualTrack:    strPtr("B2"),
								},
							},
						},
					},
				},
				{
					time: *timeOrDie("2023-07-20T14:06:16-04:00"),
					systemStatus: []TripUpdate{
						{
							TripId:     "084421_G..N",
							RouteId:    "G",
							TrainId:    "1G 1404 CHU/CRS",
							IsAssigned: true,
							StopTimeUpdate: []StopTimeUpdate{
								{
									StopID:         "F25N",
									Arrival:        timeOrDie("2023-07-20T14:08:41-04:00"),
									Departure:      timeOrDie("2023-07-20T14:08:41-04:00"),
									ScheduledTrack: strPtr("B2"),
									ActualTrack:    strPtr("B2"),
								},
							},
						},
					},
				},
				{
					time: *timeOrDie("2023-07-20T14:10:16-04:00"),
					systemStatus: []TripUpdate{
						{
							TripId:         "084421_G..N",
							RouteId:        "G",
							TrainId:        "1G 1404 CHU/CRS",
							IsAssigned:     true,
							StopTimeUpdate: []StopTimeUpdate{},
						},
					},
				},
			},
			expectedSegments: []Segment{
				{
					FromStation:    "F27N",
					ToStation:      "F26N",
					DepartAt:       *timeOrDie("2023-07-20T14:04:11-04:00"),
					ArriveAt:       *timeOrDie("2023-07-20T14:06:15-04:00"),
					TripID:         "084421_G..N",
					RouteID:        "G",
					TrainID:        "1G 1404 CHU/CRS",
					IsAssigned:     true,
					ScheduledTrack: strPtr("B2"),
					ActualTrack:    strPtr("B2"),
				},
				{
					FromStation:    "F26N",
					ToStation:      "F25N",
					DepartAt:       *timeOrDie("2023-07-20T14:06:15-04:00"),
					ArriveAt:       *timeOrDie("2023-07-20T14:08:41-04:00"),
					TripID:         "084421_G..N",
					RouteID:        "G",
					TrainID:        "1G 1404 CHU/CRS",
					IsAssigned:     true,
					ScheduledTrack: strPtr("B2"),
					ActualTrack:    strPtr("B2"),
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			var testTime time.Time

			oracle := NewMockStateOracle(t)
			store := NewMemoryStore()
			testInstance := NewStateProcessor(oracle, store)
			testInstance.now = func() time.Time {
				return testTime
			}

			segmentsGot := make([]Segment, 0)

			for _, it := range tc.iterations {
				testTime = it.time
				oracle.EXPECT().CurrentState(ctx).Return(it.systemStatus, nil).Times(1)
				completed, err := testInstance.ProcessUpdates(ctx)
				require.NoError(t, err)
				segmentsGot = append(segmentsGot, completed.CompletedSegments...)
			}

			assert.Equal(t, tc.expectedSegments, segmentsGot)
		})
	}
}
