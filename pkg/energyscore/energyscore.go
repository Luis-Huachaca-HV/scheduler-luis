package energyscore

import (
	"context"
	"fmt"
	"strconv"
	//energyscore "sigs.k8s.io/scheduler-plugins/pkg/energyscore"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fwk "k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/scheduler-plugins/apis/config" // Tu config interna
)

const Name = "EnergyScore"

// EnergyScore implementa ScorePlugin
type EnergyScore struct {
	handle           fwk.Handle
	weightMultiplier float64 // argumento de configuración
}

var _ fwk.ScorePlugin = &EnergyScore{}

func (es *EnergyScore) Name() string {
	return Name
}

func (es *EnergyScore) Score(ctx context.Context, state *fwk.CycleState, pod *corev1.Pod, nodeName string) (int64, *fwk.Status) {
	nodeInfo, err := es.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, fwk.NewStatus(fwk.Error, err.Error())
	}

	score := int64(0)
	if val, ok := nodeInfo.Node().Labels["energy-score"]; ok {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			score = int64(parsed * es.weightMultiplier)
		}
	}
	return score, fwk.NewStatus(fwk.Success)
}

func (es *EnergyScore) ScoreExtensions() fwk.ScoreExtensions {
	return nil
}

// New crea una instancia del plugin usando los argumentos de configuración
func New(_ context.Context, obj runtime.Object, handle fwk.Handle) (fwk.Plugin, error) {
	args, ok := obj.(*config.EnergyScoreArgs)
	if !ok {
		return nil, fmt.Errorf("expected EnergyScoreArgs, got %T", obj)
	}
	return &EnergyScore{
		handle:           handle,
		weightMultiplier: args.WeightMultiplier,
	}, nil
}
