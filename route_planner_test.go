package geoapify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func intPtr(v int) *int { return &v }

func TestRoutePlanner_Plan(t *testing.T) {
	tests := []struct {
		name  string
		build func(*RoutePlannerService) *RoutePlannerRequest
		check func(*testing.T, routePlannerBody)
	}{
		{
			name: "basic jobs",
			build: func(s *RoutePlannerService) *RoutePlannerRequest {
				return s.Plan().
					WithMode(ModeDrive).
					WithAgents(PlannerAgent{
						ID:            "agent-1",
						StartLocation: [2]float64{2.35, 48.85},
						EndLocation:   [2]float64{2.35, 48.85},
					}).
					WithJobs(PlannerJob{
						ID:       "job-1",
						Location: [2]float64{2.36, 48.86},
						Duration: intPtr(300),
					})
			},
			check: func(t *testing.T, b routePlannerBody) {
				assertEqual(t, b.Mode, ModeDrive)
				assertEqual(t, len(b.Agents), 1)
				assertEqual(t, b.Agents[0].ID, "agent-1")
				assertEqual(t, len(b.Jobs), 1)
				assertEqual(t, b.Jobs[0].ID, "job-1")
				assertEqual(t, *b.Jobs[0].Duration, 300)
			},
		},
		{
			name: "with shipments",
			build: func(s *RoutePlannerService) *RoutePlannerRequest {
				return s.Plan().
					WithMode(ModeDrive).
					WithAgents(PlannerAgent{ID: "a1", StartLocation: [2]float64{0, 0}}).
					WithShipments(PlannerShipment{
						ID: "s1",
						Pickup: PlannerShipmentStop{
							Location: [2]float64{1, 2},
							Duration: intPtr(60),
						},
						Delivery: PlannerShipmentStop{
							Location: [2]float64{3, 4},
							Duration: intPtr(60),
						},
						Amount: intPtr(5),
					})
			},
			check: func(t *testing.T, b routePlannerBody) {
				assertEqual(t, len(b.Shipments), 1)
				assertEqual(t, b.Shipments[0].ID, "s1")
				assertEqual(t, *b.Shipments[0].Amount, 5)
				assertEqual(t, b.Shipments[0].Pickup.Location, [2]float64{1, 2})
				assertEqual(t, *b.Shipments[0].Delivery.Duration, 60)
			},
		},
		{
			name: "with locations",
			build: func(s *RoutePlannerService) *RoutePlannerRequest {
				return s.Plan().
					WithMode(ModeDrive).
					WithLocations(
						PlannerLocation{ID: "loc1", Location: [2]float64{2.35, 48.85}},
					).
					WithAgents(PlannerAgent{
						ID:               "a1",
						StartLocationIdx: intPtr(0),
					}).
					WithJobs(PlannerJob{
						ID:          "j1",
						LocationIdx: intPtr(0),
					})
			},
			check: func(t *testing.T, b routePlannerBody) {
				assertEqual(t, len(b.Locations), 1)
				assertEqual(t, b.Locations[0].ID, "loc1")
				assertEqual(t, *b.Agents[0].StartLocationIdx, 0)
				assertEqual(t, *b.Jobs[0].LocationIdx, 0)
			},
		},
		{
			name: "all options",
			build: func(s *RoutePlannerService) *RoutePlannerRequest {
				return s.Plan().
					WithMode(ModeTruck).
					WithAgents(PlannerAgent{ID: "a1", StartLocation: [2]float64{0, 0}}).
					WithJobs(PlannerJob{ID: "j1", Location: [2]float64{1, 1}}).
					WithAvoid(RouteMatrixAvoid{Type: "tolls"}).
					WithTraffic(TrafficApproximated).
					WithType(RouteShort).
					WithMaxSpeed(90).
					WithUnits(UnitsImperial)
			},
			check: func(t *testing.T, b routePlannerBody) {
				assertEqual(t, b.Mode, ModeTruck)
				assertEqual(t, len(b.Avoid), 1)
				assertEqual(t, b.Traffic, TrafficApproximated)
				assertEqual(t, b.Type, RouteShort)
				assertEqual(t, b.MaxSpeed, 90)
				assertEqual(t, b.Units, UnitsImperial)
			},
		},
		{
			name: "defaults omitted",
			build: func(s *RoutePlannerService) *RoutePlannerRequest {
				return s.Plan().
					WithMode(ModeDrive).
					WithAgents(PlannerAgent{ID: "a1", StartLocation: [2]float64{0, 0}}).
					WithJobs(PlannerJob{ID: "j1", Location: [2]float64{1, 1}})
			},
			check: func(t *testing.T, b routePlannerBody) {
				assertEqual(t, len(b.Avoid), 0)
				assertEqual(t, b.Traffic, TrafficModel(""))
				assertEqual(t, b.Type, RouteType(""))
				assertEqual(t, b.MaxSpeed, 0)
				assertEqual(t, b.Units, Units(""))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.Method, http.MethodPost)
				assertEqual(t, r.URL.Path, "/v1/routeplanner")

				body, err := io.ReadAll(r.Body)
				assertNoError(t, err)
				var b routePlannerBody
				assertNoError(t, json.Unmarshal(body, &b))
				tt.check(t, b)

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"agents":[]}`))
			})

			req := tt.build(client.RoutePlanner())
			_, err := req.Do(context.Background())
			assertNoError(t, err)
		})
	}
}

func TestRoutePlanner_ResponseDeserialization(t *testing.T) {
	jobIdx := 0
	resp := RoutePlannerResponse{
		Properties: map[string]any{"mode": "drive"},
		Agents: []PlannerAgentResult{
			{
				AgentIndex: 0,
				Distance:   5000,
				Time:       600,
				Route: []PlannerRouteStep{
					{Type: "start", Distance: 0, Time: 0},
					{Type: "job", JobIndex: &jobIdx, Distance: 2500, Time: 300},
					{Type: "end", Distance: 5000, Time: 600},
				},
			},
		},
	}

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mustJSON(t, resp))
	})

	got, err := client.RoutePlanner().Plan().
		WithMode(ModeDrive).
		WithAgents(PlannerAgent{ID: "a1", StartLocation: [2]float64{0, 0}}).
		WithJobs(PlannerJob{ID: "j1", Location: [2]float64{1, 1}}).
		Do(context.Background())
	assertNoError(t, err)

	assertEqual(t, len(got.Agents), 1)
	assertEqual(t, got.Agents[0].AgentIndex, 0)
	assertEqual(t, got.Agents[0].Distance, 5000.0)
	assertEqual(t, got.Agents[0].Time, 600.0)
	assertEqual(t, len(got.Agents[0].Route), 3)
	assertEqual(t, got.Agents[0].Route[0].Type, "start")
	assertEqual(t, got.Agents[0].Route[1].Type, "job")
	if got.Agents[0].Route[1].JobIndex == nil {
		t.Fatal("expected job_index to be set")
	}
	assertEqual(t, *got.Agents[0].Route[1].JobIndex, 0)
}

func TestRoutePlanner_ErrorHandling(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"No agents provided"}`))
	})

	_, err := client.RoutePlanner().Plan().
		WithMode(ModeDrive).
		Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 400)
	assertEqual(t, apiErr.Message, "No agents provided")
}
