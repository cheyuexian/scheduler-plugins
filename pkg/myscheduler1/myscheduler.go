/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package myscheduler1

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"math"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type MyScheduler1 struct {
	handle framework.Handle
}

var _ = framework.ScorePlugin(&MyScheduler1{})

// Name is the name of the plugin used in the Registry and configurations.
const Name = "MyScheduler1"

func (ps *MyScheduler1) Name() string {
	return Name
}

// Score invoked at the score extension point.
func (ps *MyScheduler1) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := ps.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	klog.V(3).Info("Score %v,%v",pod.Name,nodeInfo.Node().Name)

	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

	// pe.score favors nodes with terminating pods instead of nominated pods
	// It calculates the sum of the node's terminating pods and nominated pods
	return ps.score(nodeInfo)
}

// ScoreExtensions of the Score plugin.
func (ps *MyScheduler1) ScoreExtensions() framework.ScoreExtensions {
	return ps
}

func (ps *MyScheduler1) score(nodeInfo *framework.NodeInfo) (int64, *framework.Status) {
	var terminatingPodNum, nominatedPodNum int64
	// get nominated Pods for node from nominatedPodMap
	for _, p := range nodeInfo.Pods {
		// Pod is terminating if DeletionTimestamp has been set
		if p.Pod.DeletionTimestamp != nil {
			terminatingPodNum++
		}
	}

	if nodeInfo.Node().Name == "hollow-node-5" {
		return 100000,nil
	}
	return 1,nil
	return terminatingPodNum - nominatedPodNum, nil
}

func (ps *MyScheduler1) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	klog.V(3).Info("NormalizeScore",pod.Name)
	// Find highest and lowest scores.
	var highest int64 = -math.MaxInt64
	var lowest int64 = math.MaxInt64
	for _, nodeScore := range scores {
		if nodeScore.Score > highest {
			highest = nodeScore.Score
		}
		if nodeScore.Score < lowest {
			lowest = nodeScore.Score
		}
	}

	// Transform the highest to lowest score range to fit the framework's min to max node score range.
	oldRange := highest - lowest
	newRange := framework.MaxNodeScore - framework.MinNodeScore
	for i, nodeScore := range scores {
		if oldRange == 0 {
			scores[i].Score = framework.MinNodeScore
		} else {
			scores[i].Score = ((nodeScore.Score - lowest) * newRange / oldRange) + framework.MinNodeScore
		}
	}

	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	return &MyScheduler1{handle: h}, nil
}
