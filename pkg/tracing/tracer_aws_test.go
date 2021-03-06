package tracing_test

import (
	"github.com/applike/gosoline/pkg/cfg"
	"github.com/applike/gosoline/pkg/stream"
	"github.com/applike/gosoline/pkg/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/twinj/uuid"
	"testing"
)

func TestAwsTracer_StartSubSpan(t *testing.T) {
	tracer := getTracer()

	ctx, trans := tracer.StartSpan("test_trans")
	ctx, span := tracer.StartSubSpan(ctx, "test_span")

	assert.Equal(t, trans.GetTrace().TraceId, span.GetTrace().TraceId, "the trace ids should match")
	assert.Equal(t, trans.GetTrace().Sampled, span.GetTrace().Sampled, "the sample decision should match")
	assert.NotEqual(t, trans.GetTrace().Id, span.GetTrace().Id, "the span ids should be different")
	assert.Empty(t, trans.GetTrace().GetParentId(), "the parent id of the transaction should be empty")
	assert.Empty(t, span.GetTrace().GetParentId(), "the parent id of the span should be empty")
}

func TestAwsTracer_StartSpanFromContext(t *testing.T) {
	tracer := getTracer()

	ctx, transRoot := tracer.StartSpan("test_trans")
	ctx, transChild := tracer.StartSpanFromContext(ctx, "another_trace")

	assert.Equal(t, transRoot.GetTrace().TraceId, transChild.GetTrace().TraceId, "the trace ids should match")
	assert.Equal(t, transRoot.GetTrace().Sampled, transChild.GetTrace().Sampled, "the sample decision should match")
	assert.NotEqual(t, transRoot.GetTrace().Id, transChild.GetTrace().Id, "the span ids should be different")
	assert.Empty(t, transRoot.GetTrace().GetParentId(), "the parent id of the root transaction should be empty")
	assert.NotEmpty(t, transChild.GetTrace().GetParentId(), "the parent id of the child transaction should not be empty")
	assert.Equal(t, transRoot.GetTrace().Id, transChild.GetTrace().ParentId, "span id of root should match parent id of child")
}

func TestAwsTracer_StartSpanFromTraceAble(t *testing.T) {
	tracer := getTracer()

	traceAble := &stream.Message{
		Trace: &tracing.Trace{
			Id:       uuid.NewV4().String(),
			ParentId: uuid.NewV4().String(),
			Sampled:  true,
		},
	}
	_, trans := tracer.StartSpanFromTraceAble(traceAble, "another_trace")

	assert.Equal(t, traceAble.GetTrace().TraceId, trans.GetTrace().TraceId, "the trace ids should match")
	assert.Equal(t, traceAble.GetTrace().Sampled, trans.GetTrace().Sampled, "the sample decision should match")
	assert.NotEqual(t, traceAble.GetTrace().Id, trans.GetTrace().Id, "the span ids should be different")
	assert.NotEmpty(t, trans.GetTrace().GetParentId(), "the parent id of the transaction should not be empty")
	assert.Equal(t, traceAble.GetTrace().Id, trans.GetTrace().ParentId, "span id of traceAble should match parent id of trans")
}

func getTracer() tracing.Tracer {
	return tracing.NewAwsTracerWithInterfaces(cfg.AppId{
		Project:     "test_project",
		Environment: "test_env",
		Family:      "test_family",
		Application: "test_name",
	}, tracing.Settings{Enabled: true})
}
