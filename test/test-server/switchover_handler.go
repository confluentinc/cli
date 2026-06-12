package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2/switchover/v1"
)

// handleSwitchoverPairs handles "/switchover/v1/switchover-pairs".
func handleSwitchoverPairs(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			pairs := switchoverv1.SwitchoverV1SwitchoverPairList{Data: []switchoverv1.SwitchoverV1SwitchoverPair{
				buildPair("sw-123456", "prod-kafka-dr"),
				buildPair("sw-234567", "staging-kafka-dr"),
			}}
			require.NoError(t, json.NewEncoder(w).Encode(pairs))
		case http.MethodPost:
			var req switchoverv1.SwitchoverV1SwitchoverPair
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			pair := buildPair("sw-123456", req.Spec.GetDisplayName())
			pair.Spec.Members = req.Spec.Members
			pair.Spec.SetActiveMember(req.Spec.GetActiveMember())
			require.NoError(t, json.NewEncoder(w).Encode(pair))
		}
	}
}

// handleSwitchoverPair handles "/switchover/v1/switchover-pairs/{id}".
func handleSwitchoverPair(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id == "sw-000000" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet, http.MethodPut:
			require.NoError(t, json.NewEncoder(w).Encode(buildPair(id, "prod-kafka-dr")))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// handleSwitchoverPairFailover handles "/switchover/v1/switchover-pairs/{id}:failover".
func handleSwitchoverPairFailover(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		pair := buildPair(id, "prod-kafka-dr")
		pair.Status.SetPhase("SWITCHING")
		require.NoError(t, json.NewEncoder(w).Encode(pair))
	}
}

// handleSwitchoverEndpoints handles "/switchover/v1/switchover-endpoints".
func handleSwitchoverEndpoints(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			endpoints := switchoverv1.SwitchoverV1SwitchoverEndpointList{Data: []switchoverv1.SwitchoverV1SwitchoverEndpoint{
				buildEndpoint("se-123456", "prod-kafka-dr-endpoint"),
			}}
			require.NoError(t, json.NewEncoder(w).Encode(endpoints))
		case http.MethodPost:
			var req switchoverv1.SwitchoverV1SwitchoverEndpoint
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			endpoint := buildEndpoint("se-123456", req.Spec.GetDisplayName())
			endpoint.Spec.Endpoints = req.Spec.Endpoints
			require.NoError(t, json.NewEncoder(w).Encode(endpoint))
		}
	}
}

// handleSwitchoverEndpoint handles "/switchover/v1/switchover-endpoints/{id}".
func handleSwitchoverEndpoint(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id == "se-000000" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet, http.MethodPut:
			require.NoError(t, json.NewEncoder(w).Encode(buildEndpoint(id, "prod-kafka-dr-endpoint")))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// handleSwitchoverEndpointActivate handles "/switchover/v1/switchover-endpoints/{id}:activate".
func handleSwitchoverEndpointActivate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		require.NoError(t, json.NewEncoder(w).Encode(buildEndpoint(id, "prod-kafka-dr-endpoint")))
	}
}

func buildPair(id, name string) switchoverv1.SwitchoverV1SwitchoverPair {
	return switchoverv1.SwitchoverV1SwitchoverPair{
		Id: switchoverv1.PtrString(id),
		Spec: &switchoverv1.SwitchoverV1SwitchoverPairSpec{
			DisplayName: switchoverv1.PtrString(name),
			Environment: &switchoverv1.EnvScopedObjectReference{Id: "env-123456"},
			Members: &[]switchoverv1.SwitchoverV1SwitchoverPairMember{
				{Name: "west", MemberId: "lkc-111111"},
				{Name: "east", MemberId: "lkc-222222"},
			},
			ActiveMember: switchoverv1.PtrString("west"),
		},
		Status: &switchoverv1.SwitchoverV1SwitchoverPairStatus{Phase: "READY"},
	}
}

func buildEndpoint(id, name string) switchoverv1.SwitchoverV1SwitchoverEndpoint {
	return switchoverv1.SwitchoverV1SwitchoverEndpoint{
		Id: switchoverv1.PtrString(id),
		Spec: &switchoverv1.SwitchoverV1SwitchoverEndpointSpec{
			DisplayName:    switchoverv1.PtrString(name),
			Environment:    &switchoverv1.EnvScopedObjectReference{Id: "env-123456"},
			SwitchoverPair: &switchoverv1.EnvScopedObjectReference{Id: "sw-123456"},
			Target:         switchoverv1.PtrString("west-platt"),
			DrEndpoint:     switchoverv1.PtrString("glkc-sw-123456-se-123456.global.confluent.cloud"),
			Endpoints: &[]switchoverv1.SwitchoverV1EndpointConfig{
				{Name: "west-platt", EndpointFilter: switchoverv1.SwitchoverV1EndpointFilter{ResourceId: "lkc-111111", Type: "PRIVATE"}},
				{Name: "east-platt", EndpointFilter: switchoverv1.SwitchoverV1EndpointFilter{ResourceId: "lkc-222222", Type: "PRIVATE"}},
			},
		},
		Status: &switchoverv1.SwitchoverV1SwitchoverEndpointStatus{Phase: "READY"},
	}
}
