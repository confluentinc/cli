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
				buildPair("sw-123456", "prod-kafka-dr", "west"),
				buildPair("sw-234567", "staging-kafka-dr", "west"),
			}}
			require.NoError(t, json.NewEncoder(w).Encode(pairs))
		case http.MethodPost:
			var req switchoverv1.SwitchoverV1SwitchoverPair
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			pair := buildPair("sw-123456", req.Spec.GetDisplayName(), req.Spec.GetActiveMember())
			if req.Spec.Members != nil {
				pair.Spec.Members = req.Spec.Members
			}
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
		case http.MethodGet:
			require.NoError(t, json.NewEncoder(w).Encode(buildPair(id, "prod-kafka-dr", "west")))
		case http.MethodPut:
			// Reflect the updated display name from the request body.
			var req switchoverv1.SwitchoverV1SwitchoverPairUpdateRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			name := "prod-kafka-dr"
			if req.Spec.DisplayName != nil {
				name = req.Spec.GetDisplayName()
			}
			require.NoError(t, json.NewEncoder(w).Encode(buildPair(id, name, "west")))
		case http.MethodDelete:
			// Delete is asynchronous: 202 Accepted with the resource body.
			w.WriteHeader(http.StatusAccepted)
			require.NoError(t, json.NewEncoder(w).Encode(buildPair(id, "prod-kafka-dr", "west")))
		}
	}
}

// handleSwitchoverPairFailover handles "/switchover/v1/switchover-pairs/{id}:failover".
func handleSwitchoverPairFailover(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		// Reflect the requested active member from the failover request body.
		var req switchoverv1.SwitchoverV1SwitchoverPairFailoverRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		active := "east"
		if req.Spec.ActiveMember != nil {
			active = req.Spec.GetActiveMember()
		}
		pair := buildPair(id, "prod-kafka-dr", active)
		pair.Status.SetPhase("UPDATING")
		w.WriteHeader(http.StatusAccepted)
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
			if req.Spec.Endpoints != nil {
				endpoint.Spec.Endpoints = req.Spec.Endpoints
			}
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
		case http.MethodGet:
			require.NoError(t, json.NewEncoder(w).Encode(buildEndpoint(id, "prod-kafka-dr-endpoint")))
		case http.MethodPut:
			// Reflect the updated display name from the request body.
			var req switchoverv1.SwitchoverV1SwitchoverEndpointUpdateRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			name := "prod-kafka-dr-endpoint"
			if req.Spec.DisplayName != nil {
				name = req.Spec.GetDisplayName()
			}
			require.NoError(t, json.NewEncoder(w).Encode(buildEndpoint(id, name)))
		case http.MethodDelete:
			// Delete is asynchronous: 202 Accepted with the resource body.
			w.WriteHeader(http.StatusAccepted)
			require.NoError(t, json.NewEncoder(w).Encode(buildEndpoint(id, "prod-kafka-dr-endpoint")))
		}
	}
}

func buildPair(id, name, activeMember string) switchoverv1.SwitchoverV1SwitchoverPair {
	if name == "" {
		name = "prod-kafka-dr"
	}
	if activeMember == "" {
		activeMember = "west"
	}
	return switchoverv1.SwitchoverV1SwitchoverPair{
		Id: switchoverv1.PtrString(id),
		Spec: &switchoverv1.SwitchoverV1SwitchoverPairSpec{
			DisplayName:    switchoverv1.PtrString(name),
			EnvironmentCrn: switchoverv1.PtrString("crn://confluent.cloud/organization=org-123/environment=env-123456"),
			Members: &[]switchoverv1.SwitchoverV1SwitchoverPairMember{
				{Name: "west", MemberCrn: "crn://confluent.cloud/organization=org-123/environment=env-123456/cloud-cluster=lkc-111111"},
				{Name: "east", MemberCrn: "crn://confluent.cloud/organization=org-123/environment=env-234567/cloud-cluster=lkc-222222"},
			},
			ActiveMember: switchoverv1.PtrString(activeMember),
			FirstActive:  switchoverv1.PtrString("west"),
		},
		Status: &switchoverv1.SwitchoverV1SwitchoverPairStatus{Phase: "READY_TO_FAILOVER"},
	}
}

func buildEndpoint(id, name string) switchoverv1.SwitchoverV1SwitchoverEndpoint {
	if name == "" {
		name = "prod-kafka-dr-endpoint"
	}
	return switchoverv1.SwitchoverV1SwitchoverEndpoint{
		Id: switchoverv1.PtrString(id),
		Spec: &switchoverv1.SwitchoverV1SwitchoverEndpointSpec{
			DisplayName:       switchoverv1.PtrString(name),
			EnvironmentCrn:    switchoverv1.PtrString("crn://confluent.cloud/organization=org-123/environment=env-123456"),
			ParentResourceCrn: switchoverv1.PtrString("crn://confluent.cloud/organization=org-123/environment=env-123456/switchover-pair=sw-123456"),
			Target:            switchoverv1.PtrString("west-platt"),
			Endpoints: &[]switchoverv1.SwitchoverV1EndpointConfig{
				{Name: "west-platt", EndpointFilter: switchoverv1.SwitchoverV1EndpointFilter{Type: "private", NetworkCrn: switchoverv1.PtrString("crn://confluent.cloud/organization=org-123/environment=env-123456/network=n-111111")}},
				{Name: "east-platt", EndpointFilter: switchoverv1.SwitchoverV1EndpointFilter{Type: "private", NetworkCrn: switchoverv1.PtrString("crn://confluent.cloud/organization=org-123/environment=env-234567/network=n-222222")}},
			},
		},
		Status: &switchoverv1.SwitchoverV1SwitchoverEndpointStatus{Phase: "READY"},
	}
}
