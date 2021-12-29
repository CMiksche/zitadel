package projection

import (
	"testing"

	"github.com/caos/zitadel/internal/domain"
	"github.com/caos/zitadel/internal/errors"
	"github.com/caos/zitadel/internal/eventstore"
	"github.com/caos/zitadel/internal/eventstore/handler"
	"github.com/caos/zitadel/internal/eventstore/repository"
	"github.com/caos/zitadel/internal/repository/org"
)

func TestFlowProjection_reduces(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "reduceTriggerActionsSetEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.TriggerActionsSetEventType),
					org.AggregateType,
					[]byte(`{"flowType": 1, "triggerType": 1, "actionIDs": ["id1", "id2"]}`),
				), org.TriggerActionsSetEventMapper),
			},
			reduce: (&FlowProjection{}).reduceTriggerActionsSetEventType,
			want: wantReduce{
				projection:       FlowTriggerTable,
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM zitadel.projections.flows_triggers WHERE (flow_type = $1) AND (trigger_type = $2)",
							expectedArgs: []interface{}{
								domain.FlowTypeExternalAuthentication,
								domain.TriggerTypePostAuthentication,
							},
						},
						{
							expectedStmt: "INSERT INTO zitadel.projections.flows_triggers (resource_owner, flow_type, trigger_type, action_id, trigger_sequence) VALUES ($1, $2, $3, $4, $5)",
							expectedArgs: []interface{}{
								"ro-id",
								domain.FlowTypeExternalAuthentication,
								domain.TriggerTypePostAuthentication,
								"id1",
								0,
							},
						},
						{
							expectedStmt: "INSERT INTO zitadel.projections.flows_triggers (resource_owner, flow_type, trigger_type, action_id, trigger_sequence) VALUES ($1, $2, $3, $4, $5)",
							expectedArgs: []interface{}{
								"ro-id",
								domain.FlowTypeExternalAuthentication,
								domain.TriggerTypePostAuthentication,
								"id2",
								1,
							},
						},
					},
				},
			},
		},
		{
			name: "reduceFlowClearedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.FlowClearedEventType),
					org.AggregateType,
					[]byte(`{"flowType": 1}`),
				), org.FlowClearedEventMapper),
			},
			reduce: (&FlowProjection{}).reduceFlowClearedEventType,
			want: wantReduce{
				projection:       FlowTriggerTable,
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM zitadel.projections.flows_triggers WHERE (flow_type = $1)",
							expectedArgs: []interface{}{
								domain.FlowTypeExternalAuthentication,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, tt.want)
		})
	}
}
