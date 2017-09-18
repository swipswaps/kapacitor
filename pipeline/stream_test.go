package pipeline

import (
	"bytes"
	"testing"
	"time"

	"github.com/influxdata/kapacitor/tick/ast"
)

func TestStreamNode_Tick(t *testing.T) {
	type args struct {
		wheres             []*ast.LambdaNode
		groupBy            []interface{}
		groupByMeasurement bool
		db                 string
		rp                 string
		measurement        string
		truncate           time.Duration
		round              time.Duration
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "stream with multiple where",
			args: args{
				wheres: []*ast.LambdaNode{
					{
						Expression: &ast.BinaryNode{
							Left: &ast.ReferenceNode{
								Reference: "cpu",
							},
							Right: &ast.StringNode{
								Literal: "cpu-total",
							},
							Operator: ast.TokenNotEqual,
						},
					},
					{
						Expression: &ast.BinaryNode{
							Left: &ast.ReferenceNode{
								Reference: "host",
							},
							Right: &ast.RegexNode{
								Literal: `logger\d+`,
							},
							Operator: ast.TokenRegexEqual,
						},
					},
				},
				groupBy: []interface{}{
					&ast.StarNode{},
				},
				groupByMeasurement: true,
				db:                 "telegraf",
				rp:                 "autogen",
				measurement:        "cpu",
			},
			want: `stream|from().database('telegraf').retentionPolicy('autogen').measurement('cpu').groupByMeasurement().groupBy(*).where(lambda: "cpu" != 'cpu-total' AND "host" =~ /logger\d+/)`,
		},
		{
			name: "stream with multiple group bys",
			args: args{
				wheres: []*ast.LambdaNode{
					{
						Expression: &ast.BinaryNode{
							Left: &ast.ReferenceNode{
								Reference: "cpu",
							},
							Right: &ast.StringNode{
								Literal: "cpu-total",
							},
							Operator: ast.TokenNotEqual,
						},
					},
				},
				groupBy: []interface{}{
					"host",
					"cpu",
				},
				groupByMeasurement: true,
				db:                 "telegraf",
				rp:                 "autogen",
				measurement:        "cpu",
			},
			want: `stream|from().database('telegraf').retentionPolicy('autogen').measurement('cpu').groupByMeasurement().groupBy('host', 'cpu').where(lambda: "cpu" != 'cpu-total')`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStreamNode()
			s.setPipeline(&Pipeline{})
			from := s.From()

			for i := range tt.args.wheres {
				from.Where(tt.args.wheres[i])
			}

			from.Dimensions = tt.args.groupBy
			from.GroupByMeasurementFlag = tt.args.groupByMeasurement
			from.Database = tt.args.db
			from.RetentionPolicy = tt.args.rp
			from.Measurement = tt.args.measurement
			from.Truncate = tt.args.truncate
			from.Round = tt.args.round

			var buf bytes.Buffer
			s.Tick(&buf)
			got := buf.String()
			if got != tt.want {
				t.Errorf("%q. TestStreamNode_Tick() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
